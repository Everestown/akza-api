package domain

import "time"

type MediaAsset struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	S3Key        string    `gorm:"uniqueIndex;not null"`
	URL          string    `gorm:"not null"`
	Type         MediaType `gorm:"not null"`
	OriginalName *string
	SizeBytes    *int64
	MimeType     *string
	UploadedBy   *string   `gorm:"type:uuid"`
	CreatedAt    time.Time `gorm:"not null;default:now()"`

	Uploader *Admin `gorm:"foreignKey:UploadedBy"`
}

func (MediaAsset) TableName() string { return "media_assets" }
