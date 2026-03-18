package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"github.com/akza/akza-api/internal/config"
	coreapp "github.com/akza/akza-api/internal/core/app"
	"github.com/akza/akza-api/internal/core/module"
	"github.com/akza/akza-api/internal/logger"
	"github.com/akza/akza-api/internal/modules/auth"
	"github.com/akza/akza-api/internal/modules/collections"
	"github.com/akza/akza-api/internal/modules/media"
	"github.com/akza/akza-api/internal/modules/orders"
	"github.com/akza/akza-api/internal/modules/products"
	"github.com/akza/akza-api/internal/modules/site"
	"github.com/akza/akza-api/internal/modules/variants"
	"go.uber.org/zap"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	// Logger
	env := os.Getenv(cfg.Logger.Level)
	if env == "" {
		env = "development"
	}
	log, err := logger.New(env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync() //nolint:errcheck

	// Run DB migrations via goose
	if err = runMigrations(cfg.Database.URL, log); err != nil {
		log.Fatal("migrations failed", zap.Error(err))
	}

	// Register all modules
	modules := []module.Module{
		auth.New(),
		collections.New(),
		products.New(),
		variants.New(),
		orders.New(),
		media.New(),
		site.New(),
	}

	// Bootstrap application
	app, err := coreapp.New(cfg, log, modules)
	if err != nil {
		log.Fatal("application init failed", zap.Error(err))
	}

	if err = app.Run(); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

func runMigrations(dsn string, log *zap.Logger) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("ping db for migrations: %w", err)
	}

	goose.SetLogger(goose.NopLogger())
	if err = goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err = goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	log.Info("migrations applied successfully")
	return nil
}
