package dto

import (
	"time"
	"github.com/akza/akza-api/internal/domain"
)

type CreateOrderRequest struct {
	VariantID        int64   `json:"variant_id"        binding:"required"`
	CustomerName     string  `json:"customer_name"     binding:"required,min=2,max=100"`
	TelegramUsername string  `json:"telegram_username" binding:"required,tg_username"`
	Phone            *string `json:"phone"`
	Comment          *string `json:"comment"`
}

type UpdateStatusRequest struct {
	Status domain.OrderStatus `json:"status" binding:"required"`
}

// VariantShort — variant info embedded in order response.
// Includes product and collection slugs so the frontend can build the full client URL:
// /collections/{collection_slug}/{product_slug}/{variant_slug}
type VariantShort struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Slug           string `json:"slug"`
	ProductSlug    string `json:"product_slug"`
	CollectionSlug string `json:"collection_slug"`
}

type OrderResponse struct {
	ID               int64                `json:"id"`
	VariantID        int64                `json:"variant_id"`
	Variant          *VariantShort        `json:"variant,omitempty"`
	CustomerName     string               `json:"customer_name"`
	TelegramUsername string               `json:"telegram_username"`
	Phone            *string              `json:"phone"`
	Comment          *string              `json:"comment"`
	Status           domain.OrderStatus   `json:"status"`
	TgNotifiedAt     *time.Time           `json:"tg_notified_at"`
	AllowedNext      []domain.OrderStatus `json:"allowed_next"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
}

// OrderStats — response for GET /admin/orders/stats
type OrderStats struct {
	Total     int64            `json:"total"`
	ByStatus  map[string]int64 `json:"by_status"`
	// Convenience fields
	New       int64 `json:"new"`
	Contacted int64 `json:"contacted"`
	Confirmed int64 `json:"confirmed"`
	Cancelled int64 `json:"cancelled"`
	Completed int64 `json:"completed"`
}

func FromDomain(o *domain.Order) OrderResponse {
	resp := OrderResponse{
		ID: o.ID, VariantID: o.VariantID,
		CustomerName: o.CustomerName, TelegramUsername: o.TelegramUsername,
		Phone: o.Phone, Comment: o.Comment,
		Status: o.Status, TgNotifiedAt: o.TgNotifiedAt,
		AllowedNext: o.AllowedTransitions(),
		CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
	if o.Variant.ID != 0 {
		v := &VariantShort{
			ID:   o.Variant.ID,
			Name: o.Variant.Name,
			Slug: o.Variant.Slug,
		}
		// Set product and collection slugs from preloaded relations
		if o.Variant.Product.ID != 0 {
			v.ProductSlug = o.Variant.Product.Slug
			if o.Variant.Product.Collection.ID != 0 {
				v.CollectionSlug = o.Variant.Product.Collection.Slug
			}
		}
		resp.Variant = v
	}
	return resp
}
