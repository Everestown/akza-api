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

func (r *Repository) List(ctx context.Context, status string, p pagination.CursorPage) ([]domain.Order, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).Preload("Variant").Order("created_at DESC").Limit(limit + 1)
	if status != "" { q = q.Where("status = ?", status) }
	if p.Cursor != "" {
		if _, createdAt, err := pagination.DecodeCursor(p.Cursor); err == nil { q = q.Where("created_at < ?", createdAt) }
	}
	var items []domain.Order
	return items, q.Find(&items).Error
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	var o domain.Order
	err := r.db.WithContext(ctx).Preload("Variant").Where("id = ?", id).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("order") }
	return &o, err
}

func (r *Repository) FindVariant(ctx context.Context, variantID string) (*domain.ProductVariant, error) {
	var v domain.ProductVariant
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL AND is_published = true", variantID).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("variant") }
	return &v, err
}

func (r *Repository) Create(ctx context.Context, o *domain.Order) error { return r.db.WithContext(ctx).Create(o).Error }

func (r *Repository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *Repository) SetTgNotified(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Order{}).Where("id = ?", id).
		UpdateColumn("tg_notified_at", gorm.Expr("NOW()")).Error
}
