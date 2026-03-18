package module

import (
	"github.com/akza/akza-api/internal/config"
	"github.com/akza/akza-api/internal/pkg/storage"
	"github.com/akza/akza-api/internal/pkg/telegram"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Deps contains all shared dependencies injected into each module.
type Deps struct {
	DB       *gorm.DB
	Config   *config.Config
	Logger   *zap.Logger
	S3       *storage.Client
	Telegram *telegram.Bot
}

// Module is the contract every feature module must implement.
type Module interface {
	GetName() string
	Init(deps *Deps) error
	RegisterRoutes(public, admin *gin.RouterGroup)
	Close() error
}
