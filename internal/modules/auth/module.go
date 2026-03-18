package auth

import (
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/modules/auth/handler"
	"github.com/akza/akza-api/internal/modules/auth/repository"
	"github.com/akza/akza-api/internal/modules/auth/service"
	jwtpkg "github.com/akza/akza-api/internal/pkg/jwt"
	"github.com/gin-gonic/gin"
)

type Module struct {
	module.Base
	handler *handler.Handler
}

func New() *Module { return &Module{} }

func (m *Module) GetName() string { return "auth" }

func (m *Module) Init(deps *module.Deps) error {
	jwtManager := jwtpkg.NewManager(deps.Config.JWT.Secret, deps.Config.JWT.ExpiresHours)
	repo := repository.New(deps.DB)
	svc := service.New(repo, jwtManager)
	m.handler = handler.New(svc)
	return nil
}

func (m *Module) RegisterRoutes(public, admin *gin.RouterGroup) {
	auth := public.Group("/auth")
	auth.POST("/login", m.handler.Login)

	admin.GET("/auth/me", m.handler.Me)
}
