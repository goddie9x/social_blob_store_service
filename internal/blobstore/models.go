package blobstore

import (
	"time"
)

type Blob struct {
	ID           string `gorm:"primaryKey"`
	FileName     string
	Size         int64
	ContentType  string
	CreatedAt    time.Time
	LastModified time.Time
	DataOID      uint32
}
