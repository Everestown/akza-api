package domain

import "time"

type Collection struct {
	ID          string           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug        string           `gorm:"uniqueIndex;not null"`
	Title       string           `gorm:"not null"`
	Description *string
	CoverURL    *string
	CoverS3Key  *string
	Status      CollectionStatus `gorm:"not null;default:'DRAFT'"`
	ScheduledAt *time.Time
	SortOrder   int              `gorm:"not null;default:0"`
	CreatedAt   time.Time        `gorm:"not null;default:now()"`
	UpdatedAt   time.Time        `gorm:"not null;default:now()"`
	DeletedAt   *time.Time       `gorm:"index"`

	Products []Product `gorm:"foreignKey:CollectionID"`
}

func (Collection) TableName() string { return "collections" }

func (c *Collection) IsVisible() bool {
	if c.Status == CollectionPublished {
		return true
	}
	if c.Status == CollectionScheduled && c.ScheduledAt != nil {
		return !c.ScheduledAt.After(time.Now())
	}
	return false
}

// CanChangeStatus — fixes: ARCHIVED can now return to DRAFT or PUBLISHED.
func (c *Collection) CanChangeStatus(next CollectionStatus) bool {
	switch c.Status {
	case CollectionDraft:
		return next == CollectionScheduled || next == CollectionPublished
	case CollectionScheduled:
		return next == CollectionDraft || next == CollectionPublished
	case CollectionPublished:
		return next == CollectionArchived
	case CollectionArchived:
		// Allow un-archiving
		return next == CollectionDraft || next == CollectionPublished
	}
	return false
}
