package blobstore

import (
	"blob_store_service/pkg/utils"
	"context"
	"database/sql"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type BlobStore struct {
	pool *pgxpool.Pool
}

func NewBlobStore(dbConnString string) (*BlobStore, error) {
	pool, err := pgxpool.New(context.Background(), dbConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return &BlobStore{pool: pool}, nil
}

func (bs *BlobStore) genConnect(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := bs.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %v", err)
	}
	return conn, nil
}

func (bs *BlobStore) releaseTransaction(tx pgx.Tx, err error, ctx context.Context) {
	if err != nil {
		tx.Rollback(ctx)
	} else {
		tx.Commit(ctx)
	}
}

func (bs *BlobStore) SaveBlob(filename string, contentType string, data io.Reader) (*Blob, error) {
	if !utils.IsValidFileType(contentType, []string{"image/", "video/"}) {
		return nil, fmt.Errorf("Invalid file type")
	}

	ctx := context.Background()
	conn, err := bs.genConnect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer bs.releaseTransaction(tx, err, ctx)

	lo_util := tx.LargeObjects()
	oid, err := lo_util.Create(ctx, 0)

	if err != nil {
		return nil, fmt.Errorf("failed to create large object %v", err)
	}
	lo, err := lo_util.Open(ctx, oid, pgx.LargeObjectModeWrite)

	if err != nil {
		return nil, fmt.Errorf("failed to open large object oid: %d, err %v", oid, err)
	}
	defer lo.Close()
	size, err := utils.CopyBytes(lo, data)

	if err != nil {
		return nil, err
	}

	id := uuid.New().String()
	now := time.Now()
	blob := &Blob{
		ID:           id,
		FileName:     filename,
		Size:         size,
		ContentType:  contentType,
		CreatedAt:    now,
		LastModified: now,
		DataOID:      oid,
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO blobs (id, filename, size, content_type, created_at, last_modified, data_oid)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, blob.ID, blob.FileName, blob.Size, blob.ContentType, blob.CreatedAt, blob.LastModified, blob.DataOID)
	if err != nil {
		return nil, fmt.Errorf("failed to save metadata: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return blob, nil
}

func (bs *BlobStore) GetBlob(id string) (*Blob, error) {
	ctx := context.Background()
	conn, err := bs.genConnect(ctx)

	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var blob Blob
	err = conn.QueryRow(ctx, `
        SELECT id, filename, size, content_type, created_at, last_modified, data_oid 
        FROM blobs WHERE id = $1
    `, id).Scan(
		&blob.ID, &blob.FileName, &blob.Size, &blob.ContentType,
		&blob.CreatedAt, &blob.LastModified, &blob.DataOID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to retrieve blob metadata: %v", err)
	}

	return &blob, nil
}

func (bs *BlobStore) GetListBlobWithPagination(limit, offset int) ([]*Blob, error) {
	ctx := context.Background()
	conn, err := bs.genConnect(ctx)

	if err != nil {
		return nil, err
	}
	rows, err := conn.Query(ctx, `
        SELECT id, filename, size, content_type, created_at, last_modified, data_oid 
        FROM blobs ORDER BY created_at DESC LIMIT $1 OFFSET $2
    `, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blob list: %v", err)
	}
	defer rows.Close()

	blobs := []*Blob{}
	for rows.Next() {
		var blob Blob
		err := rows.Scan(
			&blob.ID, &blob.FileName, &blob.Size, &blob.ContentType,
			&blob.CreatedAt, &blob.LastModified, &blob.DataOID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blob row: %v", err)
		}
		blobs = append(blobs, &blob)
	}

	return blobs, nil
}

func (bs *BlobStore) DeleteBlob(id string) error {
	ctx := context.Background()
	conn, err := bs.genConnect(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	blob, err := bs.GetBlob(id)
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer bs.releaseTransaction(tx, err, ctx)

	lo_util := tx.LargeObjects()
	err = lo_util.Unlink(ctx, blob.DataOID)

	if err != nil {
		return fmt.Errorf("failed to delete large object: %v", err)
	}
	_, err = tx.Exec(ctx, "DELETE FROM blobs WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete blob metadata: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (bs *BlobStore) StreamBlob(id string, w io.Writer) error {
	ctx := context.Background()
	conn, err := bs.genConnect(ctx)
	if err != nil {
		return err
	}
	blob, err := bs.GetBlob(id)
	if err != nil {
		return fmt.Errorf("failed to get blob: %v", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer bs.releaseTransaction(tx, err, ctx)

	lo_util := tx.LargeObjects()
	lo, err := lo_util.Open(ctx, blob.DataOID, pgx.LargeObjectModeRead)
	if err != nil {
		return fmt.Errorf("failed to open large object %d: %v", blob.DataOID, err)
	}
	defer lo.Close()

	_, err = utils.CopyBytes(w, lo)
	if err != nil {
		return fmt.Errorf("failed to stream blob data: %v", err)
	}

	return nil
}
