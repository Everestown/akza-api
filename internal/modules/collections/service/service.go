package service

import (
	"context"
	"fmt"
	"time"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/collections/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	slugpkg "github.com/akza/akza-api/internal/pkg/slug"
	"github.com/akza/akza-api/internal/pkg/storage"
)

type repo interface {
	ListPublicWithScheduled(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error)
	ListAll(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error)
	ListDeleted(ctx context.Context) ([]domain.Collection, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Collection, error)
	FindByID(ctx context.Context, id int64) (*domain.Collection, error)
	FindDeletedByID(ctx context.Context, id int64) (*domain.Collection, error)
	SlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, c *domain.Collection) error
	Update(ctx context.Context, c *domain.Collection) error
	SoftDelete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) error
	Reorder(ctx context.Context, ids []int64) error
	UpdateCover(ctx context.Context, id int64, url, s3Key string) error
	ClearCover(ctx context.Context, id int64) error
	// For scheduler
	PublishScheduledDue(ctx context.Context) (int64, error)
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

func build(items []domain.Collection, limit int) pagination.PageResult[dto.CollectionResponse] {
	responses := make([]dto.CollectionResponse, len(items))
	for i, c := range items { responses[i] = dto.FromDomain(&c) }
	return pagination.BuildResult(responses, limit, func(r dto.CollectionResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	})
}

func (s *Service) ListPublic(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error) {
	items, err := s.repo.ListPublicWithScheduled(ctx, p)
	if err != nil { return pagination.PageResult[dto.CollectionResponse]{}, err }
	return build(items, p.GetLimit()), nil
}

func (s *Service) ListAll(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error) {
	items, err := s.repo.ListAll(ctx, p)
	if err != nil { return pagination.PageResult[dto.CollectionResponse]{}, err }
	return build(items, p.GetLimit()), nil
}

func (s *Service) ListDeleted(ctx context.Context) ([]dto.CollectionResponse, error) {
	items, err := s.repo.ListDeleted(ctx)
	if err != nil { return nil, err }
	out := make([]dto.CollectionResponse, len(items))
	for i, c := range items { out[i] = dto.FromDomain(&c) }
	return out, nil
}

func (s *Service) GetBySlug(ctx context.Context, sl string) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindBySlug(ctx, sl)
	if err != nil { return nil, err }
	if !c.IsVisible() && !c.IsScheduledAndPending() { return nil, apperror.NotFound("collection") }
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) GetByIDAdmin(ctx context.Context, id int64) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateCollectionRequest) (*dto.CollectionResponse, error) {
	sl := req.Slug
	if sl == "" {
		sl = slugpkg.GenerateUnique(req.Title, func(candidate string) bool {
			return s.repo.SlugExists(ctx, candidate)
		})
	} else if s.repo.SlugExists(ctx, sl) {
		return nil, apperror.Conflict(fmt.Sprintf("slug %q already in use", sl))
	}
	c := &domain.Collection{
		Slug: sl, Title: req.Title, Description: req.Description,
		ScheduledAt: req.ScheduledAt.T, SortOrder: req.SortOrder,
		Status: domain.CollectionDraft,
	}
	if err := s.repo.Create(ctx, c); err != nil { return nil, err }
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id int64, req dto.UpdateCollectionRequest) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	c.Title = req.Title; c.Description = req.Description
	c.ScheduledAt = req.ScheduledAt.T; c.SortOrder = req.SortOrder
	if err = s.repo.Update(ctx, c); err != nil { return nil, err }
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, req dto.UpdateStatusRequest) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	if !req.Status.IsValid() { return nil, apperror.Validation("invalid status value") }
	if !c.CanChangeStatus(req.Status) {
		return nil, apperror.Newf("BAD_TRANSITION", 422, "cannot transition from %s to %s", c.Status, req.Status)
	}
	c.Status = req.Status
	if req.ScheduledAt.T != nil { c.ScheduledAt = req.ScheduledAt.T }
	if err = s.repo.Update(ctx, c); err != nil { return nil, err }
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SoftDelete(ctx, id)
}

// Restore un-deletes a collection within the 15-minute window.
func (s *Service) Restore(ctx context.Context, id int64) (*dto.CollectionResponse, error) {
	c, err := s.repo.FindDeletedByID(ctx, id)
	if err != nil { return nil, err }
	// Enforce 15-minute restore window
	if c.DeletedAt != nil && time.Since(*c.DeletedAt) > 15*time.Minute {
		return nil, apperror.Newf("RESTORE_EXPIRED", 422, "restore window (15 min) has expired")
	}
	if err = s.repo.Restore(ctx, id); err != nil { return nil, err }
	c.DeletedAt = nil
	resp := dto.FromDomain(c)
	return &resp, nil
}

func (s *Service) Reorder(ctx context.Context, req dto.ReorderRequest) error {
	return s.repo.Reorder(ctx, req.IDs)
}

func (s *Service) PresignCover(ctx context.Context, id int64, filename, ct string) (*dto.PresignResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	if _, err := s.repo.FindByID(ctx, id); err != nil { return nil, err }
	key := storage.BuildKey("collections", fmt.Sprintf("%d", id), filename)
	url, err := s.s3.PresignPut(ctx, key, ct)
	if err != nil { return nil, apperror.Newf("S3_ERROR", 502, "could not generate upload URL") }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) ConfirmCover(ctx context.Context, id int64, s3Key string) error {
	if s.s3 == nil { return apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	publicURL := s.s3.PublicURL(s3Key)
	return s.repo.UpdateCover(ctx, id, publicURL, s3Key)
}

func (s *Service) DeleteCover(ctx context.Context, id int64) error {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil { return err }
	if c.CoverS3Key != nil && s.s3 != nil {
		_ = s.s3.DeleteObject(ctx, *c.CoverS3Key)
	}
	return s.repo.ClearCover(ctx, id)
}

// PublishScheduledDue transitions all due SCHEDULED collections to PUBLISHED.
// Called by the scheduler goroutine.
func (s *Service) PublishScheduledDue(ctx context.Context) (int64, error) {
	return s.repo.PublishScheduledDue(ctx)
}
