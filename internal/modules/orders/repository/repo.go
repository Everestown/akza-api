package repository

import (
	"context"
	"errors"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }
func New(db *gorm.DB) *Repository { return &Repository{db: db} }

// preloadOrder sets deep joins: Variant → Product → Collection (to build client URL)
func preloadOrder(q *gorm.DB) *gorm.DB {
	return q.
		Preload("Variant").
		Preload("Variant.Product").
		Preload("Variant.Product.Collection")
}

func (r *Repository) List(ctx context.Context, statusFilter string, p pagination.CursorPage) ([]domain.Order, error) {
	limit := p.GetLimit()
	q := preloadOrder(r.db.WithContext(ctx)).Order("created_at DESC").Limit(limit + 1)
	if statusFilter != "" { q = q.Where("status = ?", statusFilter) }
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at < ? OR (created_at = ? AND id < ?))", createdAt, createdAt, id) }
	}
	var items []domain.Order
	return items, q.Find(&items).Error
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	var o domain.Order
	err := preloadOrder(r.db.WithContext(ctx)).Where("id = ?", id).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("order") }
	return &o, err
}

func (r *Repository) FindVariant(ctx context.Context, variantID int64) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL AND is_published = true", variantID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("variant") }
	return &v, err
}

// Stats returns counts grouped by order status.
type StatusCount struct {
	Status string `gorm:"column:status"`
	Count  int64  `gorm:"column:count"`
}

func (r *Repository) Stats(ctx context.Context) ([]StatusCount, error) {
	var result []StatusCount
	err := r.db.WithContext(ctx).
		Model(&domain.Order{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&result).Error
	return result, err
}

func (r *Repository) Create(ctx context.Context, o *domain.Order) error { return r.db.WithContext(ctx).Create(o).Error }
func (r *Repository) Update(ctx context.Context, o *domain.Order) error { return r.db.WithContext(ctx).Save(o).Error }
