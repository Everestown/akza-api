package domain

import "time"

type MediaAsset struct {
	ID           int64     `gorm:"primaryKey;autoIncrement"`
	S3Key        string    `gorm:"uniqueIndex;not null"`
	URL          string    `gorm:"not null"`
	Type         MediaType `gorm:"not null"`
	OriginalName *string
	SizeBytes    *int64
	MimeType     *string
	UploadedBy   *int64
	CreatedAt    time.Time `gorm:"not null;default:now()"`

	Uploader *Admin `gorm:"foreignKey:UploadedBy"`
}

func (MediaAsset) TableName() string { return "media_assets" }
