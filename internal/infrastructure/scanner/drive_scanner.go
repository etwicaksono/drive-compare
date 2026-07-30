package scanner

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"drive-compare/internal/domain"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
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

const (
	defaultListFields = "nextPageToken, files(id, name, size, mimeType, trashed)"
	defaultGetFields  = "id, name, mimeType, trashed"
)

func (r *realDriveAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	if fields == "" {
		fields = defaultListFields
	}
	call := r.service.Files.List().
		Context(ctx).
		Q(query).
		Fields(googleapi.Field(fields)).
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
	return r.service.Files.Get(fileID).Context(ctx).Fields(googleapi.Field(defaultGetFields)).Do()
}

type TrashScope string

const (
	ScopeActive  TrashScope = "active"
	ScopeTrashed TrashScope = "trashed"
	ScopeAll     TrashScope = "all"
)

type DriveScanner struct {
	api            DriveAPI
	ignorePatterns []string
	scope          TrashScope
}

func NewDriveScanner(api DriveAPI, ignorePatterns []string, scope TrashScope) (*DriveScanner, error) {
	if scope == "" {
		scope = ScopeActive
	}
	switch scope {
	case ScopeActive, ScopeTrashed, ScopeAll:
	default:
		return nil, fmt.Errorf("invalid drive scope %q", scope)
	}
	return &DriveScanner{api: api, ignorePatterns: ignorePatterns, scope: scope}, nil
}

func (s *DriveScanner) Scan(ctx context.Context, folderID string) ([]domain.FileItem, error) {
	var items []domain.FileItem
	err := s.scanFolder(ctx, folderID, "", &items)
	if err != nil {
		return nil, fmt.Errorf("failed to scan Drive folder: %w", err)
	}
	return items, nil
}

func (s *DriveScanner) trashPredicate() string {
	switch s.scope {
	case ScopeTrashed:
		return " and trashed = true"
	case ScopeAll:
		return ""
	default:
		return " and trashed = false"
	}
}

func (s *DriveScanner) scanFolder(ctx context.Context, folderID string, relPath string, items *[]domain.FileItem) error {
	query := fmt.Sprintf("'%s' in parents%s", folderID, s.trashPredicate())
	var pageToken string

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		files, nextToken, err := s.api.ListFiles(ctx, query, defaultListFields, pageToken)
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
				Deleted:      f.Trashed,
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
