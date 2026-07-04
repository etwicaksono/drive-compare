package main

import (
	"context"
	"fmt"
	"log/slog"

	"personal-digital-inventory/internal/infrastructure/config"
	"personal-digital-inventory/internal/infrastructure/database"
	"personal-digital-inventory/internal/infrastructure/gdrive"
	"personal-digital-inventory/internal/infrastructure/logging"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

type App struct {
	ctx  context.Context
	db   *gorm.DB
	cfg  *config.Config
	auth *gdrive.Auth
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

	a.auth = gdrive.NewAuth(gdrive.AuthConfig{
		ClientID:     cfg.GoogleDrive.ClientID,
		ClientSecret: cfg.GoogleDrive.ClientSecret,
		TokenFile:    cfg.GoogleDrive.TokenFile,
	})

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

func (a *App) SelectLocalDirectory() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Local Directory",
	})
	if err != nil {
		return "", fmt.Errorf("failed to open directory dialog: %w", err)
	}
	return dir, nil
}

func (a *App) GetDriveConnectionStatus() bool {
	if a.auth == nil {
		return false
	}
	return a.auth.IsConnected()
}

func (a *App) ConnectGoogleDrive() (string, error) {
	if a.auth == nil {
		return "", fmt.Errorf("Google Drive not configured")
	}
	return a.auth.GetAuthURL(), nil
}

func (a *App) ExchangeGoogleDriveCode(code string) error {
	if a.auth == nil {
		return fmt.Errorf("Google Drive not configured")
	}
	return a.auth.Exchange(a.ctx, code)
}

func (a *App) DisconnectGoogleDrive() error {
	if a.auth == nil {
		return fmt.Errorf("Google Drive not configured")
	}
	return a.auth.Disconnect()
}

func (a *App) GetIgnorePatterns() []string {
	if a.cfg == nil {
		return nil
	}
	return a.cfg.Ignore
}

func (a *App) AddIgnorePattern(pattern string) error {
	if a.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	for _, p := range a.cfg.Ignore {
		if p == pattern {
			return fmt.Errorf("pattern %q already exists", pattern)
		}
	}
	a.cfg.Ignore = append(a.cfg.Ignore, pattern)
	return nil
}

func (a *App) RemoveIgnorePattern(pattern string) error {
	if a.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	for i, p := range a.cfg.Ignore {
		if p == pattern {
			a.cfg.Ignore = append(a.cfg.Ignore[:i], a.cfg.Ignore[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("pattern %q not found", pattern)
}
