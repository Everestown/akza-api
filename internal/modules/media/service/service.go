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

const (
	maxImageBytes     = 50 * 1024 * 1024
	minImageBytes     = 50 * 1024
	maxVideoBytes     = 100 * 1024 * 1024
	minVideoBytes     = 5 * 1024 * 1024
	maxSiteImageBytes = 100 * 1024 * 1024
	maxSiteVideoBytes = 200 * 1024 * 1024
)

type MediaContext string
const (
	ContextGallery MediaContext = "gallery"
	ContextSite    MediaContext = "site"
)

type repo interface {
	List(ctx context.Context, mediaType, folder string, p pagination.CursorPage) ([]domain.MediaAsset, error)
	FindByID(ctx context.Context, id int64) (*domain.MediaAsset, error)
	Create(ctx context.Context, a *domain.MediaAsset) error
	Delete(ctx context.Context, id int64) error
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

func (s *Service) List(ctx context.Context, mediaType, folder string, p pagination.CursorPage) (pagination.PageResult[dto.MediaResponse], error) {
	items, err := s.repo.List(ctx, mediaType, folder, p)
	if err != nil { return pagination.PageResult[dto.MediaResponse]{}, err }
	responses := make([]dto.MediaResponse, len(items))
	for i, a := range items { responses[i] = dto.FromDomain(&a) }
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.MediaResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func validateSize(mediaType domain.MediaType, sizeBytes int64, ctx2 MediaContext) error {
	switch mediaType {
	case domain.MediaImage:
		if sizeBytes < minImageBytes { return apperror.Newf("FILE_TOO_SMALL", 422, fmt.Sprintf("изображение слишком маленькое, минимум %d KB", minImageBytes/1024)) }
		maxImg := int64(maxImageBytes)
		if ctx2 == ContextSite { maxImg = maxSiteImageBytes }
		if sizeBytes > maxImg { return apperror.Newf("FILE_TOO_LARGE", 422, fmt.Sprintf("изображение слишком большое, максимум %d MB", maxImg/1024/1024)) }
	case domain.MediaVideo:
		if sizeBytes < minVideoBytes { return apperror.Newf("FILE_TOO_SMALL", 422, fmt.Sprintf("видео слишком маленькое, минимум %d MB", minVideoBytes/1024/1024)) }
		maxVid := int64(maxVideoBytes)
		if ctx2 == ContextSite { maxVid = maxSiteVideoBytes }
		if sizeBytes > maxVid { return apperror.Newf("FILE_TOO_LARGE", 422, fmt.Sprintf("видео слишком большое, максимум %d MB", maxVid/1024/1024)) }
	}
	return nil
}

func (s *Service) Presign(ctx context.Context, req dto.PresignRequest, adminIDStr string) (*dto.PresignResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	ts := strconv.FormatInt(time.Now().UnixNano(), 36)
	key := storage.BuildKey("media", ts, req.Filename)
	url, err := s.s3.PresignPut(ctx, key, req.ContentType)
	if err != nil { return nil, apperror.Newf("S3_ERROR", 502, "could not generate upload URL") }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) Confirm(ctx context.Context, req dto.ConfirmRequest, adminIDStr string) (*dto.MediaResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	ctx2 := ContextGallery
	if req.Context == "site" { ctx2 = ContextSite }
	if req.SizeBytes != nil && *req.SizeBytes > 0 {
		if err := validateSize(req.Type, *req.SizeBytes, ctx2); err != nil { return nil, err }
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
