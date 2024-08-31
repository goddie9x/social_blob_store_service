package blobstore

import (
	"time"
)

type TargetType string
type BlobType string

const (
	Post          TargetType = "post"
	Comment       TargetType = "comment"
	UserAvatar    TargetType = "user_avatar"
	UserCover     TargetType = "user_cover"
	GroupAvatar   TargetType = "group_avatar"
	GroupCover    TargetType = "group_cover"
	CustomSticker TargetType = "custom_sticker"
	Story         TargetType = "story"
	PageAvatar    TargetType = "page_avatar"
	PageCover     TargetType = "page_cover"
	ChatAvatar    TargetType = "chat_avatar"
	ChatCover     TargetType = "chat_cover"

	Video BlobType = "video"
	Image BlobType = "image"
)

type Blob struct {
	ID           string     `gorm:"primaryKey" form:"id"`
	OwnerID      string     `gorm:"index" form:"-"`
	TargetID     string     `gorm:"index:idx_target_type" form:"targetId"`
	TargetType   TargetType `gorm:"index:idx_target_type" form:"targetType"`
	Type         BlobType   `gorm:"index:idx_target_type" form:"-"`
	FileName     string
	Size         int64
	ContentType  string
	CreatedAt    time.Time
	LastModified time.Time
	DataOID      uint32
}

func (b *Blob) Clone() *Blob {
	var blob Blob
	blob.OwnerID = b.OwnerID
	blob.TargetID = b.TargetID
	blob.TargetType = b.TargetType
	blob.Type = b.Type
	blob.FileName = b.FileName
	blob.Size = b.Size
	blob.ContentType = b.ContentType
	blob.CreatedAt = b.CreatedAt
	blob.LastModified = b.LastModified
	blob.DataOID = b.DataOID
	return &blob
}
