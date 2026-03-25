package products

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/products/handler"
	"github.com/akza/akza-api/internal/modules/products/repository"
	"github.com/akza/akza-api/internal/modules/products/service"
	"github.com/gin-gonic/gin"
)

type Module struct {
	module.Base
	handler *handler.Handler
}

func New() *Module { return &Module{} }
func (m *Module) GetName() string { return "products" }

func (m *Module) Init(deps *module.Deps) error {
	repo := repository.New(deps.DB)
	svc := service.New(repo, deps.S3)
	m.handler = handler.New(svc)
	return nil
}

func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/collections/:slug/products", m.handler.ListByCollection)
	public.GET("/products/:slug", m.handler.GetBySlug)

	p := admin.Group("/products")
	// IMPORTANT: static routes before wildcard to avoid conflicts
	p.GET("", m.handler.ListByCollectionAdmin)   // GET /admin/products?collection_id=UUID
	p.POST("", m.handler.Create)
	p.PATCH("/reorder", m.handler.Reorder)
	p.GET("/:id", m.handler.GetByID)
	p.PUT("/:id", m.handler.Update)
	p.PATCH("/:id/publish", m.handler.SetPublished)
	p.DELETE("/:id", m.handler.Delete)
	p.POST("/:id/cover/presign", m.handler.PresignCover)
	p.POST("/:id/cover/confirm", m.handler.ConfirmCover)
	p.DELETE("/:id/cover", m.handler.DeleteCover)
}
