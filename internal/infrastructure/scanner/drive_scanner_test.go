package scanner

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type mockDriveAPI struct {
	files      map[string][]*drive.File
	err        error
	lastFields string
}

func newMockDriveAPI() *mockDriveAPI {
	return &mockDriveAPI{files: make(map[string][]*drive.File)}
}

func (m *mockDriveAPI) addFiles(folderID string, files ...*drive.File) {
	m.files[folderID] = append(m.files[folderID], files...)
}

func (m *mockDriveAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	m.lastFields = fields
	if m.err != nil {
		return nil, "", m.err
	}

	for folderID, files := range m.files {
		activeQ := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
		trashedQ := fmt.Sprintf("'%s' in parents and trashed = true", folderID)
		allQ := fmt.Sprintf("'%s' in parents", folderID)

		var matched []*drive.File
		switch query {
		case activeQ:
			for _, f := range files {
				if !f.Trashed {
					matched = append(matched, f)
				}
			}
		case trashedQ:
			for _, f := range files {
				if f.Trashed {
					matched = append(matched, f)
				}
			}
		case allQ:
			matched = files
		}
		if matched != nil || query == activeQ || query == trashedQ || query == allQ {
			return matched, "", nil
		}
	}

	return nil, "", nil
}

func (m *mockDriveAPI) GetFile(ctx context.Context, fileID string) (*drive.File, error) {
	return nil, nil
}

func mustDriveScanner(t *testing.T, api DriveAPI, ignorePatterns []string, scope TrashScope) *DriveScanner {
	t.Helper()
	s, err := NewDriveScanner(api, ignorePatterns, scope)
	if err != nil {
		t.Fatalf("NewDriveScanner() error: %v", err)
	}
	return s
}

func TestDriveScanner_EmptyFolder(t *testing.T) {
	mock := newMockDriveAPI()
	s := mustDriveScanner(t, mock, nil, ScopeActive)

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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

	s := mustDriveScanner(t, mock, []string{"*.tmp", "node_modules"}, ScopeActive)
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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

	s := mustDriveScanner(t, mock, nil, ScopeActive)
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

func TestDriveScanner_ScopeActiveExcludesTrashed(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "active.jpg", Size: 10, MimeType: "image/jpeg"},
		&drive.File{Id: "f2", Name: "gone.jpg", Size: 20, MimeType: "image/jpeg", Trashed: true},
	)

	s := mustDriveScanner(t, mock, nil, ScopeActive)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "active" {
		t.Errorf("Name = %q, want %q", items[0].Name, "active")
	}
	if items[0].Deleted {
		t.Error("Deleted = true, want false")
	}
}

func TestDriveScanner_ScopeTrashedOnly(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "active.jpg", Size: 10, MimeType: "image/jpeg"},
		&drive.File{Id: "f2", Name: "gone.jpg", Size: 20, MimeType: "image/jpeg", Trashed: true},
	)

	s := mustDriveScanner(t, mock, nil, ScopeTrashed)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Name != "gone" {
		t.Errorf("Name = %q, want %q", items[0].Name, "gone")
	}
	if !items[0].Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestDriveScanner_ScopeAllIncludesBoth(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "photo.jpg", Size: 10, MimeType: "image/jpeg"},
		&drive.File{Id: "f2", Name: "photo.jpg", Size: 20, MimeType: "image/jpeg", Trashed: true},
	)

	s := mustDriveScanner(t, mock, nil, ScopeAll)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	var active, deleted int
	for _, item := range items {
		if item.Deleted {
			deleted++
		} else {
			active++
		}
	}
	if active != 1 || deleted != 1 {
		t.Errorf("active = %d, deleted = %d, want 1 and 1", active, deleted)
	}
}

func TestDriveScanner_EmptyScopeDefaultsToActive(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "active.jpg", Size: 10, MimeType: "image/jpeg"},
		&drive.File{Id: "f2", Name: "gone.jpg", Size: 20, MimeType: "image/jpeg", Trashed: true},
	)

	s, err := NewDriveScanner(mock, nil, "")
	if err != nil {
		t.Fatalf("NewDriveScanner() error: %v", err)
	}
	if s.scope != ScopeActive {
		t.Errorf("scope = %q, want %q", s.scope, ScopeActive)
	}

	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

func TestDriveScanner_InvalidScope(t *testing.T) {
	mock := newMockDriveAPI()

	s, err := NewDriveScanner(mock, nil, TrashScope("bogus"))
	if err == nil {
		t.Fatal("NewDriveScanner() should return error for invalid scope")
	}
	if s != nil {
		t.Error("scanner should be nil on error")
	}
}

func TestDriveScanner_TrashedFolderIsTraversed(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "folder1", Name: "Archive", MimeType: "application/vnd.google-apps.folder", Trashed: true},
	)
	mock.addFiles("folder1",
		&drive.File{Id: "f1", Name: "old.jpg", Size: 30, MimeType: "image/jpeg", Trashed: true},
	)

	s := mustDriveScanner(t, mock, nil, ScopeTrashed)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].RelativePath != "Archive/" {
		t.Errorf("RelativePath = %q, want %q", items[0].RelativePath, "Archive/")
	}
	if !items[0].Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestDriveScanner_IgnorePatternsApplyToTrashed(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "keep.jpg", Size: 10, MimeType: "image/jpeg", Trashed: true},
		&drive.File{Id: "f2", Name: "scratch.tmp", Size: 20, MimeType: "application/octet-stream", Trashed: true},
	)

	s := mustDriveScanner(t, mock, []string{"*.tmp"}, ScopeTrashed)
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

