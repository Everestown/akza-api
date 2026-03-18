package collections

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/collections/handler"
	"github.com/akza/akza-api/internal/modules/collections/repository"
	"github.com/akza/akza-api/internal/modules/collections/service"
	"github.com/gin-gonic/gin"
)

type Module struct {
	module.Base
	handler *handler.Handler
}

func New() *Module { return &Module{} }

func (m *Module) GetName() string { return "collections" }

func (m *Module) Init(deps *module.Deps) error {
	repo := repository.New(deps.DB)
	svc := service.New(repo, deps.S3)
	m.handler = handler.New(svc)
	return nil
}

func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	// Public
	public.GET("/collections", m.handler.ListPublic)
	public.GET("/collections/:slug", m.handler.GetBySlug)

	// Admin
	c := admin.Group("/collections")
	c.GET("", m.handler.ListAll)
	c.POST("", m.handler.Create)
	c.GET("/:id", m.handler.GetByID)
	c.PUT("/:id", m.handler.Update)
	c.PATCH("/:id/status", m.handler.UpdateStatus)
	c.DELETE("/:id", m.handler.Delete)
	c.PATCH("/reorder", m.handler.Reorder)
	c.POST("/:id/cover/presign", m.handler.PresignCover)
	c.POST("/:id/cover/confirm", m.handler.ConfirmCover)
}
