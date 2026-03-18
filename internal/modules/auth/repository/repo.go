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

func (r *Repository) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ErrUnauthorized
		}
		return nil, err
	}
	return &admin, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Admin, error) {
	var admin domain.Admin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("admin")
		}
		return nil, err
	}
	return &admin, nil
}
