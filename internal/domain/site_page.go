package domain

import "time"

type SitePage struct {
	ID        int64       `gorm:"primaryKey;autoIncrement"`
	Section   PageSection `gorm:"uniqueIndex;not null"`
	Content   JSONB       `gorm:"type:jsonb;not null;default:'{}'"`
	UpdatedAt time.Time   `gorm:"not null;default:now()"`
	UpdatedBy *int64

	Editor *Admin `gorm:"foreignKey:UpdatedBy"`
}

func (SitePage) TableName() string { return "site_pages" }
