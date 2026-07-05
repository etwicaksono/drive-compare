package domain

import "testing"

func TestBuiltinPatterns(t *testing.T) {
	patterns := BuiltinPatterns()
	if len(patterns) != 4 {
		t.Fatalf("expected 4 patterns, got %d", len(patterns))
	}
	ids := map[string]bool{"windows_copy": true, "macos_copy": true, "underscore_num": true, "dash_num": true}
	for _, p := range patterns {
		if !ids[p.ID] {
			t.Errorf("unexpected pattern ID: %s", p.ID)
		}
	}
}

func TestExtractBaseName(t *testing.T) {
	patterns := BuiltinPatterns()
	tests := []struct {
		name string
		want string
	}{
		{"photo (1)", "photo"},
		{"photo (2)", "photo"},
		{"photo (10)", "photo"},
		{"report copy", "report"},
		{"report copy 2", "report"},
		{"data_3", "data"},
		{"file-7", "file"},
		{"normal", "normal"},
		{"my file (1)", "my file"},
		{"doc copy", "doc"},
		{"a_1", "a"},
		{"b-2", "b"},
		{"photo (1) (1)", "photo (1)"},
	}
	for _, tt := range tests {
		got := ExtractBaseName(tt.name, patterns)
		if got != tt.want {
			t.Errorf("ExtractBaseName(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestDetectDuplicates_BasicGrouping(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "photos/", Name: "photo", Extension: ".jpg", Size: 100},
		{RelativePath: "photos/", Name: "photo (1)", Extension: ".jpg", Size: 100},
		{RelativePath: "photos/", Name: "photo (2)", Extension: ".jpg", Size: 100},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("expected 3 files in group, got %d", len(groups[0].Files))
	}
	if groups[0].BaseName != "photo" {
		t.Errorf("BaseName = %q, want %q", groups[0].BaseName, "photo")
	}
}

func TestDetectDuplicates_MultiplePatterns(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "", Name: "report", Extension: ".pdf", Size: 500},
		{RelativePath: "", Name: "report copy", Extension: ".pdf", Size: 500},
		{RelativePath: "", Name: "report_1", Extension: ".pdf", Size: 500},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(groups[0].Files))
	}
}

func TestDetectDuplicates_UniqueFiles(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "", Name: "unique", Extension: ".txt", Size: 10},
		{RelativePath: "", Name: "another", Extension: ".doc", Size: 20},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(groups))
	}
}

func TestDetectDuplicates_DifferentPaths(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "a/", Name: "photo", Extension: ".jpg", Size: 100},
		{RelativePath: "b/", Name: "photo (1)", Extension: ".jpg", Size: 100},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups (different paths), got %d", len(groups))
	}
}

func TestDetectDuplicates_NoExtension(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "", Name: "Makefile", Extension: "", Size: 50},
		{RelativePath: "", Name: "Makefile (1)", Extension: "", Size: 50},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(groups[0].Files))
	}
}

func TestDetectDuplicates_Empty(t *testing.T) {
	patterns := BuiltinPatterns()
	groups := DetectDuplicates(nil, patterns)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for empty input, got %d", len(groups))
	}
}

func TestDetectDuplicates_SingleFile(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "", Name: "photo (1)", Extension: ".jpg", Size: 100},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 0 {
		t.Errorf("expected 0 groups for single file, got %d", len(groups))
	}
}

func TestDetectDuplicates_SortedByBaseName(t *testing.T) {
	patterns := BuiltinPatterns()
	files := []FileItem{
		{RelativePath: "", Name: "zebra", Extension: ".txt", Size: 10},
		{RelativePath: "", Name: "zebra (1)", Extension: ".txt", Size: 10},
		{RelativePath: "", Name: "apple", Extension: ".txt", Size: 10},
		{RelativePath: "", Name: "apple (1)", Extension: ".txt", Size: 10},
	}

	groups := DetectDuplicates(files, patterns)
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if groups[0].BaseName != "apple" {
		t.Errorf("first group BaseName = %q, want %q", groups[0].BaseName, "apple")
	}
	if groups[1].BaseName != "zebra" {
		t.Errorf("second group BaseName = %q, want %q", groups[1].BaseName, "zebra")
	}
}

func TestGetPatternsByIDs(t *testing.T) {
	tests := []struct {
		ids  []string
		want int
	}{
		{[]string{"windows_copy"}, 1},
		{[]string{"windows_copy", "macos_copy"}, 2},
		{[]string{"windows_copy", "macos_copy", "underscore_num", "dash_num"}, 4},
		{[]string{"nonexistent"}, 0},
		{[]string{}, 0},
		{nil, 0},
	}

	for _, tt := range tests {
		got := GetPatternsByIDs(tt.ids)
		if len(got) != tt.want {
			t.Errorf("GetPatternsByIDs(%v) returned %d patterns, want %d", tt.ids, len(got), tt.want)
		}
	}
}
