package service

import (
	"context"
	"fmt"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/collections/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/akza/akza-api/internal/pkg/slug"
	"github.com/akza/akza-api/internal/pkg/storage"
)

type repo interface {
	ListPublic(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error)
	ListAll(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Collection, error)
	FindByID(ctx context.Context, id string) (*domain.Collection, error)
	SlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, c *domain.Collection) error
	Update(ctx context.Context, c *domain.Collection) error
	SoftDelete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
	UpdateCover(ctx context.Context, id, url, s3Key string) error
}

type Service struct {
	repo repo
	s3   *storage.Client
}

func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

func (s *Service) ListPublic(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error) {
	items, err := s.repo.ListPublic(ctx, p)
	if err != nil {
		return pagination.PageResult[dto.CollectionResponse]{}, err
	}
	responses := make([]dto.CollectionResponse, len(items))
	for i, c := range items {
		responses[i] = dto.FromDomain(&c)
	}
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.CollectionResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func (s *Service) ListAll(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error) {
	items, err := s.repo.ListAll(ctx, p)
	if err != nil {
		return pagination.PageResult[dto.CollectionResponse]{}, err
	}
	responses := make([]dto.CollectionResponse, len(items))
	for i, c := range items {
		responses[i] = dto.FromDomain(&c)
	}
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.CollectionResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func (s *Service) GetBySlug(ctx context.Context, sl string) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindBySlug(ctx, sl)
	if err != nil {
		return nil, err
	}
	if !c.IsVisible() {
		return nil, apperror.NotFound("collection")
	}
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) GetByIDAdmin(ctx context.Context, id string) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateCollectionRequest) (*dto.CollectionResponse, error) {
	sl := req.Slug
	if sl == "" {
		sl = slug.GenerateUnique(req.Title, func(candidate string) bool {
			return s.repo.SlugExists(ctx, candidate)
		})
	} else if s.repo.SlugExists(ctx, sl) {
		return nil, apperror.Conflict(fmt.Sprintf("slug %q already in use", sl))
	}

	c := &domain.Collection{
		Slug:        sl,
		Title:       req.Title,
		Description: req.Description,
		ScheduledAt: req.ScheduledAt.T,
		SortOrder:   req.SortOrder,
		Status:      domain.CollectionDraft,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateCollectionRequest) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Title = req.Title
	c.Description = req.Description
	c.ScheduledAt = req.ScheduledAt.T
	c.SortOrder = req.SortOrder
	if err = s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req dto.UpdateStatusRequest) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !req.Status.IsValid() {
		return nil, apperror.Validation("invalid status value")
	}
	if !c.CanChangeStatus(req.Status) {
		return nil, apperror.Newf("BAD_TRANSITION", 422,
			"cannot transition from %s to %s", c.Status, req.Status)
	}
	c.Status = req.Status
	if req.ScheduledAt.T != nil {
		c.ScheduledAt = req.ScheduledAt.T
	}
	if err = s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Reorder(ctx context.Context, req dto.ReorderRequest) error {
	return s.repo.Reorder(ctx, req.IDs)
}

func (s *Service) PresignCover(ctx context.Context, id, filename, contentType string) (*dto.PresignResponse, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}
	if s.s3 == nil {
		return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 storage not configured")
	}
	key := storage.BuildKey("collections", id, filename)
	url, err := s.s3.PresignPut(ctx, key, contentType)
	if err != nil {
		return nil, err
	}
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) ConfirmCover(ctx context.Context, id, s3Key string) error {
	if s.s3 == nil {
		return apperror.New("S3_NOT_CONFIGURED", 503, "S3 storage not configured")
	}
	url := s.s3.PublicURL(s3Key)
	return s.repo.UpdateCover(ctx, id, url, s3Key)
}
