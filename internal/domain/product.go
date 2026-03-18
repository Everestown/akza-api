package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONB map[string]interface{}

func (j JSONB) Value() (driver.Value, error) {
	if j == nil { return "{}", nil }
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil { *j = JSONB{}; return nil }
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("JSONB: cannot scan type %T", value)
	}
	return json.Unmarshal(bytes, j)
}

type Product struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CollectionID    string     `gorm:"not null"`
	Slug            string     `gorm:"uniqueIndex;not null"`
	Title           string     `gorm:"not null"`
	Description     *string
	Characteristics JSONB      `gorm:"type:jsonb;not null;default:'{}'"`
	Price           float64    `gorm:"type:numeric(12,2);not null"`
	PriceHidden     bool       `gorm:"not null;default:false"`
	CoverURL        *string
	CoverS3Key      *string
	SortOrder       int        `gorm:"not null;default:0"`
	IsPublished     bool       `gorm:"not null;default:false"`
	CreatedAt       time.Time  `gorm:"not null;default:now()"`
	UpdatedAt       time.Time  `gorm:"not null;default:now()"`
	DeletedAt       *time.Time `gorm:"index"`

	Collection Collection       `gorm:"foreignKey:CollectionID"`
	Variants   []ProductVariant `gorm:"foreignKey:ProductID"`
}

func (Product) TableName() string { return "products" }

func (p *Product) IsPublishable() bool {
	for _, v := range p.Variants {
		if v.IsPublished && v.DeletedAt == nil {
			return true
		}
	}
	return false
}
