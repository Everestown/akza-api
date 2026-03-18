package handler

import (
	"context"
	"github.com/akza/akza-api/internal/modules/products/dto"
	"github.com/akza/akza-api/internal/pkg/apperror"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type svc interface {
	ListByCollection(ctx context.Context, collectionID string, onlyPublished bool, p pagination.CursorPage) (pagination.PageResult[dto.ProductResponse], error)
	ListByCollectionSlug(ctx context.Context, collectionSlug string, onlyPublished bool, p pagination.CursorPage) (pagination.PageResult[dto.ProductResponse], error)
	GetBySlug(ctx context.Context, slug string) (*dto.ProductResponse, error)
	GetByIDAdmin(ctx context.Context, id string) (*dto.ProductResponse, error)
	Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	Update(ctx context.Context, id string, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	SetPublished(ctx context.Context, id string, published bool) error
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, req dto.ReorderRequest) error
	PresignCover(ctx context.Context, id, filename, ct string) (*dto.PresignResponse, error)
	ConfirmCover(ctx context.Context, id, s3Key string) error
}

type Handler struct{ svc svc }
func New(svc svc) *Handler { return &Handler{svc: svc} }

func ve(err error) error {
	if v, ok := err.(validator.ValidationErrors); ok { return apperror.Validation(v.Error()) }
	return apperror.Validation(err.Error())
}

// ListByCollection — public: /collections/:slug/products
func (h *Handler) ListByCollection(c *gin.Context) {
	var p pagination.CursorPage; _ = c.ShouldBindQuery(&p)
	result, err := h.svc.ListByCollectionSlug(c.Request.Context(), c.Param("slug"), true, p)
	if err != nil { middleware.Err(c, err); return }
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

// ListByCollectionAdmin — admin: GET /admin/products?collection_id=UUID
func (h *Handler) ListByCollectionAdmin(c *gin.Context) {
	collectionID := c.Query("collection_id")
	if collectionID == "" {
		middleware.Err(c, apperror.Validation("collection_id query param required"))
		return
	}
	var p pagination.CursorPage; _ = c.ShouldBindQuery(&p)
	result, err := h.svc.ListByCollection(c.Request.Context(), collectionID, false, p)
	if err != nil { middleware.Err(c, err); return }
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

func (h *Handler) GetBySlug(c *gin.Context) {
	p, err := h.svc.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, p)
}

func (h *Handler) GetByID(c *gin.Context) {
	p, err := h.svc.GetByIDAdmin(c.Request.Context(), c.Param("id"))
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, p)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	p, err := h.svc.Create(c.Request.Context(), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.Created(c, p)
}

func (h *Handler) Update(c *gin.Context) {
	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	p, err := h.svc.Update(c.Request.Context(), c.Param("id"), req)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, p)
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

func (h *Handler) Reorder(c *gin.Context) {
	var req dto.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.Reorder(c.Request.Context(), req); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}

func (h *Handler) PresignCover(c *gin.Context) {
	var req dto.PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil { middleware.Err(c, ve(err)); return }
	resp, err := h.svc.PresignCover(c.Request.Context(), c.Param("id"), req.Filename, req.ContentType)
	if err != nil { middleware.Err(c, err); return }
	middleware.OK(c, resp)
}

func (h *Handler) ConfirmCover(c *gin.Context) {
	var body struct { S3Key string `json:"s3_key" binding:"required"` }
	if err := c.ShouldBindJSON(&body); err != nil { middleware.Err(c, ve(err)); return }
	if err := h.svc.ConfirmCover(c.Request.Context(), c.Param("id"), body.S3Key); err != nil { middleware.Err(c, err); return }
	middleware.NoContent(c)
}
