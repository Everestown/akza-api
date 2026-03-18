package app

import (
	"fmt"

	"github.com/akza/akza-api/internal/core/module"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ModuleRegistry manages the lifecycle of all feature modules.
type ModuleRegistry struct {
	modules []module.Module
	log     *zap.Logger
}

func NewModuleRegistry(log *zap.Logger) *ModuleRegistry {
	return &ModuleRegistry{log: log}
}

// Register adds a module to the registry.
func (r *ModuleRegistry) Register(m module.Module) {
	r.modules = append(r.modules, m)
}

// InitAll calls Init on every registered module.
func (r *ModuleRegistry) InitAll(deps *module.Deps) error {
	for _, m := range r.modules {
		r.log.Info("initializing module", zap.String("name", m.GetName()))
		if err := m.Init(deps); err != nil {
			return fmt.Errorf("module %q init: %w", m.GetName(), err)
		}
	}
	return nil
}

// MountAll calls RegisterRoutes on every module.
func (r *ModuleRegistry) MountAll(public, admin *gin.RouterGroup) {
	for _, m := range r.modules {
		m.RegisterRoutes(public, admin)
	}
}

// CloseAll calls Close on every module in reverse order.
func (r *ModuleRegistry) CloseAll() {
	for i := len(r.modules) - 1; i >= 0; i-- {
		if err := r.modules[i].Close(); err != nil {
			r.log.Error("module close error",
				zap.String("name", r.modules[i].GetName()),
				zap.Error(err),
			)
		}
	}
}
