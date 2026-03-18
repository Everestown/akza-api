package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type PresignRequest struct {
	Filename    string `json:"filename"     binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

type ConfirmRequest struct {
	S3Key        string           `json:"s3_key"         binding:"required"`
	Type         domain.MediaType `json:"type"           binding:"required"`
	OriginalName string           `json:"original_name"`
	SizeBytes    *int64           `json:"size_bytes"`
	MimeType     string           `json:"mime_type"`
}

type MediaResponse struct {
	ID           string           `json:"id"`
	S3Key        string           `json:"s3_key"`
	URL          string           `json:"url"`
	Type         domain.MediaType `json:"type"`
	OriginalName *string          `json:"original_name"`
	SizeBytes    *int64           `json:"size_bytes"`
	MimeType     *string          `json:"mime_type"`
	CreatedAt    time.Time        `json:"created_at"`
}

func FromDomain(a *domain.MediaAsset) MediaResponse {
	return MediaResponse{
		ID: a.ID, S3Key: a.S3Key, URL: a.URL, Type: a.Type,
		OriginalName: a.OriginalName, SizeBytes: a.SizeBytes, MimeType: a.MimeType, CreatedAt: a.CreatedAt,
	}
}
