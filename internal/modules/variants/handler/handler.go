package handler

import (
	"context"
	"fmt"

	"github.com/akza/akza-api/internal/modules/variants/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/httputil"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	ListByProduct(ctx context.Context, productID int64, onlyPublished bool) ([]dto.VariantResponse, error)
	ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]dto.VariantResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.VariantResponse, error)
	GetByIDAdmin(ctx context.Context, id int64) (*dto.VariantResponse, error)
	Create(ctx context.Context, req dto.CreateVariantRequest) (*dto.VariantResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateVariantRequest) (*dto.VariantResponse, error)
	SetPublished(ctx context.Context, id int64, pub bool) error
	Delete(ctx context.Context, id int64) error
	PresignImage(ctx context.Context, id int64, filename, ct string) (*dto.PresignResponse, error)
	ConfirmImage(ctx context.Context, variantID int64, req dto.ConfirmImageRequest) (*dto.ImageResponse, error)
	DeleteImage(ctx context.Context, variantID, imageID int64) error
	ReorderImages(ctx context.Context, variantID int64, req dto.ReorderImagesRequest) error
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }
func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

func (h *Handler) ListByProduct(c *gin.Context) {
	items, err := h.svc.ListByProductSlug(c.Request.Context(), c.Param("slug"), true)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, items)
}

func (h *Handler) GetBySlug(c *gin.Context) {
	v, err := h.svc.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, v)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	v, err := h.svc.GetByIDAdmin(c.Request.Context(), id)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, v)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	v, err := h.svc.Create(c.Request.Context(), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, v)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	var req dto.UpdateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	v, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, v)
}

func (h *Handler) SetPublished(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	var body struct{ Published bool `json:"published"` }
	if err := c.ShouldBindJSON(&body); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.SetPublished(c.Request.Context(), id, body.Published); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	if err := h.svc.Delete(c.Request.Context(), id); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}

func (h *Handler) PresignImage(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	var req struct {
		Filename    string `json:"filename"     binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	resp, err := h.svc.PresignImage(c.Request.Context(), id, req.Filename, req.ContentType)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, resp)
}

func (h *Handler) ConfirmImage(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	var req dto.ConfirmImageRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	img, err := h.svc.ConfirmImage(c.Request.Context(), id, req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, img)
}

func (h *Handler) DeleteImage(c *gin.Context) {
	variantID, ok := httputil.ParseID(c); if !ok { return }
	imageID, ok := httputil.ParseParam(c, "imageID"); if !ok { return }
	if err := h.svc.DeleteImage(c.Request.Context(), variantID, imageID); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}

func (h *Handler) ReorderImages(c *gin.Context) {
	id, ok := httputil.ParseID(c); if !ok { return }
	var req dto.ReorderImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.ReorderImages(c.Request.Context(), id, req); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}

// ListByProductAdmin — admin: GET /admin/variants?product_id=123
func (h *Handler) ListByProductAdmin(c *gin.Context) {
	v := c.Query("product_id")
	if v == "" { middleware.Err(c, apperror.Validation("product_id query param required")); return }
	var pid int64
	if _, err := fmt.Sscanf(v, "%d", &pid); err != nil {
		middleware.Err(c, apperror.Validation("product_id must be integer")); return
	}
	items, err := h.svc.ListByProduct(c.Request.Context(), pid, false)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, items)
}
