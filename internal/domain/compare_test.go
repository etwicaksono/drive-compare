package domain

import (
	"testing"
)

func TestCompareFiles_AllLocal(t *testing.T) {
	local := []FileItem{
		{RelativePath: "", Name: "a", Extension: ".txt", Size: 100},
		{RelativePath: "docs/", Name: "b", Extension: ".pdf", Size: 200},
	}

	result := CompareFiles(local, nil)

	if len(result.LocalOnly) != 2 {
		t.Errorf("LocalOnly = %d, want 2", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 0 {
		t.Errorf("DriveOnly = %d, want 0", len(result.DriveOnly))
	}
	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0", len(result.Both))
	}
}

func TestCompareFiles_AllDrive(t *testing.T) {
	drive := []FileItem{
		{RelativePath: "", Name: "a", Extension: ".txt", Size: 100},
		{RelativePath: "photos/", Name: "b", Extension: ".jpg", Size: 5000},
	}

	result := CompareFiles(nil, drive)

	if len(result.LocalOnly) != 0 {
		t.Errorf("LocalOnly = %d, want 0", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 2 {
		t.Errorf("DriveOnly = %d, want 2", len(result.DriveOnly))
	}
	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0", len(result.Both))
	}
}

func TestCompareFiles_AllMatch(t *testing.T) {
	files := []FileItem{
		{RelativePath: "", Name: "a", Extension: ".txt", Size: 100},
		{RelativePath: "docs/", Name: "b", Extension: ".pdf", Size: 200},
	}

	result := CompareFiles(files, files)

	if len(result.LocalOnly) != 0 {
		t.Errorf("LocalOnly = %d, want 0", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 0 {
		t.Errorf("DriveOnly = %d, want 0", len(result.DriveOnly))
	}
	if len(result.Both) != 2 {
		t.Errorf("Both = %d, want 2", len(result.Both))
	}
}

func TestCompareFiles_Mixed(t *testing.T) {
	local := []FileItem{
		{RelativePath: "", Name: "shared", Extension: ".txt", Size: 100},
		{RelativePath: "", Name: "local_only", Extension: ".doc", Size: 200},
		{RelativePath: "photos/", Name: "img001", Extension: ".jpg", Size: 5000},
	}

	drive := []FileItem{
		{RelativePath: "", Name: "shared", Extension: ".txt", Size: 100},
		{RelativePath: "", Name: "drive_only", Extension: ".xls", Size: 300},
		{RelativePath: "photos/", Name: "img001", Extension: ".jpg", Size: 5000},
	}

	result := CompareFiles(local, drive)

	if len(result.LocalOnly) != 1 {
		t.Errorf("LocalOnly = %d, want 1", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 1 {
		t.Errorf("DriveOnly = %d, want 1", len(result.DriveOnly))
	}
	if len(result.Both) != 2 {
		t.Errorf("Both = %d, want 2", len(result.Both))
	}

	if result.LocalOnly[0].Name != "local_only" {
		t.Errorf("LocalOnly[0].Name = %q, want %q", result.LocalOnly[0].Name, "local_only")
	}
	if result.DriveOnly[0].Name != "drive_only" {
		t.Errorf("DriveOnly[0].Name = %q, want %q", result.DriveOnly[0].Name, "drive_only")
	}
}

func TestCompareFiles_SameNameDifferentSize(t *testing.T) {
	local := []FileItem{
		{RelativePath: "", Name: "file", Extension: ".txt", Size: 100},
	}
	drive := []FileItem{
		{RelativePath: "", Name: "file", Extension: ".txt", Size: 200},
	}

	result := CompareFiles(local, drive)

	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0 (different size)", len(result.Both))
	}
	if len(result.LocalOnly) != 1 {
		t.Errorf("LocalOnly = %d, want 1", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 1 {
		t.Errorf("DriveOnly = %d, want 1", len(result.DriveOnly))
	}
}

func TestCompareFiles_SameNameDifferentPath(t *testing.T) {
	local := []FileItem{
		{RelativePath: "dir1/", Name: "file", Extension: ".txt", Size: 100},
	}
	drive := []FileItem{
		{RelativePath: "dir2/", Name: "file", Extension: ".txt", Size: 100},
	}

	result := CompareFiles(local, drive)

	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0 (different path)", len(result.Both))
	}
	if len(result.LocalOnly) != 1 {
		t.Errorf("LocalOnly = %d, want 1", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 1 {
		t.Errorf("DriveOnly = %d, want 1", len(result.DriveOnly))
	}
}

func TestCompareFiles_EmptyBoth(t *testing.T) {
	result := CompareFiles(nil, nil)

	if len(result.LocalOnly) != 0 {
		t.Errorf("LocalOnly = %d, want 0", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 0 {
		t.Errorf("DriveOnly = %d, want 0", len(result.DriveOnly))
	}
	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0", len(result.Both))
	}
}

func TestCompareFiles_SameNameDifferentExtension(t *testing.T) {
	local := []FileItem{
		{RelativePath: "", Name: "file", Extension: ".txt", Size: 100},
	}
	drive := []FileItem{
		{RelativePath: "", Name: "file", Extension: ".pdf", Size: 100},
	}

	result := CompareFiles(local, drive)

	if len(result.Both) != 0 {
		t.Errorf("Both = %d, want 0 (different extension)", len(result.Both))
	}
	if len(result.LocalOnly) != 1 {
		t.Errorf("LocalOnly = %d, want 1", len(result.LocalOnly))
	}
	if len(result.DriveOnly) != 1 {
		t.Errorf("DriveOnly = %d, want 1", len(result.DriveOnly))
	}
}

func TestFileKey(t *testing.T) {
	f := FileItem{RelativePath: "docs/", Name: "report", Extension: ".pdf", Size: 4096}
	key := fileKey(f)
	expected := "docs/report.pdf|4096"

	if key != expected {
		t.Errorf("fileKey() = %q, want %q", key, expected)
	}
}
