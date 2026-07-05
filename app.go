package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"personal-digital-inventory/internal/domain"
	"personal-digital-inventory/internal/infrastructure/config"
	"personal-digital-inventory/internal/infrastructure/database"
	"personal-digital-inventory/internal/infrastructure/gdrive"
	"personal-digital-inventory/internal/infrastructure/logging"
	"personal-digital-inventory/internal/infrastructure/scanner"

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

	if cfg.GoogleDrive.ClientID == "" || cfg.GoogleDrive.ClientSecret == "" {
		slog.Warn("Google Drive credentials not configured. Set PDI_GOOGLE_CLIENT_ID and PDI_GOOGLE_CLIENT_SECRET environment variables, or fill google_drive.client_id and google_drive.client_secret in config.yaml")
	} else {
		a.auth = gdrive.NewAuth(gdrive.AuthConfig{
			ClientID:     cfg.GoogleDrive.ClientID,
			ClientSecret: cfg.GoogleDrive.ClientSecret,
			TokenFile:    cfg.GoogleDrive.TokenFile,
		})
	}

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
		return "", fmt.Errorf("Google Drive not configured. Set PDI_GOOGLE_CLIENT_ID and PDI_GOOGLE_CLIENT_SECRET environment variables, or fill them in config.yaml")
	}
	return a.auth.ConnectWithCallback(a.ctx)
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
	return a.cfg.Ignore.Patterns
}

func (a *App) AddIgnorePattern(pattern string) error {
	if a.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}
	for _, p := range a.cfg.Ignore.Patterns {
		if p == pattern {
			return fmt.Errorf("pattern %q already exists", pattern)
		}
	}
	a.cfg.Ignore.Patterns = append(a.cfg.Ignore.Patterns, pattern)
	return nil
}

func (a *App) RemoveIgnorePattern(pattern string) error {
	if a.cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	for i, p := range a.cfg.Ignore.Patterns {
		if p == pattern {
			a.cfg.Ignore.Patterns = append(a.cfg.Ignore.Patterns[:i], a.cfg.Ignore.Patterns[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("pattern %q not found", pattern)
}

func (a *App) GetRespectGitignore() bool {
	if a.cfg == nil || a.cfg.Ignore.RespectGitignore == nil {
		return true
	}
	return *a.cfg.Ignore.RespectGitignore
}

func (a *App) SetRespectGitignore(value bool) {
	if a.cfg != nil {
		a.cfg.Ignore.RespectGitignore = &value
	}
}

type DriveFolder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *App) ListDriveFolders(parentID string) ([]DriveFolder, error) {
	if a.auth == nil {
		return nil, fmt.Errorf("Google Drive not configured")
	}
	if !a.auth.IsConnected() {
		return nil, fmt.Errorf("Google Drive not connected. Please connect in Settings first.")
	}

	client, err := a.auth.GetClient(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Drive client: %w", err)
	}

	api, err := scanner.NewRealDriveAPI(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive API: %w", err)
	}

	if parentID == "" {
		parentID = "root"
	}

	query := fmt.Sprintf("'%s' in parents and mimeType = 'application/vnd.google-apps.folder' and trashed = false", parentID)
	var folders []DriveFolder
	var pageToken string

	for {
		files, nextToken, err := api.ListFiles(a.ctx, query, "", pageToken)
		if err != nil {
			return nil, fmt.Errorf("failed to list Drive folders: %w", err)
		}

		for _, f := range files {
			folders = append(folders, DriveFolder{ID: f.Id, Name: f.Name})
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	return folders, nil
}

type CompareResultItem struct {
	RelativePath string `json:"relativePath"`
	Name         string `json:"name"`
	Extension    string `json:"extension"`
	Size         int64  `json:"size"`
}

type CompareResponse struct {
	LocalOnly []CompareResultItem `json:"localOnly"`
	DriveOnly []CompareResultItem `json:"driveOnly"`
	Both      []CompareResultItem `json:"both"`
}

func (a *App) RunComparison(localPath string, driveFolderID string) (*CompareResponse, error) {
	if localPath == "" {
		return nil, fmt.Errorf("local directory is required")
	}
	if driveFolderID == "" {
		return nil, fmt.Errorf("Google Drive folder is required")
	}
	if a.auth == nil {
		return nil, fmt.Errorf("Google Drive not configured")
	}
	if !a.auth.IsConnected() {
		return nil, fmt.Errorf("Google Drive not connected. Please connect in Settings first.")
	}

	slog.Info("Starting comparison", "localPath", localPath, "driveFolderID", driveFolderID)

	localScanner := scanner.NewLocalScanner(a.cfg.Ignore.Patterns, *a.cfg.Ignore.RespectGitignore)
	localFiles, err := localScanner.Scan(a.ctx, localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to scan local directory: %w", err)
	}
	slog.Info("Local scan complete", "files", len(localFiles))

	client, err := a.auth.GetClient(a.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Drive client: %w", err)
	}

	api, err := scanner.NewRealDriveAPI(client)
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive API: %w", err)
	}

	driveScanner := scanner.NewDriveScanner(api, a.cfg.Ignore.Patterns)
	driveFiles, err := driveScanner.Scan(a.ctx, driveFolderID)
	if err != nil {
		return nil, fmt.Errorf("failed to scan Google Drive: %w", err)
	}
	slog.Info("Drive scan complete", "files", len(driveFiles))

	result := domain.CompareFiles(localFiles, driveFiles)

	resp := &CompareResponse{
		LocalOnly: toResultItems(result.LocalOnly),
		DriveOnly: toResultItems(result.DriveOnly),
		Both:      toResultItems(result.Both),
	}

	slog.Info("Comparison complete",
		"localOnly", len(resp.LocalOnly),
		"driveOnly", len(resp.DriveOnly),
		"both", len(resp.Both),
	)

	return resp, nil
}

func toResultItems(items []domain.FileItem) []CompareResultItem {
	result := make([]CompareResultItem, len(items))
	for i, item := range items {
		result[i] = CompareResultItem{
			RelativePath: item.RelativePath,
			Name:         item.Name,
			Extension:    item.Extension,
			Size:         item.Size,
		}
	}
	return result
}

func (a *App) ExportToClipboard(text string) error {
	if text == "" {
		return fmt.Errorf("nothing to export")
	}
	return runtime.ClipboardSetText(a.ctx, text)
}

func (a *App) ExportToFile(content string) (string, error) {
	if content == "" {
		return "", fmt.Errorf("nothing to export")
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export file list",
		DefaultFilename: "export.txt",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text Files", Pattern: "*.txt"},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to open save dialog: %w", err)
	}
	if path == "" {
		return "", nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return path, nil
}
