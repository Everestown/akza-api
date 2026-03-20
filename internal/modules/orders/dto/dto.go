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

type VariantShort struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

type OrderResponse struct {
	ID               int64              `json:"id"`
	VariantID        int64              `json:"variant_id"`
	Variant          *VariantShort      `json:"variant,omitempty"`
	CustomerName     string             `json:"customer_name"`
	TelegramUsername string             `json:"telegram_username"`
	Phone            *string            `json:"phone"`
	Comment          *string            `json:"comment"`
	Status           domain.OrderStatus `json:"status"`
	TgNotifiedAt     *time.Time         `json:"tg_notified_at"`
	AllowedNext      []domain.OrderStatus `json:"allowed_next"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
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
		resp.Variant = &VariantShort{ID: o.Variant.ID, Slug: o.Variant.Slug}
	}
	return resp
}
