package domain

import "testing"

func TestResizeAudit_BasicMatching(t *testing.T) {
	source := []FileItem{
		{RelativePath: "movies/", Name: "video1", Extension: ".mp4", Size: 1000},
		{RelativePath: "movies/", Name: "video2", Extension: ".mp4", Size: 2000},
		{RelativePath: "movies/", Name: "video3", Extension: ".mp4", Size: 3000},
	}
	resized := []FileItem{
		{RelativePath: "movies/", Name: "video1_resized", Extension: ".mp4", Size: 500},
		{RelativePath: "movies/", Name: "video3_resized", Extension: ".mp4", Size: 1500},
	}

	result := ResizeAudit(source, resized, "_resized")

	if len(result.Pending) != 1 {
		t.Fatalf("Pending = %d, want 1", len(result.Pending))
	}
	if result.Pending[0].Name != "video2" {
		t.Errorf("Pending[0].Name = %q, want %q", result.Pending[0].Name, "video2")
	}
	if len(result.Resized) != 2 {
		t.Errorf("Resized = %d, want 2", len(result.Resized))
	}
	if len(result.ExtraRes) != 0 {
		t.Errorf("ExtraRes = %d, want 0", len(result.ExtraRes))
	}
}

func TestResizeAudit_AllPending(t *testing.T) {
	source := []FileItem{
		{RelativePath: "", Name: "a", Extension: ".mp4", Size: 100},
		{RelativePath: "", Name: "b", Extension: ".mp4", Size: 200},
	}

	result := ResizeAudit(source, nil, "_resized")

	if len(result.Pending) != 2 {
		t.Errorf("Pending = %d, want 2", len(result.Pending))
	}
	if len(result.Resized) != 0 {
		t.Errorf("Resized = %d, want 0", len(result.Resized))
	}
}

func TestResizeAudit_AllResized(t *testing.T) {
	source := []FileItem{
		{RelativePath: "", Name: "clip", Extension: ".avi", Size: 500},
	}
	resized := []FileItem{
		{RelativePath: "", Name: "clip_resized", Extension: ".avi", Size: 250},
	}

	result := ResizeAudit(source, resized, "_resized")

	if len(result.Pending) != 0 {
		t.Errorf("Pending = %d, want 0", len(result.Pending))
	}
	if len(result.Resized) != 1 {
		t.Errorf("Resized = %d, want 1", len(result.Resized))
	}
}

func TestResizeAudit_ExtraResized(t *testing.T) {
	source := []FileItem{
		{RelativePath: "", Name: "video1", Extension: ".mp4", Size: 1000},
	}
	resized := []FileItem{
		{RelativePath: "", Name: "video1_resized", Extension: ".mp4", Size: 500},
		{RelativePath: "", Name: "orphan_resized", Extension: ".mp4", Size: 300},
	}

	result := ResizeAudit(source, resized, "_resized")

	if len(result.ExtraRes) != 1 {
		t.Fatalf("ExtraRes = %d, want 1", len(result.ExtraRes))
	}
	if result.ExtraRes[0].Name != "orphan_resized" {
		t.Errorf("ExtraRes[0].Name = %q, want %q", result.ExtraRes[0].Name, "orphan_resized")
	}
}

func TestResizeAudit_DifferentPaths(t *testing.T) {
	source := []FileItem{
		{RelativePath: "folder_a/", Name: "video", Extension: ".mp4", Size: 1000},
	}
	resized := []FileItem{
		{RelativePath: "folder_b/", Name: "video_resized", Extension: ".mp4", Size: 500},
	}

	result := ResizeAudit(source, resized, "_resized")

	if len(result.Pending) != 1 {
		t.Errorf("Pending = %d, want 1 (different paths should not match)", len(result.Pending))
	}
	if len(result.ExtraRes) != 1 {
		t.Errorf("ExtraRes = %d, want 1", len(result.ExtraRes))
	}
}

func TestResizeAudit_CustomSuffix(t *testing.T) {
	source := []FileItem{
		{RelativePath: "", Name: "movie", Extension: ".mkv", Size: 5000},
	}
	resized := []FileItem{
		{RelativePath: "", Name: "movie_compressed", Extension: ".mkv", Size: 2000},
	}

	result := ResizeAudit(source, resized, "_compressed")

	if len(result.Resized) != 1 {
		t.Errorf("Resized = %d, want 1", len(result.Resized))
	}
	if len(result.Pending) != 0 {
		t.Errorf("Pending = %d, want 0", len(result.Pending))
	}
}

func TestResizeAudit_Empty(t *testing.T) {
	result := ResizeAudit(nil, nil, "_resized")

	if len(result.Pending) != 0 {
		t.Errorf("Pending = %d, want 0", len(result.Pending))
	}
	if len(result.Resized) != 0 {
		t.Errorf("Resized = %d, want 0", len(result.Resized))
	}
	if len(result.ExtraRes) != 0 {
		t.Errorf("ExtraRes = %d, want 0", len(result.ExtraRes))
	}
}

func TestResizeAudit_DifferentExtensions(t *testing.T) {
	source := []FileItem{
		{RelativePath: "", Name: "video", Extension: ".mp4", Size: 1000},
	}
	resized := []FileItem{
		{RelativePath: "", Name: "video_resized", Extension: ".avi", Size: 500},
	}

	result := ResizeAudit(source, resized, "_resized")

	if len(result.Pending) != 1 {
		t.Errorf("Pending = %d, want 1 (different extensions should not match)", len(result.Pending))
	}
}

func TestStripSuffix(t *testing.T) {
	tests := []struct {
		name   string
		suffix string
		want   string
	}{
		{"video_resized", "_resized", "video"},
		{"video", "_resized", "video"},
		{"a_resized_resized", "_resized", "a_resized"},
		{"_resized", "_resized", ""},
	}
	for _, tt := range tests {
		got := stripSuffix(tt.name, tt.suffix)
		if got != tt.want {
			t.Errorf("stripSuffix(%q, %q) = %q, want %q", tt.name, tt.suffix, got, tt.want)
		}
	}
}
