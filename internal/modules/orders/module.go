package orders

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/orders/handler"
	"github.com/akza/akza-api/internal/modules/orders/repository"
	"github.com/akza/akza-api/internal/modules/orders/service"
	"github.com/gin-gonic/gin"
)

type Module struct { module.Base; handler *handler.Handler }
func New() *Module { return &Module{} }
func (m *Module) GetName() string { return "orders" }
func (m *Module) Init(deps *module.Deps) error {
	siteBase := "https://akza.ru"
	if deps.Config.CORS.AllowedOrigins != nil && len(deps.Config.CORS.AllowedOrigins) > 0 {
		siteBase = deps.Config.CORS.AllowedOrigins[len(deps.Config.CORS.AllowedOrigins)-1]
	}
	m.handler = handler.New(service.New(repository.New(deps.DB), deps.Telegram, siteBase))
	return nil
}
func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.POST("/orders", m.handler.Create)
	o := admin.Group("/orders")
	o.GET("", m.handler.List)
	o.GET("/:id", m.handler.GetByID)
	o.GET("/stats", m.handler.Stats)
	o.PATCH("/:id/status", m.handler.UpdateStatus)
}
