package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/akza/akza-api/internal/config"
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/pkg/database"
	jwtpkg "github.com/akza/akza-api/internal/pkg/jwt"
	"github.com/akza/akza-api/internal/pkg/middleware"
	"github.com/akza/akza-api/internal/pkg/storage"
	"github.com/akza/akza-api/internal/pkg/telegram"
	akzavalidator "github.com/akza/akza-api/internal/pkg/validator"
	"github.com/akza/akza-api/docs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App is the application root that wires all dependencies.
type App struct {
	cfg      *config.Config
	log      *zap.Logger
	db       *gorm.DB
	router   *gin.Engine
	registry *ModuleRegistry
	server   *http.Server
}

// New initialises the application: database, S3, Telegram, router, modules.
func New(cfg *config.Config, log *zap.Logger, modules []module.Module) (*App, error) {
	// Database
	db, err := database.Connect(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	log.Info("database connected")

	// S3 (optional – skip if credentials are empty)
	var s3Client *storage.Client
	if cfg.S3.AccessKey != "" {
		s3Client, err = storage.New(cfg.S3)
		if err != nil {
			return nil, fmt.Errorf("s3: %w", err)
		}
		log.Info("s3 connected", zap.String("bucket", cfg.S3.Bucket))
	} else {
		log.Warn("s3 not configured – presign endpoints will be unavailable")
	}

	// Telegram bot
	tgBot := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.AdminChatID, log)
	if tgBot.IsConfigured() {
		log.Info("telegram bot configured")
	} else {
		log.Warn("telegram bot not configured – notifications disabled")
	}

	// JWT
	jwtManager := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.ExpiresHours)

	// Gin router
	if cfg.Server.Address != "" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.ZapLogger(log))

	// Register custom validation tags (slug, tg_username)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		akzavalidator.RegisterCustomValidators(v)
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "akza-api"})
	})

	// Swagger UI — accessible at /swagger/index.html
	docs.RegisterSwagger(router)

	// Route groups
	v1 := router.Group("/api/v1")
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(jwtManager))

	// Module registry
	registry := NewModuleRegistry(log)
	for _, m := range modules {
		registry.Register(m)
	}

	deps := &module.Deps{
		DB:       db,
		Config:   cfg,
		Logger:   log,
		S3:       s3Client,
		Telegram: tgBot,
	}
	if err = registry.InitAll(deps); err != nil {
		return nil, fmt.Errorf("modules init: %w", err)
	}
	registry.MountAll(v1, adminGroup)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{cfg: cfg, log: log, db: db, router: router, registry: registry, server: srv}, nil
}

// Run starts the HTTP server and blocks until OS signal or error.
func (a *App) Run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		a.log.Info("server starting", zap.String("addr", a.cfg.Server.Address))
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		a.log.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.log.Error("graceful shutdown failed", zap.Error(err))
	}
	a.registry.CloseAll()

	sqlDB, _ := a.db.DB()
	_ = sqlDB.Close()

	a.log.Info("server stopped")
	return nil
}
