package repository

import (
	"context"
	"errors"
	"time"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }
func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListByProduct(ctx context.Context, productID int64, onlyPublished bool) ([]domain.ProductVariant, error) {
	q := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Where("product_id = ? AND deleted_at IS NULL", productID).Order("sort_order ASC")
	if onlyPublished { q = q.Where("is_published = true") }
	var items []domain.ProductVariant
	return items, q.Find(&items).Error
}

func (r *Repository) ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]domain.ProductVariant, error) {
	sub := r.db.Model(&domain.Product{}).Select("id").Where("slug = ? AND deleted_at IS NULL", productSlug)
	q := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Where("product_id IN (?) AND deleted_at IS NULL", sub).Order("sort_order ASC")
	if onlyPublished { q = q.Where("is_published = true") }
	var items []domain.ProductVariant
	return items, q.Find(&items).Error
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Where("slug = ? AND deleted_at IS NULL", slug).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("variant") }
	return &v, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := r.db.WithContext(ctx).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Where("id = ? AND deleted_at IS NULL", id).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("variant") }
	return &v, err
}

func (r *Repository) SlugExists(ctx context.Context, slug string) bool {
	var c int64
	r.db.WithContext(ctx).Model(&domain.ProductVariant{}).Where("slug = ? AND deleted_at IS NULL", slug).Count(&c)
	return c > 0
}

func (r *Repository) Create(ctx context.Context, v *domain.ProductVariant) error { return r.db.WithContext(ctx).Create(v).Error }
func (r *Repository) Update(ctx context.Context, v *domain.ProductVariant) error { return r.db.WithContext(ctx).Save(v).Error }

func (r *Repository) SetPublished(ctx context.Context, id int64, pub bool) error {
	return r.db.WithContext(ctx).Model(&domain.ProductVariant{}).Where("id = ?", id).Update("is_published", pub).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.ProductVariant{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *Repository) AddImage(ctx context.Context, img *domain.VariantImage) error {
	return r.db.WithContext(ctx).Create(img).Error
}

func (r *Repository) FindImage(ctx context.Context, id int64) (*domain.VariantImage, error) {
	var img domain.VariantImage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&img).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("image") }
	return &img, err
}

func (r *Repository) DeleteImage(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.VariantImage{}, id).Error
}

func (r *Repository) ReorderImages(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&domain.VariantImage{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil { return err }
		}
		return nil
	})
}

func (r *Repository) CountImages(ctx context.Context, variantID int64) (int64, error) {
	var count int64
	return count, r.db.WithContext(ctx).Model(&domain.VariantImage{}).Where("variant_id = ?", variantID).Count(&count).Error
}
