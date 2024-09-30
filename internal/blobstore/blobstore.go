package blobstore

import (
	"blob_store_service/internal/utils"
	"blob_store_service/pkg/middlewares"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"
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
		return nil, fmt.Errorf("failed to connect to database")
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
func (bs *BlobStore) scanBlob(row pgx.Row) (*Blob, error) {
	var blob Blob
	err := row.Scan(
		&blob.ID, &blob.FileName, &blob.Size, &blob.ContentType,
		&blob.CreatedAt, &blob.LastModified, &blob.DataOID,
		&blob.OwnerID, &blob.TargetID, &blob.TargetType, &blob.Type,
	)

	if err != nil {
		return nil, err
	}
	return &blob, nil
}

func (bs *BlobStore) SaveBlob(data io.Reader, blob *Blob) (*Blob, error) {
	fileType := utils.GetValidFileType(blob.ContentType, []string{string(Image), string(Video)})
	if fileType == "" {
		return nil, fmt.Errorf("invalid file type")
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

	now := time.Now()
	blob.ID = uuid.New().String()
	blob.Size = size
	blob.CreatedAt = now
	blob.DataOID = oid
	blob.Type = BlobType(fileType)
	_, err = tx.Exec(ctx,
		`INSERT INTO blobs (id, filename, size, content_type, created_at,
        last_modified, data_oid, owner_id, target_id, target_type, type)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		blob.ID, blob.FileName, blob.Size, blob.ContentType, blob.CreatedAt,
		blob.LastModified, blob.DataOID, blob.OwnerID, blob.TargetID, blob.TargetType, blob.Type,
	)
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

	row := conn.QueryRow(ctx, `
        SELECT * 
        FROM blobs WHERE id = $1
    `, id)
	blob, err := bs.scanBlob(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("blob not found")
		}
		return nil, fmt.Errorf("failed to retrieve blob %s metadata: %v", id, err)
	}

	return blob, nil
}

func (bs *BlobStore) GetListBlobWithPagination(limit, page int, filters map[string]interface{}) ([]*Blob, error) {
	ctx := context.Background()
	conn, err := bs.genConnect(ctx)
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * limit

	query := `
        SELECT *
        FROM blobs 
    `

	whereClauses := []string{}
	args := []interface{}{limit, offset}

	argIdx := 3 // Starting index for additional args (1 and 2 are for limit and offset)
	for key, value := range filters {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", key, argIdx))
		args = append(args, value)
		argIdx++
	}

	if len(whereClauses) > 0 {
		query += "WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY created_at DESC LIMIT $1 OFFSET $2"

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve blob list: %v", err)
	}
	defer rows.Close()

	blobs := []*Blob{}
	for rows.Next() {
		blob, err := bs.scanBlob(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blob row: %v", err)
		}
		blobs = append(blobs, blob)
	}

	return blobs, nil
}

func (bs *BlobStore) DeleteBlob(id string, currentUser middlewares.UserAuth) error {
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
	if blob.OwnerID != currentUser.UserId && currentUser.Role == middlewares.User {
		return fmt.Errorf("You do not have permission")
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
