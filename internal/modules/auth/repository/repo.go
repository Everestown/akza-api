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
	var a domain.Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.Unauthorized("invalid credentials") }
	return &a, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Admin, error) {
	var a domain.Admin
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("admin") }
	return &a, err
}
