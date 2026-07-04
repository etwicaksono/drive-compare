package scanner

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/api/drive/v3"
)

type mockDriveAPI struct {
	files map[string][]*drive.File
	err   error
}

func newMockDriveAPI() *mockDriveAPI {
	return &mockDriveAPI{files: make(map[string][]*drive.File)}
}

func (m *mockDriveAPI) addFiles(folderID string, files ...*drive.File) {
	m.files[folderID] = append(m.files[folderID], files...)
}

func (m *mockDriveAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	if m.err != nil {
		return nil, "", m.err
	}

	for folderID, files := range m.files {
		expected := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
		if query == expected {
			return files, "", nil
		}
	}

	return nil, "", nil
}

func (m *mockDriveAPI) GetFile(ctx context.Context, fileID string) (*drive.File, error) {
	return nil, nil
}

func TestDriveScanner_EmptyFolder(t *testing.T) {
	mock := newMockDriveAPI()
	s := NewDriveScanner(mock, nil)

	items, err := s.Scan(context.Background(), "root-folder-id")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestDriveScanner_SingleFile(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root", &drive.File{
		Id:       "file1",
		Name:     "photo.jpg",
		Size:     2048,
		MimeType: "image/jpeg",
	})

	s := NewDriveScanner(mock, nil)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != "photo" {
		t.Errorf("Name = %q, want %q", items[0].Name, "photo")
	}
	if items[0].Extension != ".jpg" {
		t.Errorf("Extension = %q, want %q", items[0].Extension, ".jpg")
	}
	if items[0].Size != 2048 {
		t.Errorf("Size = %d, want %d", items[0].Size, 2048)
	}
	if items[0].RelativePath != "" {
		t.Errorf("RelativePath = %q, want %q", items[0].RelativePath, "")
	}
}

func TestDriveScanner_NestedFolders(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "file1", Name: "root.txt", Size: 10, MimeType: "text/plain"},
		&drive.File{Id: "folder1", Name: "Documents", MimeType: "application/vnd.google-apps.folder"},
	)
	mock.addFiles("folder1",
		&drive.File{Id: "file2", Name: "doc.pdf", Size: 4096, MimeType: "application/pdf"},
		&drive.File{Id: "folder2", Name: "Work", MimeType: "application/vnd.google-apps.folder"},
	)
	mock.addFiles("folder2",
		&drive.File{Id: "file3", Name: "report.docx", Size: 8192, MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
	)

	s := NewDriveScanner(mock, nil)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	expected := map[string]string{
		"root.txt":    "",
		"doc.pdf":     "Documents/",
		"report.docx": "Documents/Work/",
	}

	for _, item := range items {
		fullName := item.Name + item.Extension
		wantPath, ok := expected[fullName]
		if !ok {
			t.Errorf("unexpected file: %s", fullName)
			continue
		}
		if item.RelativePath != wantPath {
			t.Errorf("%s: RelativePath = %q, want %q", fullName, item.RelativePath, wantPath)
		}
	}
}

func TestDriveScanner_SkipsGoogleDocs(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "file1", Name: "photo.jpg", Size: 1024, MimeType: "image/jpeg"},
		&drive.File{Id: "file2", Name: "My Doc", Size: 0, MimeType: "application/vnd.google-apps.document"},
		&drive.File{Id: "file3", Name: "My Sheet", Size: 0, MimeType: "application/vnd.google-apps.spreadsheet"},
		&drive.File{Id: "file4", Name: "My Slides", Size: 0, MimeType: "application/vnd.google-apps.presentation"},
	)

	s := NewDriveScanner(mock, nil)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item (only regular file), got %d", len(items))
	}

	if items[0].Name != "photo" {
		t.Errorf("Name = %q, want %q", items[0].Name, "photo")
	}
}

func TestDriveScanner_IgnorePatterns(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "file1", Name: "keep.txt", Size: 10, MimeType: "text/plain"},
		&drive.File{Id: "file2", Name: "temp.tmp", Size: 20, MimeType: "application/octet-stream"},
		&drive.File{Id: "folder1", Name: "node_modules", MimeType: "application/vnd.google-apps.folder"},
	)
	mock.addFiles("folder1",
		&drive.File{Id: "file3", Name: "pkg.js", Size: 100, MimeType: "text/javascript"},
	)

	s := NewDriveScanner(mock, []string{"*.tmp", "node_modules"})
	items, err := s.Scan(context.Background(), "root")
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

func TestDriveScanner_APIError(t *testing.T) {
	mock := newMockDriveAPI()
	mock.err = fmt.Errorf("API rate limit exceeded")

	s := NewDriveScanner(mock, nil)
	_, err := s.Scan(context.Background(), "root")
	if err == nil {
		t.Fatal("Scan() should return error when API fails")
	}
}

func TestIsGoogleDoc(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{"application/vnd.google-apps.document", true},
		{"application/vnd.google-apps.spreadsheet", true},
		{"application/vnd.google-apps.presentation", true},
		{"application/vnd.google-apps.folder", false},
		{"image/jpeg", false},
		{"application/pdf", false},
		{"text/plain", false},
	}

	for _, tt := range tests {
		got := isGoogleDoc(tt.mimeType)
		if got != tt.want {
			t.Errorf("isGoogleDoc(%q) = %v, want %v", tt.mimeType, got, tt.want)
		}
	}
}

func TestDriveScanner_CancelledContext(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root", &drive.File{
		Id:       "file1",
		Name:     "photo.jpg",
		Size:     1024,
		MimeType: "image/jpeg",
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s := NewDriveScanner(mock, nil)
	_, err := s.Scan(ctx, "root")
	if err == nil {
		t.Fatal("Scan() should return error for cancelled context")
	}
}

func TestDriveScanner_FileWithoutExtension(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root", &drive.File{
		Id:       "file1",
		Name:     "Makefile",
		Size:     100,
		MimeType: "application/octet-stream",
	})

	s := NewDriveScanner(mock, nil)
	items, err := s.Scan(context.Background(), "root")
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
