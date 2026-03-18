package handler

import (
	"context"
	"github.com/akza/akza-api/internal/modules/variants/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	ListByProduct(ctx context.Context, productID string, onlyPublished bool) ([]dto.VariantResponse, error)
	ListByProductSlug(ctx context.Context, productSlug string, onlyPublished bool) ([]dto.VariantResponse, error)
	GetBySlug(ctx context.Context, slug string) (*dto.VariantResponse, error)
	GetByIDAdmin(ctx context.Context, id string) (*dto.VariantResponse, error)
	Create(ctx context.Context, req dto.CreateVariantRequest) (*dto.VariantResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateVariantRequest) (*dto.VariantResponse, error)
	SetPublished(ctx context.Context, id string, pub bool) error
	Delete(ctx context.Context, id string) error
	PresignImage(ctx context.Context, id, filename, ct string) (*dto.PresignResponse, error)
	ConfirmImage(ctx context.Context, variantID string, req dto.ConfirmImageRequest) (*dto.ImageResponse, error)
	DeleteImage(ctx context.Context, variantID, imageID string) error
	ReorderImages(ctx context.Context, variantID string, req dto.ReorderImagesRequest) error
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }
func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

// ListByProduct is called from the public route /products/:slug/variants
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
	v, err := h.svc.GetByIDAdmin(c.Request.Context(), c.Param("id"))
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
	var req dto.UpdateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	v, err := h.svc.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, v)
}
func (h *Handler) SetPublished(c *gin.Context) {
	var body struct{ Published bool `json:"published"` }
	if err := c.ShouldBindJSON(&body); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.SetPublished(c.Request.Context(), c.Param("id"), body.Published); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
func (h *Handler) PresignImage(c *gin.Context) {
	var req dto.PresignImageRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	resp, err := h.svc.PresignImage(c.Request.Context(), c.Param("id"), req.Filename, req.ContentType)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, resp)
}
func (h *Handler) ConfirmImage(c *gin.Context) {
	var req dto.ConfirmImageRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	img, err := h.svc.ConfirmImage(c.Request.Context(), c.Param("id"), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, img)
}
func (h *Handler) DeleteImage(c *gin.Context) {
	if err := h.svc.DeleteImage(c.Request.Context(), c.Param("id"), c.Param("imageID")); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
func (h *Handler) ReorderImages(c *gin.Context) {
	var req dto.ReorderImagesRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.ReorderImages(c.Request.Context(), c.Param("id"), req); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
