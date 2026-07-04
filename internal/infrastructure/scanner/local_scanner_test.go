package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func createFile(t *testing.T, path string, size int) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestLocalScanner_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestLocalScanner_SingleFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "photo.jpg"), 1024)

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	item := items[0]
	if item.Name != "photo" {
		t.Errorf("Name = %q, want %q", item.Name, "photo")
	}
	if item.Extension != ".jpg" {
		t.Errorf("Extension = %q, want %q", item.Extension, ".jpg")
	}
	if item.Size != 1024 {
		t.Errorf("Size = %d, want %d", item.Size, 1024)
	}
	if item.RelativePath != "" {
		t.Errorf("RelativePath = %q, want %q", item.RelativePath, "")
	}
}

func TestLocalScanner_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "root.txt"), 10)
	createFile(t, filepath.Join(dir, "Documents", "doc.pdf"), 2048)
	createFile(t, filepath.Join(dir, "Documents", "Work", "report.docx"), 4096)
	createFile(t, filepath.Join(dir, "Photos", "DCIM", "img001.png"), 8192)

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	expected := map[string]struct {
		name    string
		ext     string
		relPath string
		size    int64
	}{
		"root.txt":    {name: "root", ext: ".txt", relPath: "", size: 10},
		"doc.pdf":     {name: "doc", ext: ".pdf", relPath: "Documents/", size: 2048},
		"report.docx": {name: "report", ext: ".docx", relPath: "Documents/Work/", size: 4096},
		"img001.png":  {name: "img001", ext: ".png", relPath: "Photos/DCIM/", size: 8192},
	}

	for _, item := range items {
		fullName := item.Name + item.Extension
		exp, ok := expected[fullName]
		if !ok {
			t.Errorf("unexpected file: %s", fullName)
			continue
		}
		if item.Name != exp.name {
			t.Errorf("%s: Name = %q, want %q", fullName, item.Name, exp.name)
		}
		if item.Extension != exp.ext {
			t.Errorf("%s: Extension = %q, want %q", fullName, item.Extension, exp.ext)
		}
		if item.RelativePath != exp.relPath {
			t.Errorf("%s: RelativePath = %q, want %q", fullName, item.RelativePath, exp.relPath)
		}
		if item.Size != exp.size {
			t.Errorf("%s: Size = %d, want %d", fullName, item.Size, exp.size)
		}
	}
}

func TestLocalScanner_FileWithoutExtension(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "Makefile"), 100)

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "Makefile" {
		t.Errorf("Name = %q, want %q", items[0].Name, "Makefile")
	}
	if items[0].Extension != "" {
		t.Errorf("Extension = %q, want empty", items[0].Extension)
	}
}

func TestLocalScanner_FileWithMultipleDots(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "archive.tar.gz"), 500)

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "archive.tar" {
		t.Errorf("Name = %q, want %q", items[0].Name, "archive.tar")
	}
	if items[0].Extension != ".gz" {
		t.Errorf("Extension = %q, want %q", items[0].Extension, ".gz")
	}
}

func TestLocalScanner_NonExistentDirectory(t *testing.T) {
	s := NewLocalScanner()
	_, err := s.Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("Scan() should return error for non-existent directory")
	}
}

func TestLocalScanner_ZeroSizeFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "empty.txt"), 0)

	s := NewLocalScanner()
	items, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Size != 0 {
		t.Errorf("Size = %d, want 0", items[0].Size)
	}
}

func TestSplitNameExt(t *testing.T) {
	tests := []struct {
		filename string
		wantName string
		wantExt  string
	}{
		{"photo.jpg", "photo", ".jpg"},
		{"document.pdf", "document", ".pdf"},
		{"archive.tar.gz", "archive.tar", ".gz"},
		{"Makefile", "Makefile", ""},
		{".gitignore", ".gitignore", ""},
		{"file.with.many.dots.txt", "file.with.many.dots", ".txt"},
	}

	for _, tt := range tests {
		name, ext := splitNameExt(tt.filename)
		if name != tt.wantName {
			t.Errorf("splitNameExt(%q) name = %q, want %q", tt.filename, name, tt.wantName)
		}
		if ext != tt.wantExt {
			t.Errorf("splitNameExt(%q) ext = %q, want %q", tt.filename, ext, tt.wantExt)
		}
	}
}
