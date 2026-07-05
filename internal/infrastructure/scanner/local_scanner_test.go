package scanner

import (
	"context"
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

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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
	s := NewLocalScanner(nil, false)
	_, err := s.Scan(context.Background(), "/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("Scan() should return error for non-existent directory")
	}
}

func TestLocalScanner_ZeroSizeFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "empty.txt"), 0)

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
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

func TestLocalScanner_IgnoreDirectory(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "keep.txt"), 10)
	createFile(t, filepath.Join(dir, "node_modules", "pkg", "index.js"), 100)
	createFile(t, filepath.Join(dir, ".git", "config"), 50)

	s := NewLocalScanner([]string{"node_modules", ".git"}, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "keep" {
		t.Errorf("Name = %q, want %q", items[0].Name, "keep")
	}
}

func TestLocalScanner_IgnoreFileByGlob(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "document.pdf"), 100)
	createFile(t, filepath.Join(dir, "temp.tmp"), 50)
	createFile(t, filepath.Join(dir, "cache.cache"), 30)

	s := NewLocalScanner([]string{"*.tmp", "*.cache"}, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "document" {
		t.Errorf("Name = %q, want %q", items[0].Name, "document")
	}
}

func TestLocalScanner_IgnoreSpecificFiles(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "photo.jpg"), 100)
	createFile(t, filepath.Join(dir, "Thumbs.db"), 50)
	createFile(t, filepath.Join(dir, "desktop.ini"), 30)

	s := NewLocalScanner([]string{"Thumbs.db", "desktop.ini"}, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "photo" {
		t.Errorf("Name = %q, want %q", items[0].Name, "photo")
	}
}

func TestLocalScanner_IgnoreNestedMatch(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "src", "app.ts"), 100)
	createFile(t, filepath.Join(dir, "src", "node_modules", "lib", "index.js"), 200)

	s := NewLocalScanner([]string{"node_modules"}, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "app" {
		t.Errorf("Name = %q, want %q", items[0].Name, "app")
	}
}

func TestLocalScanner_NoIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "node_modules", "pkg.js"), 100)
	createFile(t, filepath.Join(dir, "file.tmp"), 50)

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 2 {
		t.Errorf("expected 2 items (no ignore), got %d", len(items))
	}
}

func TestLocalScanner_EmptyIgnorePatterns(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "file.txt"), 10)

	s := NewLocalScanner([]string{}, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestShouldIgnore(t *testing.T) {
	s := NewLocalScanner([]string{"node_modules", ".git", "*.tmp", "Thumbs.db"}, false)

	tests := []struct {
		name string
		want bool
	}{
		{"node_modules", true},
		{".git", true},
		{"file.tmp", true},
		{"Thumbs.db", true},
		{"photo.jpg", false},
		{"document.pdf", false},
		{".gitignore", false},
	}

	for _, tt := range tests {
		got := s.shouldIgnore(tt.name, "", "")
		if got != tt.want {
			t.Errorf("shouldIgnore(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestLocalScanner_CancelledContext(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "file1.txt"), 10)
	createFile(t, filepath.Join(dir, "file2.txt"), 20)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewLocalScanner(nil, false)
	_, err := s.Scan(ctx, dir)
	if err == nil {
		t.Fatal("Scan() should return error for cancelled context")
	}
}

func TestLocalScanner_GitignoreBasic(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, ".gitignore"), 0)
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("dist\n*.log\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "dist"), 0755)
	createFile(t, filepath.Join(dir, "dist", "bundle.js"), 100)
	createFile(t, filepath.Join(dir, "error.log"), 50)
	createFile(t, filepath.Join(dir, "main.go"), 200)

	s := NewLocalScanner(nil, true)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items (.gitignore + main.go), got %d", len(items))
	}
}

func TestLocalScanner_GitignoreCommentAndBlank(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("# comment\n\nbuild\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "build"), 0755)
	createFile(t, filepath.Join(dir, "build", "app.exe"), 500)
	createFile(t, filepath.Join(dir, "readme.md"), 100)

	s := NewLocalScanner(nil, true)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items (.gitignore + readme.md), got %d", len(items))
	}
}

func TestLocalScanner_GitignoreDisabled(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("dist\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "dist"), 0755)
	createFile(t, filepath.Join(dir, "dist", "bundle.js"), 100)
	createFile(t, filepath.Join(dir, "main.go"), 200)

	s := NewLocalScanner(nil, false)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items (.gitignore + dist/bundle.js + main.go), got %d", len(items))
	}
}

func TestLocalScanner_GitignoreNested(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(sub, ".gitignore"), []byte("*.tmp\n"), 0644)
	createFile(t, filepath.Join(dir, "root.tmp"), 10)
	createFile(t, filepath.Join(sub, "nested.tmp"), 20)
	createFile(t, filepath.Join(sub, "keep.txt"), 30)

	s := NewLocalScanner(nil, true)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	names := make(map[string]bool)
	for _, item := range items {
		names[item.Name+item.Extension] = true
	}

	if !names["root.tmp"] {
		t.Error("root.tmp should NOT be ignored (no .gitignore in root)")
	}
	if names["nested.tmp"] {
		t.Error("nested.tmp should be ignored by sub/.gitignore")
	}
	if !names["keep.txt"] {
		t.Error("keep.txt should be kept")
	}
}

func TestLocalScanner_GitignoreTrailingSlash(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("vendor/\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "vendor"), 0755)
	createFile(t, filepath.Join(dir, "vendor", "lib.go"), 100)
	createFile(t, filepath.Join(dir, "main.go"), 200)

	s := NewLocalScanner(nil, true)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items (.gitignore + main.go), got %d", len(items))
	}
}

func TestLocalScanner_GitignoreNoFile(t *testing.T) {
	dir := t.TempDir()
	createFile(t, filepath.Join(dir, "file.txt"), 10)

	s := NewLocalScanner(nil, true)
	items, err := s.Scan(context.Background(), dir)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}