func TestDriveScanner_GoogleDocsSkippedInTrash(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root",
		&drive.File{Id: "f1", Name: "photo.jpg", Size: 10, MimeType: "image/jpeg", Trashed: true},
		&drive.File{Id: "f2", Name: "My Doc", MimeType: "application/vnd.google-apps.document", Trashed: true},
	)

	s := mustDriveScanner(t, mock, nil, ScopeTrashed)
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
}

func TestDriveScanner_RequestsTrashedField(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root", &drive.File{Id: "f1", Name: "photo.jpg", Size: 10, MimeType: "image/jpeg"})

	s := mustDriveScanner(t, mock, nil, ScopeActive)
	if _, err := s.Scan(context.Background(), "root"); err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if mock.lastFields == "" {
		t.Fatal("scanner passed an empty fields argument")
	}
	if !strings.Contains(mock.lastFields, "trashed") {
		t.Errorf("fields = %q, want it to include %q", mock.lastFields, "trashed")
	}
}

// pagedDriveAPI serves a fixed set of pages so the pagination loop in scanFolder
// can be exercised without a live Drive service.
type pagedDriveAPI struct {
	pages      [][]*drive.File
	tokensSeen []string
	calls      int
}

func (p *pagedDriveAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	p.tokensSeen = append(p.tokensSeen, pageToken)
	idx := p.calls
	p.calls++
	if idx >= len(p.pages) {
		return nil, "", nil
	}
	nextToken := ""
	if idx+1 < len(p.pages) {
		nextToken = fmt.Sprintf("token-%d", idx+1)
	}
	return p.pages[idx], nextToken, nil
}

func (p *pagedDriveAPI) GetFile(ctx context.Context, fileID string) (*drive.File, error) {
	return nil, nil
}

func TestDriveScanner_FollowsPagination(t *testing.T) {
	api := &pagedDriveAPI{pages: [][]*drive.File{
		{{Id: "f1", Name: "a.jpg", Size: 1, MimeType: "image/jpeg"}},
		{{Id: "f2", Name: "b.jpg", Size: 2, MimeType: "image/jpeg"}},
		{{Id: "f3", Name: "c.jpg", Size: 3, MimeType: "image/jpeg"}},
	}}

	s := mustDriveScanner(t, api, nil, ScopeActive)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("expected 3 items across pages, got %d", len(items))
	}
	wantTokens := []string{"", "token-1", "token-2"}
	if len(api.tokensSeen) != len(wantTokens) {
		t.Fatalf("page calls = %d, want %d", len(api.tokensSeen), len(wantTokens))
	}
	for i, want := range wantTokens {
		if api.tokensSeen[i] != want {
			t.Errorf("call %d pageToken = %q, want %q", i, api.tokensSeen[i], want)
		}
	}
}

func TestDriveScanner_TrashPredicate(t *testing.T) {
	tests := []struct {
		scope TrashScope
		want  string
	}{
		{ScopeActive, " and trashed = false"},
		{ScopeTrashed, " and trashed = true"},
		{ScopeAll, ""},
	}

	for _, tt := range tests {
		s := mustDriveScanner(t, newMockDriveAPI(), nil, tt.scope)
		if got := s.trashPredicate(); got != tt.want {
			t.Errorf("scope %q: trashPredicate() = %q, want %q", tt.scope, got, tt.want)
		}
	}
}

// failingSubfolderAPI succeeds for the root listing but fails once the scanner
// descends into the child folder.
type failingSubfolderAPI struct {
	childID string
	err     error
}

func (f *failingSubfolderAPI) ListFiles(ctx context.Context, query string, fields string, pageToken string) ([]*drive.File, string, error) {
	if strings.Contains(query, f.childID) {
		return nil, "", f.err
	}
	return []*drive.File{
		{Id: f.childID, Name: "Sub", MimeType: "application/vnd.google-apps.folder"},
	}, "", nil
}

func (f *failingSubfolderAPI) GetFile(ctx context.Context, fileID string) (*drive.File, error) {
	return nil, nil
}

func TestDriveScanner_SubfolderErrorPropagates(t *testing.T) {
	api := &failingSubfolderAPI{childID: "child", err: fmt.Errorf("permission denied")}

	s := mustDriveScanner(t, api, nil, ScopeActive)
	_, err := s.Scan(context.Background(), "root")
	if err == nil {
		t.Fatal("Scan() should return error when a subfolder listing fails")
	}
	if !strings.Contains(err.Error(), "child") {
		t.Errorf("error = %q, want it to mention the failing folder ID", err.Error())
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error = %q, want it to wrap the API error", err.Error())
	}
}

