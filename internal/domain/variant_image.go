package domain

import "time"

type VariantImage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	VariantID int64     `gorm:"not null"`
	URL       string    `gorm:"not null"`
	S3Key     string    `gorm:"not null"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}

func (VariantImage) TableName() string { return "variant_images" }
