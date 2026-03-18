package handler

import (
	"context"
	"github.com/akza/akza-api/internal/modules/media/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	List(ctx context.Context, mediaType string, p pagination.CursorPage) (pagination.PageResult[dto.MediaResponse], error)
	Presign(ctx context.Context, adminID string, req dto.PresignRequest) (*dto.PresignResponse, error)
	Confirm(ctx context.Context, adminID string, req dto.ConfirmRequest) (*dto.MediaResponse, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }
func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

func (h *Handler) List(c *gin.Context) {
	var p pagination.CursorPage; _ = c.ShouldBindQuery(&p)
	result, err := h.svc.List(c.Request.Context(), c.Query("type"), p)
	if err != nil { middleware.Err(c, err); return }
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

func (h *Handler) Presign(c *gin.Context) {
	var req dto.PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	resp, err := h.svc.Presign(c.Request.Context(), c.GetString(middleware.AdminIDKey), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, resp)
}

func (h *Handler) Confirm(c *gin.Context) {
	var req dto.ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	asset, err := h.svc.Confirm(c.Request.Context(), c.GetString(middleware.AdminIDKey), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, asset)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
