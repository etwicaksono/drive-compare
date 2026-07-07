package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"drive-compare/internal/domain"
	"drive-compare/internal/infrastructure/config"
	"drive-compare/internal/infrastructure/gdrive"
	"drive-compare/internal/infrastructure/logging"
	"drive-compare/internal/infrastructure/scanner"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx  context.Context
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
		slog.Error("Failed to load config", "error", err)
		return
	}
	a.cfg = cfg
	slog.Info("Config loaded")

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
}

func (a *App) shutdown(ctx context.Context) {
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
	SourceAOnly []CompareResultItem `json:"sourceAOnly"`
	SourceBOnly []CompareResultItem `json:"sourceBOnly"`
	Both        []CompareResultItem `json:"both"`
}

func (a *App) scanSource(sourceType string, sourceID string) ([]domain.FileItem, error) {
	switch sourceType {
	case "local":
		if sourceID == "" {
			return nil, fmt.Errorf("local directory is required")
		}
		s := scanner.NewLocalScanner(a.cfg.Ignore.Patterns, *a.cfg.Ignore.RespectGitignore)
		return s.Scan(a.ctx, sourceID)
	case "drive":
		if sourceID == "" {
			return nil, fmt.Errorf("Google Drive folder is required")
		}
		if a.auth == nil || !a.auth.IsConnected() {
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
		s := scanner.NewDriveScanner(api, a.cfg.Ignore.Patterns)
		return s.Scan(a.ctx, sourceID)
	default:
		return nil, fmt.Errorf("invalid source type %q, must be \"local\" or \"drive\"", sourceType)
	}
}

func (a *App) RunComparison(sourceTypeA string, sourceIDA string, sourceTypeB string, sourceIDB string) (*CompareResponse, error) {
	slog.Info("Starting comparison", "sourceA", sourceTypeA, "sourceB", sourceTypeB)

	filesA, err := a.scanSource(sourceTypeA, sourceIDA)
	if err != nil {
		return nil, fmt.Errorf("source A scan failed: %w", err)
	}
	slog.Info("Source A scan complete", "type", sourceTypeA, "files", len(filesA))

	filesB, err := a.scanSource(sourceTypeB, sourceIDB)
	if err != nil {
		return nil, fmt.Errorf("source B scan failed: %w", err)
	}
	slog.Info("Source B scan complete", "type", sourceTypeB, "files", len(filesB))

	result := domain.CompareFiles(filesA, filesB)

	resp := &CompareResponse{
		SourceAOnly: toResultItems(result.LocalOnly),
		SourceBOnly: toResultItems(result.DriveOnly),
		Both:        toResultItems(result.Both),
	}

	slog.Info("Comparison complete",
		"sourceAOnly", len(resp.SourceAOnly),
		"sourceBOnly", len(resp.SourceBOnly),
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

type DuplicatePatternInfo struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Example string `json:"example"`
}

func (a *App) GetDuplicatePatterns() []DuplicatePatternInfo {
	examples := map[string]string{
		"windows_copy":   "photo (1).jpg, photo (2).jpg",
		"macos_copy":     "photo copy.jpg, photo copy 2.jpg",
		"underscore_num": "photo_1.jpg, photo_2.jpg",
		"dash_num":       "photo-1.jpg, photo-2.jpg",
	}

	patterns := domain.BuiltinPatterns()
	result := make([]DuplicatePatternInfo, len(patterns))
	for i, p := range patterns {
		result[i] = DuplicatePatternInfo{ID: p.ID, Label: p.Label, Example: examples[p.ID]}
	}
	return result
}

type DuplicateGroupResponse struct {
	BaseName     string              `json:"baseName"`
	Extension    string              `json:"extension"`
	RelativePath string              `json:"relativePath"`
	Files        []CompareResultItem `json:"files"`
}

type DuplicateScanResponse struct {
	Groups      []DuplicateGroupResponse `json:"groups"`
	TotalGroups int                      `json:"totalGroups"`
	TotalFiles  int                      `json:"totalFiles"`
}

func (a *App) RunDuplicateScan(sourceType string, sourceID string, patternIDs []string) (*DuplicateScanResponse, error) {
	if len(patternIDs) == 0 {
		return nil, fmt.Errorf("at least one pattern is required")
	}

	patterns := domain.GetPatternsByIDs(patternIDs)
	if len(patterns) == 0 {
		return nil, fmt.Errorf("no valid patterns selected")
	}

	files, err := a.scanSource(sourceType, sourceID)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	slog.Info("Duplicate scan complete", "source", sourceType, "files", len(files))

	groups := domain.DetectDuplicates(files, patterns)

	resp := &DuplicateScanResponse{
		Groups:      make([]DuplicateGroupResponse, len(groups)),
		TotalGroups: len(groups),
	}

	for i, g := range groups {
		resp.Groups[i] = DuplicateGroupResponse{
			BaseName:     g.BaseName,
			Extension:    g.Extension,
			RelativePath: g.RelativePath,
			Files:        toResultItems(g.Files),
		}
		resp.TotalFiles += len(g.Files)
	}

	slog.Info("Duplicates found", "groups", resp.TotalGroups, "files", resp.TotalFiles)

	return resp, nil
}

type ResizeAuditResponse struct {
	Pending  []CompareResultItem `json:"pending"`
	Resized  []CompareResultItem `json:"resized"`
	ExtraRes []CompareResultItem `json:"extraRes"`
}

func (a *App) RunResizeAudit(sourceType string, sourceID string, resizedType string, resizedID string, suffix string) (*ResizeAuditResponse, error) {
	if suffix == "" {
		return nil, fmt.Errorf("suffix is required")
	}

	slog.Info("Starting resize audit", "sourceType", sourceType, "resizedType", resizedType, "suffix", suffix)

	sourceFiles, err := a.scanSource(sourceType, sourceID)
	if err != nil {
		return nil, fmt.Errorf("source scan failed: %w", err)
	}
	slog.Info("Source scan complete", "files", len(sourceFiles))

	resizedFiles, err := a.scanSource(resizedType, resizedID)
	if err != nil {
		return nil, fmt.Errorf("resized scan failed: %w", err)
	}
	slog.Info("Resized scan complete", "files", len(resizedFiles))

	result := domain.ResizeAudit(sourceFiles, resizedFiles, suffix)

	resp := &ResizeAuditResponse{
		Pending:  toResultItems(result.Pending),
		Resized:  toResultItems(result.Resized),
		ExtraRes: toResultItems(result.ExtraRes),
	}

	slog.Info("Resize audit complete",
		"pending", len(resp.Pending),
		"resized", len(resp.Resized),
		"extraRes", len(resp.ExtraRes),
	)

	return resp, nil
}
