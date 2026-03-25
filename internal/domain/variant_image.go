package domain

import "time"

// VariantMediaType distinguishes images from videos in the gallery
type VariantMediaType string

const (
	VariantMediaImage VariantMediaType = "IMAGE"
	VariantMediaVideo VariantMediaType = "VIDEO"
)

func (t VariantMediaType) IsVideo() bool { return t == VariantMediaVideo }

type VariantImage struct {
	ID        int64            `gorm:"primaryKey;autoIncrement"`
	VariantID int64            `gorm:"not null"`
	URL       string           `gorm:"not null"`
	S3Key     string           `gorm:"not null"`
	MediaType VariantMediaType `gorm:"type:varchar(10);not null;default:'IMAGE'"`
	SortOrder int              `gorm:"not null;default:0"`
	CreatedAt time.Time        `gorm:"not null;default:now()"`
}

func (VariantImage) TableName() string { return "variant_images" }
