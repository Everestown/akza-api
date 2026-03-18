package repository

import (
	"context"
	"errors"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }
func New(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) GetAll(ctx context.Context) ([]domain.SitePage, error) {
	var pages []domain.SitePage
	return pages, r.db.WithContext(ctx).Find(&pages).Error
}

func (r *Repository) GetBySection(ctx context.Context, section domain.PageSection) (*domain.SitePage, error) {
	var p domain.SitePage
	err := r.db.WithContext(ctx).Where("section = ?", section).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("site page") }
	return &p, err
}

func (r *Repository) Upsert(ctx context.Context, section domain.PageSection, content domain.JSONB, adminID string) (*domain.SitePage, error) {
	var p domain.SitePage
	if err := r.db.WithContext(ctx).Where("section = ?", section).First(&p).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) { return nil, err }
		p = domain.SitePage{Section: section, Content: content, UpdatedBy: &adminID}
		return &p, r.db.WithContext(ctx).Create(&p).Error
	}
	p.Content = content; p.UpdatedBy = &adminID
	return &p, r.db.WithContext(ctx).Save(&p).Error
}
