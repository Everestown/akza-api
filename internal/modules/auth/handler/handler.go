package handler

import (
	"context"

	"github.com/akza/akza-api/internal/modules/auth/dto"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
)

type svc interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Me(ctx context.Context, adminID string) (*dto.AdminInfo, error)
}

type Handler struct{ svc svc }

func New(svc svc) *Handler { return &Handler{svc: svc} }

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, errValidation(err))
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, resp)
}

func (h *Handler) Me(c *gin.Context) {
	adminID := c.GetString(middleware.AdminIDKey)
	info, err := h.svc.Me(c.Request.Context(), adminID)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, info)
}
