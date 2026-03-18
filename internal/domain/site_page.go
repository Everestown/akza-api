package domain

import "time"

type SitePage struct {
	ID        string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Section   PageSection `gorm:"uniqueIndex;not null"`
	Content   JSONB       `gorm:"type:jsonb;not null;default:'{}'"`
	UpdatedAt time.Time   `gorm:"not null;default:now()"`
	UpdatedBy *string     `gorm:"type:uuid"`

	Editor *Admin `gorm:"foreignKey:UpdatedBy"`
}

func (SitePage) TableName() string { return "site_pages" }
