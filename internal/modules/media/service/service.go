package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/media/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/akza/akza-api/internal/pkg/storage"
)

// Max file sizes in bytes
const (
	maxImageBytes = 40 * 1024 * 1024  // 40 MB
	minImageBytes = 5 * 1024 * 1024   // 5 MB
	maxVideoBytes = 100 * 1024 * 1024 // 100 MB
	minVideoBytes = 5 * 1024 * 1024   // 5 MB
)

type repo interface {
	List(ctx context.Context, mediaType string, p pagination.CursorPage) ([]domain.MediaAsset, error)
	FindByID(ctx context.Context, id int64) (*domain.MediaAsset, error)
	Create(ctx context.Context, a *domain.MediaAsset) error
	Delete(ctx context.Context, id int64) error
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

func (s *Service) Presign(ctx context.Context, req dto.PresignRequest, adminIDStr string) (*dto.PresignResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
	_ = adminID
	ts := strconv.FormatInt(time.Now().UnixNano(), 36)
	key := storage.BuildKey("media", ts, req.Filename)
	url, err := s.s3.PresignPut(ctx, key, req.ContentType)
	if err != nil { return nil, apperror.Newf("S3_ERROR", 502, "could not generate upload URL") }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) Confirm(ctx context.Context, req dto.ConfirmRequest, adminIDStr string) (*dto.MediaResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	// Validate file size if provided
	if req.SizeBytes != nil {
		switch req.Type {
		case domain.MediaImage:
			if *req.SizeBytes < minImageBytes { return nil, apperror.Newf("FILE_TOO_SMALL", 422, fmt.Sprintf("image must be at least %d MB", minImageBytes/1024/1024)) }
			if *req.SizeBytes > maxImageBytes { return nil, apperror.Newf("FILE_TOO_LARGE", 422, fmt.Sprintf("image cannot exceed %d MB", maxImageBytes/1024/1024)) }
		case domain.MediaVideo:
			if *req.SizeBytes < minVideoBytes { return nil, apperror.Newf("FILE_TOO_SMALL", 422, fmt.Sprintf("video must be at least %d MB", minVideoBytes/1024/1024)) }
			if *req.SizeBytes > maxVideoBytes { return nil, apperror.Newf("FILE_TOO_LARGE", 422, fmt.Sprintf("video cannot exceed %d MB", maxVideoBytes/1024/1024)) }
		}
	}
	adminID, _ := strconv.ParseInt(adminIDStr, 10, 64)
	originalName := req.OriginalName
	mimeType := req.MimeType
	asset := &domain.MediaAsset{
		S3Key: req.S3Key, URL: s.s3.PublicURL(req.S3Key), Type: req.Type,
		OriginalName: &originalName, SizeBytes: req.SizeBytes, MimeType: &mimeType,
		UploadedBy: &adminID,
	}
	if err := s.repo.Create(ctx, asset); err != nil { return nil, err }
	resp := dto.FromDomain(asset)
	return &resp, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	asset, err := s.repo.FindByID(ctx, id)
	if err != nil { return err }
	if s.s3 != nil { _ = s.s3.DeleteObject(ctx, asset.S3Key) }
	return s.repo.Delete(ctx, id)
}
