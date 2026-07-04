package scanner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"personal-digital-inventory/internal/domain"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"net/http"
)

type DriveAPI interface {
	ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error)
	GetFile(ctx context.Context, fileID string) (*drive.File, error)
}

type realDriveAPI struct {
	service *drive.Service
}

func NewRealDriveAPI(client *http.Client) (*realDriveAPI, error) {
	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}
	return &realDriveAPI{service: srv}, nil
}

func (r *realDriveAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	call := r.service.Files.List().
		Context(ctx).
		Q(query).
		Fields("nextPageToken", "files(id, name, size, mimeType)").
		PageSize(1000)

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := call.Do()
	if err != nil {
		return nil, "", err
	}

	return result.Files, result.NextPageToken, nil
}

func (r *realDriveAPI) GetFile(ctx context.Context, fileID string) (*drive.File, error) {
	return r.service.Files.Get(fileID).Context(ctx).Fields("id, name, mimeType").Do()
}

type DriveScanner struct {
	api            DriveAPI
	ignorePatterns []string
}

func NewDriveScanner(api DriveAPI, ignorePatterns []string) *DriveScanner {
	return &DriveScanner{api: api, ignorePatterns: ignorePatterns}
}

func (s *DriveScanner) Scan(ctx context.Context, folderID string) ([]domain.FileItem, error) {
	var items []domain.FileItem
	err := s.scanFolder(ctx, folderID, "", &items)
	if err != nil {
		return nil, fmt.Errorf("failed to scan Drive folder: %w", err)
	}
	return items, nil
}

func (s *DriveScanner) scanFolder(ctx context.Context, folderID string, relPath string, items *[]domain.FileItem) error {
	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var pageToken string

	for {
		files, nextToken, err := s.api.ListFiles(ctx, query, "", pageToken)
		if err != nil {
			return fmt.Errorf("failed to list files in folder %s: %w", folderID, err)
		}

		for _, f := range files {
			if s.shouldIgnore(f.Name) {
				continue
			}

			if f.MimeType == "application/vnd.google-apps.folder" {
				subPath := relPath + f.Name + "/"
				if err := s.scanFolder(ctx, f.Id, subPath, items); err != nil {
					return err
				}
				continue
			}

			if isGoogleDoc(f.MimeType) {
				continue
			}

			name, ext := splitNameExt(f.Name)

			*items = append(*items, domain.FileItem{
				RelativePath: relPath,
				Name:         name,
				Extension:    ext,
				Size:         f.Size,
			})
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	return nil
}

func (s *DriveScanner) shouldIgnore(name string) bool {
	for _, pattern := range s.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func isGoogleDoc(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/vnd.google-apps.") &&
		mimeType != "application/vnd.google-apps.folder"
}
