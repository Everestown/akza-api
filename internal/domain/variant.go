package domain

import "time"

type ProductVariant struct {
	ID          int64      `gorm:"primaryKey;autoIncrement"`
	ProductID   int64      `gorm:"not null"`
	Slug        string     `gorm:"uniqueIndex;not null"`
	Attributes  JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	IsPublished bool       `gorm:"not null;default:false"`
	SortOrder   int        `gorm:"not null;default:0"`
	CreatedAt   time.Time  `gorm:"not null;default:now()"`
	UpdatedAt   time.Time  `gorm:"not null;default:now()"`
	DeletedAt   *time.Time `gorm:"index"`

	Product Product        `gorm:"foreignKey:ProductID"`
	Images  []VariantImage `gorm:"foreignKey:VariantID"`
}

func (ProductVariant) TableName() string { return "product_variants" }

func (v *ProductVariant) IsPublishable() bool { return len(v.Images) > 0 }
