package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type CreateVariantRequest struct {
	ProductID  int64        `json:"product_id"  binding:"required"`
	Slug       string       `json:"slug"        binding:"omitempty,slug"`
	Attributes domain.JSONB `json:"attributes"`
	SortOrder  int          `json:"sort_order"`
}

type UpdateVariantRequest struct {
	Attributes domain.JSONB `json:"attributes"`
	SortOrder  int          `json:"sort_order"`
}

type ConfirmImageRequest struct {
	S3Key        string `json:"s3_key"        binding:"required"`
	OriginalName string `json:"original_name"`
}

type ReorderImagesRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

type ImageResponse struct {
	ID        int64     `json:"id"`
	URL       string    `json:"url"`
	S3Key     string    `json:"s3_key"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type VariantResponse struct {
	ID          int64        `json:"id"`
	ProductID   int64        `json:"product_id"`
	Slug        string       `json:"slug"`
	Attributes  domain.JSONB `json:"attributes"`
	IsPublished bool         `json:"is_published"`
	SortOrder   int          `json:"sort_order"`
	Images      []ImageResponse `json:"images"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func imageFromDomain(img *domain.VariantImage) ImageResponse {
	return ImageResponse{ID: img.ID, URL: img.URL, S3Key: img.S3Key, SortOrder: img.SortOrder, CreatedAt: img.CreatedAt}
}

func FromDomain(v *domain.ProductVariant) VariantResponse {
	images := make([]ImageResponse, len(v.Images))
	for i, img := range v.Images { images[i] = imageFromDomain(&img) }
	return VariantResponse{
		ID: v.ID, ProductID: v.ProductID, Slug: v.Slug, Attributes: v.Attributes,
		IsPublished: v.IsPublished, SortOrder: v.SortOrder, Images: images,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}
