package service

import (
	"context"
	"fmt"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/products/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	slugpkg "github.com/akza/akza-api/internal/pkg/slug"
	"github.com/akza/akza-api/internal/pkg/storage"
)

type repo interface {
	ListByCollection(ctx context.Context, collectionID string, onlyPublished bool, p pagination.CursorPage) ([]domain.Product, error)
	ListByCollectionSlug(ctx context.Context, collectionSlug string, onlyPublished bool, p pagination.CursorPage) ([]domain.Product, error)
	FindBySlug(ctx context.Context, slug string) (*domain.Product, error)
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	SlugExists(ctx context.Context, slug string) bool
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, p *domain.Product) error
	SoftDelete(ctx context.Context, id string) error
	Reorder(ctx context.Context, ids []string) error
	UpdateCover(ctx context.Context, id, url, s3Key string) error
	SetPublished(ctx context.Context, id string, published bool) error
}

type Service struct{ repo repo; s3 *storage.Client }
func New(repo repo, s3 *storage.Client) *Service { return &Service{repo: repo, s3: s3} }

// ListByCollection is used by the admin panel (filters by UUID).
func (s *Service) ListByCollection(ctx context.Context, collectionID string, onlyPublished bool, p pagination.CursorPage) (pagination.PageResult[dto.ProductResponse], error) {
	items, err := s.repo.ListByCollection(ctx, collectionID, onlyPublished, p)
	if err != nil { return pagination.PageResult[dto.ProductResponse]{}, err }
	responses := make([]dto.ProductResponse, len(items))
	for i, p := range items { responses[i] = dto.FromDomain(&p) }
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.ProductResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

// ListByCollectionSlug is used by the public API (filters by collection slug).
func (s *Service) ListByCollectionSlug(ctx context.Context, collectionSlug string, onlyPublished bool, p pagination.CursorPage) (pagination.PageResult[dto.ProductResponse], error) {
	items, err := s.repo.ListByCollectionSlug(ctx, collectionSlug, onlyPublished, p)
	if err != nil { return pagination.PageResult[dto.ProductResponse]{}, err }
	responses := make([]dto.ProductResponse, len(items))
	for i, p := range items { responses[i] = dto.FromDomain(&p) }
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.ProductResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func (s *Service) GetBySlug(ctx context.Context, slug string) (*dto.ProductResponse, error) {
	p, err := s.repo.FindBySlug(ctx, slug)
	if err != nil { return nil, err }
	resp := dto.FromDomain(p)
	return &resp, nil
}

func (s *Service) GetByIDAdmin(ctx context.Context, id string) (*dto.ProductResponse, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	resp := dto.FromDomain(p)
	return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	sl := req.Slug
	if sl == "" {
		sl = slugpkg.GenerateUnique(req.Title, func(c string) bool { return s.repo.SlugExists(ctx, c) })
	} else if s.repo.SlugExists(ctx, sl) {
		return nil, apperror.Conflict(fmt.Sprintf("slug %q already in use", sl))
	}
	product := &domain.Product{
		CollectionID: req.CollectionID, Slug: sl, Title: req.Title,
		Description: req.Description, Characteristics: req.Characteristics,
		Price: req.Price, PriceHidden: req.PriceHidden, SortOrder: req.SortOrder,
	}
	if err := s.repo.Create(ctx, product); err != nil { return nil, err }
	resp := dto.FromDomain(product)
	return &resp, nil
}

func (s *Service) Update(ctx context.Context, id string, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	p.Title = req.Title; p.Description = req.Description
	p.Characteristics = req.Characteristics; p.Price = req.Price; p.PriceHidden = req.PriceHidden; p.SortOrder = req.SortOrder
	if err = s.repo.Update(ctx, p); err != nil { return nil, err }
	resp := dto.FromDomain(p)
	return &resp, nil
}

func (s *Service) SetPublished(ctx context.Context, id string, published bool) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SetPublished(ctx, id, published)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return err }
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service) Reorder(ctx context.Context, req dto.ReorderRequest) error { return s.repo.Reorder(ctx, req.IDs) }

func (s *Service) PresignCover(ctx context.Context, id, filename, ct string) (*dto.PresignResponse, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil { return nil, err }
	if s.s3 == nil { return nil, apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	key := storage.BuildKey("products", id, filename)
	url, err := s.s3.PresignPut(ctx, key, ct)
	if err != nil { return nil, err }
	return &dto.PresignResponse{UploadURL: url, S3Key: key}, nil
}

func (s *Service) ConfirmCover(ctx context.Context, id, s3Key string) error {
	if s.s3 == nil { return apperror.New("S3_NOT_CONFIGURED", 503, "S3 not configured") }
	return s.repo.UpdateCover(ctx, id, s.s3.PublicURL(s3Key), s3Key)
}
