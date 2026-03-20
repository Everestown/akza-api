package dto

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/akza/akza-api/internal/domain"
)

// FlexTime parses time.Time from HTML datetime-local and RFC3339.
type FlexTime struct{ T *time.Time }

func (f *FlexTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil { return nil }
	if s == "" { return nil }
	formats := []string{
		time.RFC3339, "2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02",
	}
	clean := strings.TrimRight(s, "Z")
	for _, fmt := range formats {
		if t, err := time.ParseInLocation(fmt, clean, time.UTC); err == nil {
			f.T = &t; return nil
		}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil { f.T = &t }
	return nil
}

// ── Requests ────────────────────────────────────────────────────────────────

type CreateCollectionRequest struct {
	Title       string   `json:"title"        binding:"required,min=2,max=100"`
	Slug        string   `json:"slug"         binding:"omitempty,slug"`
	Description *string  `json:"description"`
	ScheduledAt FlexTime `json:"scheduled_at"`
	SortOrder   int      `json:"sort_order"`
}

type UpdateCollectionRequest struct {
	Title       string   `json:"title"        binding:"required,min=2,max=100"`
	Description *string  `json:"description"`
	ScheduledAt FlexTime `json:"scheduled_at"`
	SortOrder   int      `json:"sort_order"`
}

type UpdateStatusRequest struct {
	Status      domain.CollectionStatus `json:"status" binding:"required"`
	ScheduledAt FlexTime                `json:"scheduled_at"`
}

type ReorderRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1"`
}

type PresignRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	Filename    string `json:"filename"     binding:"required"`
}

// ── Responses ───────────────────────────────────────────────────────────────

type CollectionResponse struct {
	ID          int64                   `json:"id"`
	Slug        string                  `json:"slug"`
	Title       string                  `json:"title"`
	Description *string                 `json:"description"`
	CoverURL    *string                 `json:"cover_url"`
	Status      domain.CollectionStatus `json:"status"`
	ScheduledAt *time.Time              `json:"scheduled_at"`
	// IsPending is true when status=SCHEDULED and time has NOT yet come.
	// The client shows a teaser (locked) state.
	IsPending   bool                    `json:"is_pending"`
	SortOrder   int                     `json:"sort_order"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	S3Key     string `json:"s3_key"`
}

func FromDomain(c *domain.Collection) CollectionResponse {
	return CollectionResponse{
		ID: c.ID, Slug: c.Slug, Title: c.Title, Description: c.Description,
		CoverURL: c.CoverURL, Status: c.Status, ScheduledAt: c.ScheduledAt,
		IsPending: c.IsScheduledAndPending(),
		SortOrder: c.SortOrder, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}
