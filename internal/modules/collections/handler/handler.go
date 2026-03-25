package handler

import (
	"context"

	"github.com/akza/akza-api/internal/modules/collections/dto"
	"github.com/akza/akza-api/internal/pkg/httputil"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/akza/akza-api/internal/pkg/pagination"
	"github.com/gin-gonic/gin"
)

type svc interface {
	ListPublic(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error)
	ListAll(ctx context.Context, p pagination.CursorPage) (pagination.PageResult[dto.CollectionResponse], error)
	GetBySlug(ctx context.Context, slug string) (*dto.CollectionResponse, error)
	GetByIDAdmin(ctx context.Context, id int64) (*dto.CollectionResponse, error)
	Create(ctx context.Context, req dto.CreateCollectionRequest) (*dto.CollectionResponse, error)
	Update(ctx context.Context, id int64, req dto.UpdateCollectionRequest) (*dto.CollectionResponse, error)
	UpdateStatus(ctx context.Context, id int64, req dto.UpdateStatusRequest) (*dto.CollectionResponse, error)
	Delete(ctx context.Context, id int64) error
	Reorder(ctx context.Context, req dto.ReorderRequest) error
	PresignCover(ctx context.Context, id int64, filename, contentType string) (*dto.PresignResponse, error)
	ConfirmCover(ctx context.Context, id int64, s3Key string) error
	DeleteCover(ctx context.Context, id int64) error
}

type Handler struct{ svc svc }

func New(svc svc) *Handler { return &Handler{svc: svc} }

func (h *Handler) ListPublic(c *gin.Context) {
	var p pagination.CursorPage
	_ = c.ShouldBindQuery(&p)
	result, err := h.svc.ListPublic(c.Request.Context(), p)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

func (h *Handler) GetBySlug(c *gin.Context) {
	col, err := h.svc.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, col)
}

func (h *Handler) ListAll(c *gin.Context) {
	var p pagination.CursorPage
	_ = c.ShouldBindQuery(&p)
	result, err := h.svc.ListAll(c.Request.Context(), p)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.Paginated(c, result.Data, result.Cursor, result.HasMore, result.Limit)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	col, err := h.svc.GetByIDAdmin(c.Request.Context(), id)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, col)
}

func (h *Handler) Create(c *gin.Context) {
	var req dto.CreateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	col, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.Created(c, col)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var req dto.UpdateCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	col, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, col)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	col, err := h.svc.UpdateStatus(c.Request.Context(), id, req)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, col)
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.NoContent(c)
}

func (h *Handler) Reorder(c *gin.Context) {
	var req dto.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	if err := h.svc.Reorder(c.Request.Context(), req); err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.NoContent(c)
}

func (h *Handler) PresignCover(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var req dto.PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	resp, err := h.svc.PresignCover(c.Request.Context(), id, req.Filename, req.ContentType)
	if err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.OK(c, resp)
}

func (h *Handler) ConfirmCover(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var body struct {
		S3Key string `json:"s3_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		middleware.Err(c, validationErr(err))
		return
	}
	if err := h.svc.ConfirmCover(c.Request.Context(), id, body.S3Key); err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.NoContent(c)
}

func (h *Handler) DeleteCover(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteCover(c.Request.Context(), id); err != nil {
		middleware.Err(c, err)
		return
	}
	middleware.NoContent(c)
}
