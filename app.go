package main

import (
	"context"
	"fmt"
	"log/slog"

	"personal-digital-inventory/internal/infrastructure/config"
	"personal-digital-inventory/internal/infrastructure/database"
	"personal-digital-inventory/internal/infrastructure/logging"

	"gorm.io/gorm"
)

type App struct {
	ctx context.Context
	db  *gorm.DB
	cfg *config.Config
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Warn("Failed to load config, starting with limited functionality", "error", err)
		return
	}
	a.cfg = cfg

	if err := logging.Setup(cfg.Logging.Level); err != nil {
		slog.Warn("Failed to configure logging", "error", err)
	}

	dbCfg := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		slog.Warn("Database connection failed, starting without database connectivity", "error", err)
		return
	}

	a.db = db
	slog.Info("Database connected successfully")

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get underlying sql.DB for migrations", "error", err)
		return
	}

	if err := database.RunMigrations(sqlDB); err != nil {
		slog.Error("Failed to run database migrations", "error", err)
		return
	}
	slog.Info("Database migrations completed")
}

func (a *App) shutdown(ctx context.Context) {
	if a.db != nil {
		if err := database.Close(a.db); err != nil {
			slog.Error("Failed to close database", "error", err)
		}
	}
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
