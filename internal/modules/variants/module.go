package variants

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/variants/handler"
	"github.com/akza/akza-api/internal/modules/variants/repository"
	"github.com/akza/akza-api/internal/modules/variants/service"
	"github.com/gin-gonic/gin"
)

type Module struct { module.Base; handler *handler.Handler }
func New() *Module { return &Module{} }
func (m *Module) GetName() string { return "variants" }
func (m *Module) Init(deps *module.Deps) error {
	m.handler = handler.New(service.New(repository.New(deps.DB), deps.S3))
	return nil
}
func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/variants/:slug", m.handler.GetBySlug)
	public.GET("/products/:slug/variants", m.handler.ListByProduct)

	v := admin.Group("/variants")
	v.POST("", m.handler.Create)
	v.GET("/:id", m.handler.GetByID)
	v.PUT("/:id", m.handler.Update)
	v.PATCH("/:id/publish", m.handler.SetPublished)
	v.DELETE("/:id", m.handler.Delete)
	v.POST("/:id/images/presign", m.handler.PresignImage)
	v.POST("/:id/images/confirm", m.handler.ConfirmImage)
	v.DELETE("/:id/images/:imageID", m.handler.DeleteImage)
	v.PATCH("/:id/images/reorder", m.handler.ReorderImages)
}
