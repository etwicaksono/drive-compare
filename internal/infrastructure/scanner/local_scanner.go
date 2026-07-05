package scanner

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"personal-digital-inventory/internal/domain"
)

type LocalScanner struct {
	ignorePatterns   []string
	respectGitignore bool
	gitignoreCache   map[string][]string
}

func NewLocalScanner(ignorePatterns []string, respectGitignore bool) *LocalScanner {
	return &LocalScanner{
		ignorePatterns:   ignorePatterns,
		respectGitignore: respectGitignore,
		gitignoreCache:   make(map[string][]string),
	}
}

func (s *LocalScanner) Scan(ctx context.Context, rootDir string) ([]domain.FileItem, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for %q: %w", rootDir, err)
	}

	var items []domain.FileItem

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err != nil {
			return fmt.Errorf("error accessing %q: %w", path, err)
		}

		name := d.Name()

		if path != absRoot && s.shouldIgnore(name, filepath.Dir(path), absRoot) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			if s.respectGitignore {
				s.loadGitignore(path)
			}
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

		fileName, ext := splitNameExt(filepath.Base(path))

		relDir := filepath.ToSlash(filepath.Dir(relPath))
		if relDir == "." {
			relDir = ""
		} else {
			relDir += "/"
		}

		items = append(items, domain.FileItem{
			RelativePath: relDir,
			Name:         fileName,
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

func (s *LocalScanner) shouldIgnore(name string, dir string, root string) bool {
	for _, pattern := range s.ignorePatterns {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			return true
		}
	}

	if s.respectGitignore {
		current := dir
		for {
			if patterns, ok := s.gitignoreCache[current]; ok {
				for _, pattern := range patterns {
					matched, err := filepath.Match(pattern, name)
					if err == nil && matched {
						return true
					}
				}
			}
			if current == root {
				break
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}

	return false
}

func (s *LocalScanner) loadGitignore(dir string) {
	gitignorePath := filepath.Join(dir, ".gitignore")
	f, err := os.Open(gitignorePath)
	if err != nil {
		return
	}
	defer f.Close()

	var patterns []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSuffix(line, "/")
		if line != "" {
			patterns = append(patterns, line)
		}
	}

	if len(patterns) > 0 {
		s.gitignoreCache[dir] = patterns
	}
}

func splitNameExt(filename string) (string, string) {
	if strings.HasPrefix(filename, ".") && strings.Count(filename, ".") == 1 {
		return filename, ""
	}
	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	return name, ext
}
