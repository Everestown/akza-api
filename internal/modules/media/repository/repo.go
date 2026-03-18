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

func (r *Repository) List(ctx context.Context, mediaType string, p pagination.CursorPage) ([]domain.MediaAsset, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit + 1)
	if mediaType != "" { q = q.Where("type = ?", mediaType) }
	if p.Cursor != "" {
		if _, createdAt, err := pagination.DecodeCursor(p.Cursor); err == nil { q = q.Where("created_at < ?", createdAt) }
	}
	var items []domain.MediaAsset
	return items, q.Find(&items).Error
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.MediaAsset, error) {
	var a domain.MediaAsset
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("media asset") }
	return &a, err
}

func (r *Repository) Create(ctx context.Context, a *domain.MediaAsset) error { return r.db.WithContext(ctx).Create(a).Error }
func (r *Repository) Delete(ctx context.Context, id string) error { return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.MediaAsset{}).Error }
