package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"personal-digital-inventory/internal/domain"
)

type LocalScanner struct{}

func NewLocalScanner() *LocalScanner {
	return &LocalScanner{}
}

func (s *LocalScanner) Scan(rootDir string) ([]domain.FileItem, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %q: %w", rootDir, err)
	}

	var items []domain.FileItem

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %q: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for %q: %w", path, err)
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %q: %w", path, err)
		}

		relPath = filepath.ToSlash(relPath)

		name, ext := splitNameExt(filepath.Base(path))

		relDir := filepath.ToSlash(filepath.Dir(relPath))
		if relDir == "." {
			relDir = ""
		} else {
			relDir += "/"
		}

		items = append(items, domain.FileItem{
			RelativePath: relDir,
			Name:         name,
			Extension:    ext,
			Size:         info.Size(),
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory %q: %w", absRoot, err)
	}

	return items, nil
}

func splitNameExt(filename string) (string, string) {
	if strings.HasPrefix(filename, ".") && strings.Count(filename, ".") == 1 {
		return filename, ""
	}
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return name, ext
}
