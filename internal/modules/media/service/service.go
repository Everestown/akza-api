package service

import (
	"context"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/media/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/akza/akza-api/internal/pkg/storage"
)

type repo interface {
	List(ctx context.Context, mediaType string, p pagination.CursorPage) ([]domain.MediaAsset, error)
	FindByID(ctx context.Context, id string) (*domain.MediaAsset, error)
	Create(ctx context.Context, a *domain.MediaAsset) error
	Delete(ctx context.Context, id string) error
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

func (s *Service) List(ctx context.Context, mediaType string, p pagination.CursorPage) (pagination.PageResult[dto.MediaResponse], error) {
	items, err := s.repo.List(ctx, mediaType, p)
	if err != nil { return pagination.PageResult[dto.MediaResponse]{}, err }
	responses := make([]dto.MediaResponse, len(items))
	for i, a := range items { responses[i] = dto.FromDomain(&a) }
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.MediaResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func (s *Service) Presign(ctx context.Context, adminID string, req dto.PresignRequest) (*dto.PresignResponse, error) {
	if s.s3 == nil { return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	key := storage.BuildKey("media", adminID, req.Filename)
	url, err := s.s3.PresignPut(ctx, key, req.ContentType)
	if err != nil { return nil, err }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) Confirm(ctx context.Context, adminID string, req dto.ConfirmRequest) (*dto.MediaResponse, error) {
	if s.s3 == nil { return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	name := req.OriginalName; mime := req.MimeType
	asset := &domain.MediaAsset{
		S3Key: req.S3Key, URL: s.s3.PublicURL(req.S3Key), Type: req.Type,
		OriginalName: &name, SizeBytes: req.SizeBytes, MimeType: &mime, UploadedBy: &adminID,
	}
	if err := s.repo.Create(ctx, asset); err != nil { return nil, err }
	resp := dto.FromDomain(asset); return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	asset, err := s.repo.FindByID(ctx, id)
	if err != nil { return err }
	if s.s3 != nil { _ = s.s3.DeleteObject(ctx, asset.S3Key) }
	return s.repo.Delete(ctx, id)
}
