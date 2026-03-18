package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type CreateVariantRequest struct {
	ProductID  string       `json:"product_id"  binding:"required,uuid"`
	Slug       string       `json:"slug"        binding:"omitempty,slug"`
	Attributes domain.JSONB `json:"attributes"`
	SortOrder  int          `json:"sort_order"`
}

type UpdateVariantRequest struct {
	Attributes domain.JSONB `json:"attributes"`
	SortOrder  int          `json:"sort_order"`
}

type PresignImageRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Filename    string `json:"filename"     binding:"required"`
}

type ConfirmImageRequest struct {
	S3Key        string `json:"s3_key"         binding:"required"`
	OriginalName string `json:"original_name"`
}

type ReorderImagesRequest struct {
	IDs []string `json:"ids" binding:"required,min=1"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

type ImageResponse struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	S3Key     string    `json:"s3_key"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type VariantResponse struct {
	ID          string        `json:"id"`
	ProductID   string        `json:"product_id"`
	Slug        string        `json:"slug"`
	Attributes  domain.JSONB  `json:"attributes"`
	IsPublished bool          `json:"is_published"`
	SortOrder   int           `json:"sort_order"`
	Images      []ImageResponse `json:"images"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func FromDomain(v *domain.ProductVariant) VariantResponse {
	images := make([]ImageResponse, len(v.Images))
	for i, img := range v.Images {
		images[i] = ImageResponse{ID: img.ID, URL: img.URL, S3Key: img.S3Key, SortOrder: img.SortOrder, CreatedAt: img.CreatedAt}
	}
	return VariantResponse{
		ID: v.ID, ProductID: v.ProductID, Slug: v.Slug, Attributes: v.Attributes,
		IsPublished: v.IsPublished, SortOrder: v.SortOrder, Images: images,
		CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt,
	}
}
