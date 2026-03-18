package service

import (
	"context"

	"github.com/akza/akza-api/internal/domain"
	"github.com/akza/akza-api/internal/modules/orders/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/akza/akza-api/internal/pkg/telegram"
)

type repo interface {
	List(ctx context.Context, status string, p pagination.CursorPage) ([]domain.Order, error)
	FindByID(ctx context.Context, id string) (*domain.Order, error)
	FindVariant(ctx context.Context, variantID string) (*domain.ProductVariant, error)
	Create(ctx context.Context, o *domain.Order) error
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error
	SetTgNotified(ctx context.Context, id string) error
}

type Service struct{ repo repo; tg *telegram.Bot; siteBase string }
func New(repo repo, tg *telegram.Bot, siteBase string) *Service { return &Service{repo: repo, tg: tg, siteBase: siteBase} }

func (s *Service) List(ctx context.Context, q dto.ListOrdersQuery) (pagination.PageResult[dto.OrderResponse], error) {
	p := pagination.CursorPage{Cursor: q.Cursor, Limit: q.Limit}
	items, err := s.repo.List(ctx, q.Status, p)
	if err != nil { return pagination.PageResult[dto.OrderResponse]{}, err }
	responses := make([]dto.OrderResponse, len(items))
	for i, o := range items { responses[i] = dto.FromDomain(&o) }
	return pagination.BuildResult(responses, p.GetLimit(), func(r dto.OrderResponse) string {
		return pagination.EncodeCursor(r.ID, r.CreatedAt)
	}), nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*dto.OrderResponse, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	resp := dto.FromDomain(o); return &resp, nil
}

func (s *Service) Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error) {
	variant, err := s.repo.FindVariant(ctx, req.VariantID)
	if err != nil { return nil, err }
	order := &domain.Order{
		VariantID: req.VariantID, CustomerName: req.CustomerName,
		TelegramUsername: req.TelegramUsername, Phone: req.Phone, Comment: req.Comment,
		Status: domain.OrderNew,
	}
	if err = s.repo.Create(ctx, order); err != nil { return nil, err }

	// Async Telegram notification — silent fail
	s.tg.SendOrderNotification(order, variant, s.siteBase)
	go func() { _ = s.repo.SetTgNotified(context.Background(), order.ID) }()

	order.Variant = *variant
	resp := dto.FromDomain(order); return &resp, nil
}

func (s *Service) UpdateStatus(ctx context.Context, id string, req dto.UpdateStatusRequest) (*dto.OrderResponse, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil { return nil, err }
	if !req.Status.IsValid() { return nil, apperror.Validation("invalid order status") }
	if !o.CanTransitionTo(req.Status) {
		return nil, apperror.Newf("BAD_TRANSITION", 422, "cannot transition order from %s to %s", o.Status, req.Status)
	}
	if err = s.repo.UpdateStatus(ctx, id, req.Status); err != nil { return nil, err }
	o.Status = req.Status
	resp := dto.FromDomain(o); return &resp, nil
}
