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

const maxGalleryItems = 12

type repo interface {
	ListByProduct(ctx context.Context, productID int64, onlyPublished bool) ([]domain.ProductVariant, error)
	ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]domain.ProductVariant, error)
	FindBySlug(ctx context.Context, slug string) (*domain.ProductVariant, error)
	FindByID(ctx context.Context, id int64) (*domain.ProductVariant, error)
	SlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, v *domain.ProductVariant) error
	Update(ctx context.Context, v *domain.ProductVariant) error
	SetPublished(ctx context.Context, id int64, pub bool) error
	SoftDelete(ctx context.Context, id int64) error
	AddImage(ctx context.Context, img *domain.VariantImage) error
	FindImage(ctx context.Context, id int64) (*domain.VariantImage, error)
	DeleteImage(ctx context.Context, id int64) error
	ReorderImages(ctx context.Context, ids []int64) error
	CountImages(ctx context.Context, variantID int64) (int64, error)
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

func (s *Service) ListByProduct(ctx context.Context, productID int64, onlyPublished bool) ([]dto.VariantResponse, error) {
	items, err := s.repo.ListByProduct(ctx, productID, onlyPublished)
	if err != nil { return nil, err }
	out := make([]dto.VariantResponse, len(items))
	for i, v := range items { out[i] = dto.FromDomain(&v) }
	return out, nil
}

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

func (s *Service) GetByIDAdmin(ctx context.Context, id int64) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateVariantRequest) (*dto.VariantResponse, error) {
	sl := req.Slug
	if sl == "" {
		sl = slugpkg.GenerateUnique(fmt.Sprintf("variant-%d", req.ProductID), func(c string) bool { return s.repo.SlugExists(ctx, c) })
	} else if s.repo.SlugExists(ctx, sl) {
		return nil, apperror.Conflict(fmt.Sprintf("slug %q already in use", sl))
	}
	v := &domain.ProductVariant{ProductID: req.ProductID, Slug: sl, Attributes: req.Attributes, SortOrder: req.SortOrder}
	if err := s.repo.Create(ctx, v); err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) Update(ctx context.Context, id int64, req dto.UpdateVariantRequest) (*dto.VariantResponse, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	v.Attributes = req.Attributes; v.SortOrder = req.SortOrder
	if err = s.repo.Update(ctx, v); err != nil { return nil, err }
	resp := dto.FromDomain(v); return &resp, nil
}

func (s *Service) SetPublished(ctx context.Context, id int64, pub bool) error {
	return s.repo.SetPublished(ctx, id, pub)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) PresignImage(ctx context.Context, id int64, filename, ct string) (*dto.PresignResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	count, err := s.repo.CountImages(ctx, id)
	if err != nil { return nil, err }
	if count >= maxGalleryItems {
		return nil, apperror.Newf("GALLERY_FULL", 422, fmt.Sprintf("maximum %d images per variant reached", maxGalleryItems))
	}
	key := storage.BuildKey("variants", fmt.Sprintf("%d", id), filename)
	url, err := s.s3.PresignPut(ctx, key, ct)
	if err != nil { return nil, apperror.Newf("S3_ERROR", 502, "could not generate upload URL") }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) ConfirmImage(ctx context.Context, variantID int64, req dto.ConfirmImageRequest) (*dto.ImageResponse, error) {
	if s.s3 == nil { return nil, apperror.Newf("S3_DISABLED", 503, "S3 not configured") }
	if _, err := s.repo.FindByID(ctx, variantID); err != nil { return nil, err }
	count, err := s.repo.CountImages(ctx, variantID)
	if err != nil { return nil, err }
	img := &domain.VariantImage{
		VariantID: variantID, URL: s.s3.PublicURL(req.S3Key),
		S3Key: req.S3Key, SortOrder: int(count),
	}
	if err = s.repo.AddImage(ctx, img); err != nil { return nil, err }
	resp := dto.ImageResponse{ID: img.ID, URL: img.URL, S3Key: img.S3Key, SortOrder: img.SortOrder, CreatedAt: img.CreatedAt}
	return &resp, nil
}

func (s *Service) DeleteImage(ctx context.Context, variantID, imageID int64) error {
	img, err := s.repo.FindImage(ctx, imageID)
	if err != nil { return err }
	if img.VariantID != variantID { return apperror.NotFound("image") }
	if s.s3 != nil { _ = s.s3.DeleteObject(ctx, img.S3Key) }
	return s.repo.DeleteImage(ctx, imageID)
}

func (s *Service) ReorderImages(ctx context.Context, variantID int64, req dto.ReorderImagesRequest) error {
	return s.repo.ReorderImages(ctx, req.IDs)
}
