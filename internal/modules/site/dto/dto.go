package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type UpdateContentRequest struct {
	Content domain.JSONB `json:"content" binding:"required"`
}

type SitePageResponse struct {
	ID        string           `json:"id"`
	Section   domain.PageSection `json:"section"`
	Content   domain.JSONB     `json:"content"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func FromDomain(p *domain.SitePage) SitePageResponse {
	return SitePageResponse{ID: p.ID, Section: p.Section, Content: p.Content, UpdatedAt: p.UpdatedAt}
}
