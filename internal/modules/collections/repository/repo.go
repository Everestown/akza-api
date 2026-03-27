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

func (r *Repository) ListPublic(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("(status = 'PUBLISHED') OR (status = 'SCHEDULED' AND scheduled_at <= ?)", time.Now()).
		Order("sort_order ASC, created_at ASC").
		Limit(limit + 1)
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, id) }
	}
	var items []domain.Collection
	return items, q.Find(&items).Error
}

// ListPublicWithScheduled returns published + all scheduled (for drop teaser).
func (r *Repository) ListPublicWithScheduled(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Where("status IN ('PUBLISHED','SCHEDULED')").
		Order("sort_order ASC, created_at ASC").
		Limit(limit + 1)
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, id) }
	}
	var items []domain.Collection
	return items, q.Find(&items).Error
}

func (r *Repository) ListAll(ctx context.Context, p pagination.CursorPage) ([]domain.Collection, error) {
	limit := p.GetLimit()
	q := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("sort_order ASC, created_at ASC").
		Limit(limit + 1)
	if p.Cursor != "" {
		id, createdAt, err := pagination.DecodeCursor(p.Cursor)
		if err == nil { q = q.Where("(created_at > ? OR (created_at = ? AND id > ?))", createdAt, createdAt, id) }
	}
	var items []domain.Collection
	return items, q.Find(&items).Error
}

func (r *Repository) FindBySlug(ctx context.Context, slug string) (*domain.Collection, error) {
	var c domain.Collection
	err := r.db.WithContext(ctx).Where("slug = ? AND deleted_at IS NULL", slug).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("collection") }
	return &c, err
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Collection, error) {
	var c domain.Collection
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("collection") }
	return &c, err
}

func (r *Repository) SlugExists(ctx context.Context, slug string) bool {
	var count int64
	r.db.WithContext(ctx).Model(&domain.Collection{}).Where("slug = ? AND deleted_at IS NULL", slug).Count(&count)
	return count > 0
}

func (r *Repository) Create(ctx context.Context, c *domain.Collection) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *Repository) Update(ctx context.Context, c *domain.Collection) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Collection{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (r *Repository) Reorder(ctx context.Context, ids []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := tx.Model(&domain.Collection{}).Where("id = ?", id).Update("sort_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) UpdateCover(ctx context.Context, id int64, url, s3Key string) error {
	return r.db.WithContext(ctx).Model(&domain.Collection{}).Where("id = ?", id).
		Updates(map[string]interface{}{"cover_url": url, "cover_s3_key": s3Key}).Error
}

func (r *Repository) ClearCover(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Collection{}).Where("id = ?", id).
		Updates(map[string]interface{}{"cover_url": nil, "cover_s3_key": nil}).Error
}

// Restore removes the soft-delete mark from a collection.
func (r *Repository) Restore(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Collection{}).
		Where("id = ?", id).Update("deleted_at", nil).Error
}

// FindDeletedByID finds a soft-deleted collection by ID.
func (r *Repository) FindDeletedByID(ctx context.Context, id int64) (*domain.Collection, error) {
	var c domain.Collection
	err := r.db.WithContext(ctx).Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) { return nil, apperror.NotFound("collection") }
	return &c, err
}

// ListDeleted returns recently soft-deleted collections (last 15 minutes).
func (r *Repository) ListDeleted(ctx context.Context) ([]domain.Collection, error) {
	since := time.Now().Add(-15 * time.Minute)
	var items []domain.Collection
	err := r.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL AND deleted_at >= ?", since).
		Order("deleted_at DESC").Find(&items).Error
	return items, err
}

// PublishScheduledDue sets status=PUBLISHED for all SCHEDULED collections where scheduled_at <= now().
// Returns the number of rows updated.
func (r *Repository) PublishScheduledDue(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Model(&domain.Collection{}).
		Where("status = 'SCHEDULED' AND scheduled_at IS NOT NULL AND scheduled_at <= ? AND deleted_at IS NULL", time.Now()).
		Update("status", domain.CollectionPublished)
	return result.RowsAffected, result.Error
}
