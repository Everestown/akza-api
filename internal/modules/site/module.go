package site

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/site/handler"
	"github.com/akza/akza-api/internal/modules/site/repository"
	"github.com/akza/akza-api/internal/modules/site/service"
	"github.com/gin-gonic/gin"
)

type Module struct { module.Base; handler *handler.Handler }
func New() *Module { return &Module{} }
func (m *Module) GetName() string { return "site" }
func (m *Module) Init(deps *module.Deps) error {
	m.handler = handler.New(service.New(repository.New(deps.DB)))
	return nil
}
func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	public.GET("/site/content/:section", m.handler.GetBySection)
	s := admin.Group("/site")
	s.GET("/content", m.handler.GetAll)
	s.PUT("/content/:section", m.handler.UpdateSection)
}
