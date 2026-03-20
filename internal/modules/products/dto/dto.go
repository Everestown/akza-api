package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type CreateProductRequest struct {
	CollectionID    int64        `json:"collection_id"    binding:"required"`
	Title           string       `json:"title"            binding:"required,min=2,max=100"`
	Slug            string       `json:"slug"             binding:"omitempty,slug"`
	Description     *string      `json:"description"`
	Characteristics domain.JSONB `json:"characteristics"`
	Price           float64      `json:"price"            binding:"required,gt=0"`
	PriceHidden     bool         `json:"price_hidden"`
	SortOrder       int          `json:"sort_order"`
}

type UpdateProductRequest struct {
	Title           string       `json:"title"            binding:"required,min=2,max=100"`
	Description     *string      `json:"description"`
	Characteristics domain.JSONB `json:"characteristics"`
	Price           float64      `json:"price"            binding:"required,gt=0"`
	PriceHidden     bool         `json:"price_hidden"`
	SortOrder       int          `json:"sort_order"`
}

type ReorderRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

type PresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Filename    string `json:"filename"     binding:"required"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

type ProductResponse struct {
	ID              int64        `json:"id"`
	CollectionID    int64        `json:"collection_id"`
	Slug            string       `json:"slug"`
	Title           string       `json:"title"`
	Description     *string      `json:"description"`
	Characteristics domain.JSONB `json:"characteristics"`
	Price           float64      `json:"price"`
	PriceHidden     bool         `json:"price_hidden"`
	CoverURL        *string      `json:"cover_url"`
	SortOrder       int          `json:"sort_order"`
	IsPublished     bool         `json:"is_published"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func FromDomain(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID: p.ID, CollectionID: p.CollectionID, Slug: p.Slug, Title: p.Title,
		Description: p.Description, Characteristics: p.Characteristics,
		Price: p.Price, PriceHidden: p.PriceHidden,
		CoverURL: p.CoverURL, SortOrder: p.SortOrder, IsPublished: p.IsPublished,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}
