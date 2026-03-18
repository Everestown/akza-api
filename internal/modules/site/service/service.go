package service

import (
	"context"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/site/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
)

type repo interface {
	GetAll(ctx context.Context) ([]domain.SitePage, error)
	GetBySection(ctx context.Context, section domain.PageSection) (*domain.SitePage, error)
	Upsert(ctx context.Context, section domain.PageSection, content domain.JSONB, adminID string) (*domain.SitePage, error)
}

type Service struct{ repo repo }
func New(repo repo) *Service { return &Service{repo: repo} }

func (s *Service) GetAll(ctx context.Context) ([]dto.SitePageResponse, error) {
	pages, err := s.repo.GetAll(ctx)
	if err != nil { return nil, err }
	out := make([]dto.SitePageResponse, len(pages))
	for i, p := range pages { out[i] = dto.FromDomain(&p) }
	return out, nil
}

func (s *Service) GetBySection(ctx context.Context, section string) (*dto.SitePageResponse, error) {
	sec := domain.PageSection(section)
	if !sec.IsValid() { return nil, apperror.Validation("invalid section") }
	p, err := s.repo.GetBySection(ctx, sec)
	if err != nil { return nil, err }
	resp := dto.FromDomain(p); return &resp, nil
}

func (s *Service) UpdateSection(ctx context.Context, section, adminID string, req dto.UpdateContentRequest) (*dto.SitePageResponse, error) {
	sec := domain.PageSection(section)
	if !sec.IsValid() { return nil, apperror.Validation("invalid section") }
	p, err := s.repo.Upsert(ctx, sec, req.Content, adminID)
	if err != nil { return nil, err }
	resp := dto.FromDomain(p); return &resp, nil
}
