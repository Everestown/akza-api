package handler

import (
	"context"
	"github.com/akza/akza-api/internal/modules/site/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	GetAll(ctx context.Context) ([]dto.SitePageResponse, error)
	GetBySection(ctx context.Context, section string) (*dto.SitePageResponse, error)
	UpdateSection(ctx context.Context, section, adminID string, req dto.UpdateContentRequest) (*dto.SitePageResponse, error)
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }
func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

func (h *Handler) GetBySection(c *gin.Context) {
	p, err := h.svc.GetBySection(c.Request.Context(), c.Param("section"))
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, p)
}

func (h *Handler) GetAll(c *gin.Context) {
	pages, err := h.svc.GetAll(c.Request.Context())
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, pages)
}

func (h *Handler) UpdateSection(c *gin.Context) {
	var req dto.UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	p, err := h.svc.UpdateSection(c.Request.Context(), c.Param("section"), c.GetString(middleware.AdminIDKey), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, p)
}