func TestDriveScanner_InvalidIgnorePatternIsSkipped(t *testing.T) {
	mock := newMockDriveAPI()
	mock.addFiles("root", &drive.File{Id: "f1", Name: "photo.jpg", Size: 10, MimeType: "image/jpeg"})

	// "[" is malformed: filepath.Match returns an error, which must not count as a match.
	s := mustDriveScanner(t, mock, []string{"["}, ScopeActive)
	items, err := s.Scan(context.Background(), "root")
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
}

// newTestRealDriveAPI builds a realDriveAPI pointed at a local test server so the
// Drive wrappers can be exercised without network access.
func newTestRealDriveAPI(t *testing.T, srv *httptest.Server) *realDriveAPI {
	t.Helper()
	svc, err := drive.NewService(context.Background(),
		option.WithoutAuthentication(),
		option.WithEndpoint(srv.URL),
		option.WithHTTPClient(srv.Client()),
	)
	if err != nil {
		t.Fatalf("drive.NewService() error: %v", err)
	}
	return &realDriveAPI{service: svc}
}

func TestNewRealDriveAPI(t *testing.T) {
	api, err := NewRealDriveAPI(http.DefaultClient)
	if err != nil {
		t.Fatalf("NewRealDriveAPI() error: %v", err)
	}
	if api == nil || api.service == nil {
		t.Fatal("NewRealDriveAPI() returned an API without a Drive service")
	}
}

func TestRealDriveAPI_ListFilesRequestsDefaultFields(t *testing.T) {
	var gotQuery, gotFields, gotPageToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("q")
		gotFields = r.URL.Query().Get("fields")
		gotPageToken = r.URL.Query().Get("pageToken")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"nextPageToken":"next","files":[{"id":"f1","name":"a.jpg","mimeType":"image/jpeg","size":"12","trashed":true}]}`)
	}))
	defer srv.Close()

	api := newTestRealDriveAPI(t, srv)

	files, next, err := api.ListFiles(context.Background(), "'root' in parents", "", "page-2")
	if err != nil {
		t.Fatalf("ListFiles() error: %v", err)
	}
	if len(files) != 1 || files[0].Id != "f1" {
		t.Fatalf("unexpected files: %+v", files)
	}
	if !files[0].Trashed {
		t.Error("Trashed = false, want true")
	}
	if next != "next" {
		t.Errorf("nextPageToken = %q, want %q", next, "next")
	}
	if gotQuery != "'root' in parents" {
		t.Errorf("q = %q, want %q", gotQuery, "'root' in parents")
	}
	if gotFields != defaultListFields {
		t.Errorf("fields = %q, want %q", gotFields, defaultListFields)
	}
	if gotPageToken != "page-2" {
		t.Errorf("pageToken = %q, want %q", gotPageToken, "page-2")
	}
}

func TestRealDriveAPI_ListFilesHonorsExplicitFields(t *testing.T) {
	var gotFields, gotPageToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = r.URL.Query().Get("fields")
		gotPageToken = r.URL.Query().Get("pageToken")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"files":[]}`)
	}))
	defer srv.Close()

	api := newTestRealDriveAPI(t, srv)

	if _, _, err := api.ListFiles(context.Background(), "q", "files(id)", ""); err != nil {
		t.Fatalf("ListFiles() error: %v", err)
	}
	if gotFields != "files(id)" {
		t.Errorf("fields = %q, want %q", gotFields, "files(id)")
	}
	if gotPageToken != "" {
		t.Errorf("pageToken = %q, want it to be omitted", gotPageToken)
	}
}

func TestRealDriveAPI_ListFilesError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":{"code":403,"message":"insufficient permissions"}}`)
	}))
	defer srv.Close()

	api := newTestRealDriveAPI(t, srv)

	if _, _, err := api.ListFiles(context.Background(), "q", "", ""); err == nil {
		t.Fatal("ListFiles() should return an error on HTTP failure")
	}
}

func TestRealDriveAPI_GetFileRequestsTrashedField(t *testing.T) {
	var gotFields, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFields = r.URL.Query().Get("fields")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":"file-1","name":"Folder","mimeType":"application/vnd.google-apps.folder","trashed":true}`)
	}))
	defer srv.Close()

	api := newTestRealDriveAPI(t, srv)

	f, err := api.GetFile(context.Background(), "file-1")
	if err != nil {
		t.Fatalf("GetFile() error: %v", err)
	}
	if f.Id != "file-1" || !f.Trashed {
		t.Errorf("unexpected file: %+v", f)
	}
	if gotFields != defaultGetFields {
		t.Errorf("fields = %q, want %q", gotFields, defaultGetFields)
	}
	if !strings.HasSuffix(gotPath, "/files/file-1") {
		t.Errorf("path = %q, want it to target /files/file-1", gotPath)
	}
}

func TestRealDriveAPI_GetFileError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":{"code":404,"message":"File not found"}}`)
	}))
	defer srv.Close()

	api := newTestRealDriveAPI(t, srv)

	if _, err := api.GetFile(context.Background(), "missing"); err == nil {
		t.Fatal("GetFile() should return an error when the file is missing")
	}
}
