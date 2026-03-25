package repository

import (
	"context"
	"errors"
	"time"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }
func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListByCollection(ctx context.Context, collectionID int64, onlyPublished bool, p pagination.CursorPage) ([]domain.Product, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).Where("collection_id = ? AND deleted_at IS NULL", collectionID).
		Order("sort_order ASC, created_at ASC").Limit(limit + 1)
	if onlyPublished { q = q.Where("is_published = true") }
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, id) }
	}
	var items []domain.Product
	return items, q.Find(&items).Error
}

func (r *Repository) ListByCollectionSlug(ctx context.Context, collectionSlug string, onlyPublished bool, p pagination.CursorPage) ([]domain.Product, error) {
	sub := r.db.Model(&domain.Collection{}).Select("id").Where("slug = ? AND deleted_at IS NULL", collectionSlug)
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).Where("collection_id IN (?) AND deleted_at IS NULL", sub).
		Order("sort_order ASC, created_at ASC").Limit(limit + 1)
	if onlyPublished { q = q.Where("is_published = true") }
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, id) }
	}
	var items []domain.Product
	return items, q.Find(&items).Error
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.WithContext(ctx).
		Preload("Variants", "deleted_at IS NULL AND is_published = true").
		Preload("Variants.Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Where("slug = ? AND deleted_at IS NULL", slug).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("product") }
	return &p, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	var p domain.Product
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("product") }
	return &p, err
}

func (r *Repository) SlugExists(ctx context.Context, slug string) bool {
	var c int64
	r.db.WithContext(ctx).Model(&domain.Product{}).Where("slug = ? AND deleted_at IS NULL", slug).Count(&c)
	return c > 0
}

func (r *Repository) Create(ctx context.Context, p *domain.Product) error { return r.db.WithContext(ctx).Create(p).Error }
func (r *Repository) Update(ctx context.Context, p *domain.Product) error { return r.db.WithContext(ctx).Save(p).Error }

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Product{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *Repository) Reorder(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&domain.Product{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil { return err }
		}
		return nil
	})
}

func (r *Repository) UpdateCover(ctx context.Context, id int64, url, s3Key string) error {
	return r.db.WithContext(ctx).Model(&domain.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{"cover_url": url, "cover_s3_key": s3Key}).Error
}

func (r *Repository) SetPublished(ctx context.Context, id int64, published bool) error {
	return r.db.WithContext(ctx).Model(&domain.Product{}).Where("id = ?", id).Update("is_published", published).Error
}

func (r *Repository) ClearCover(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Product{}).Where("id = ?", id).
		Updates(map[string]interface{}{"cover_url": nil, "cover_s3_key": nil}).Error
}
