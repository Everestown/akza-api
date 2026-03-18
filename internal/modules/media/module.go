package media

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/media/handler"
	"github.com/akza/akza-api/internal/modules/media/repository"
	"github.com/akza/akza-api/internal/modules/media/service"
	"github.com/gin-gonic/gin"
)

type Module struct { module.Base; handler *handler.Handler }
func New() *Module { return &Module{} }
func (m *Module) GetName() string { return "media" }
func (m *Module) Init(deps *module.Deps) error {
	m.handler = handler.New(service.New(repository.New(deps.DB), deps.S3))
	return nil
}
func (m *Module) RegisterRoutes(_ *gin.RouterGroup, admin *gin.RouterGroup) {
	med := admin.Group("/media")
	med.POST("/presign", m.handler.Presign)
	med.POST("/confirm", m.handler.Confirm)
	med.GET("", m.handler.List)
	med.DELETE("/:id", m.handler.Delete)
}
