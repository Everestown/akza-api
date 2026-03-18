package handler

import (
	"context"
	"github.com/akza/akza-api/internal/modules/orders/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	List(ctx context.Context, q dto.ListOrdersQuery) (pagination.PageResult[dto.OrderResponse], error)
	GetByID(ctx context.Context, id string) (*dto.OrderResponse, error)
	Create(ctx context.Context, req dto.CreateOrderRequest) (*dto.OrderResponse, error)
	UpdateStatus(ctx context.Context, id string, req dto.UpdateStatusRequest) (*dto.OrderResponse, error)
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }
func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

func (h *Handler) List(c *gin.Context) {
	var q dto.ListOrdersQuery; _ = c.ShouldBindQuery(&q)
	result, err := h.svc.List(c.Request.Context(), q)
	if err != nil { middleware.Err(c, err); return }
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

func (h *Handler) GetByID(c *gin.Context) {
	o, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, o)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	o, err := h.svc.Create(c.Request.Context(), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, o)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	o, err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, o)
}
