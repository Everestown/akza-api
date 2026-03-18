package service

import (
	"context"
	"fmt"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/variants/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	slugpkg "github.com/akza/akza-api/internal/pkg/slug"
	"github.com/akza/akza-api/internal/pkg/storage"
)

type repo interface {
	ListByProduct(ctx context.Context, productID string, onlyPublished bool) ([]domain.ProductVariant, error)
	ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]domain.ProductVariant, error)
	FindBySlug(ctx context.Context, slug string) (*domain.ProductVariant, error)
	FindByID(ctx context.Context, id string) (*domain.ProductVariant, error)
	SlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, v *domain.ProductVariant) error
	Update(ctx context.Context, v *domain.ProductVariant) error
	SetPublished(ctx context.Context, id string, pub bool) error
	SoftDelete(ctx context.Context, id string) error
	AddImage(ctx context.Context, img *domain.VariantImage) error
	FindImage(ctx context.Context, id string) (*domain.VariantImage, error)
	DeleteImage(ctx context.Context, id string) error
	ReorderImages(ctx context.Context, ids []string) error
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

// ListByProduct filters by product UUID (used by admin).
func (s *Service) ListByProduct(ctx context.Context, productID string, onlyPublished bool) ([]dto.VariantResponse, error) {
	items, err := s.repo.ListByProduct(ctx, productID, onlyPublished)
	if err != nil { return nil, err }
	out := make([]dto.VariantResponse, len(items))
	for i, v := range items { out[i] = dto.FromDomain(&v) }
	return out, nil
}

// ListByProductSlug filters by product slug (used by public API).
func (s *Service) ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]dto.VariantResponse, error) {
	items, err := s.repo.ListByProductSlug(ctx, productSlug, onlyPublished)
	if err != nil { return nil, err }
	out := make([]dto.VariantResponse, len(items))
	for i, v := range items { out[i] = dto.FromDomain(&v) }
	return out, nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*dto.VariantResponse, error) {
	v, err := s.repo.FindBySlug(ctx, slug)
	if err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) GetByIDAdmin(ctx context.Context, id string) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateVariantRequest) (*dto.VariantResponse, error) {
	sl := req.Slug
	if sl == "" {
		sl = slugpkg.GenerateUnique(req.ProductID, func(c string) bool { return s.repo.SlugExists(ctx, c) })
	} else if s.repo.SlugExists(ctx, sl) {
		return nil, apperror.Conflict(fmt.Sprintf("slug %q already in use", sl))
	}
	v := &domain.ProductVariant{ProductID: req.ProductID, Slug: sl, Attributes: req.Attributes, SortOrder: req.SortOrder}
	if err := s.repo.Create(ctx, v); err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateVariantRequest) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	v.Attributes = req.Attributes; v.SortOrder = req.SortOrder
	if err = s.repo.Update(ctx, v); err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) SetPublished(ctx context.Context, id string, pub bool) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SetPublished(ctx, id, pub)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) PresignImage(ctx context.Context, id, filename, ct string) (*dto.PresignResponse, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return nil, err }
	if s.s3 == nil { return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	key := storage.BuildKey("variants", id, filename)
	url, err := s.s3.PresignPut(ctx, key, ct)
	if err != nil { return nil, err }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) ConfirmImage(ctx context.Context, variantID string, req dto.ConfirmImageRequest) (*dto.ImageResponse, error) {
	if s.s3 == nil { return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	v, err := s.repo.FindByID(ctx, variantID)
	if err != nil { return nil, err }
	img := &domain.VariantImage{VariantID: v.ID, S3Key: req.S3Key, URL: s.s3.PublicURL(req.S3Key), SortOrder: len(v.Images)}
	if err = s.repo.AddImage(ctx, img); err != nil { return nil, err }
	return &dto.ImageResponse{ID: img.ID, URL: img.URL, S3Key: img.S3Key, SortOrder: img.SortOrder, CreatedAt: img.CreatedAt}, nil
}

func (s *Service) DeleteImage(ctx context.Context, variantID, imageID string) error {
	img, err := s.repo.FindImage(ctx, imageID)
	if err != nil { return err }
	if img.VariantID != variantID { return apperror.ErrForbidden }
	if s.s3 != nil { _ = s.s3.DeleteObject(ctx, img.S3Key) }
	return s.repo.DeleteImage(ctx, imageID)
}

func (s *Service) ReorderImages(ctx context.Context, variantID string, req dto.ReorderImagesRequest) error {
	if _, err := s.repo.FindByID(ctx, variantID); err != nil { return err }
	return s.repo.ReorderImages(ctx, req.IDs)
}
