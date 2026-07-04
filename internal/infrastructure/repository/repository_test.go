package repository

import (
	"testing"
	"time"

	"personal-digital-inventory/internal/domain"
)

var _ domain.ComparisonRepository = (*GormComparisonRepository)(nil)
var _ domain.IgnorePatternRepository = (*GormIgnorePatternRepository)(nil)

func TestComparisonModel_ToDomain(t *testing.T) {
	now := time.Now()
	finished := now.Add(time.Minute)

	model := ComparisonModel{
		ID:             1,
		LocalPath:      "/local/path",
		DrivePath:      "Drive/path",
		StartedAt:      now,
		FinishedAt:     &finished,
		TotalLocalOnly: 10,
		TotalDriveOnly: 5,
		TotalBoth:      100,
	}

	d := model.ToDomain()

	if d.ID != 1 {
		t.Errorf("ID = %d, want 1", d.ID)
	}
	if d.LocalPath != "/local/path" {
		t.Errorf("LocalPath = %q, want %q", d.LocalPath, "/local/path")
	}
	if d.DrivePath != "Drive/path" {
		t.Errorf("DrivePath = %q, want %q", d.DrivePath, "Drive/path")
	}
	if !d.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v, want %v", d.StartedAt, now)
	}
	if d.FinishedAt == nil || !d.FinishedAt.Equal(finished) {
		t.Errorf("FinishedAt = %v, want %v", d.FinishedAt, finished)
	}
	if d.TotalLocalOnly != 10 {
		t.Errorf("TotalLocalOnly = %d, want 10", d.TotalLocalOnly)
	}
	if d.TotalDriveOnly != 5 {
		t.Errorf("TotalDriveOnly = %d, want 5", d.TotalDriveOnly)
	}
	if d.TotalBoth != 100 {
		t.Errorf("TotalBoth = %d, want 100", d.TotalBoth)
	}
}

func TestComparisonFromDomain(t *testing.T) {
	now := time.Now()

	c := &domain.Comparison{
		ID:             2,
		LocalPath:      "D:\\Photos",
		DrivePath:      "Backup/Photos",
		StartedAt:      now,
		FinishedAt:     nil,
		TotalLocalOnly: 3,
		TotalDriveOnly: 7,
		TotalBoth:      50,
	}

	model := ComparisonFromDomain(c)

	if model.ID != 2 {
		t.Errorf("ID = %d, want 2", model.ID)
	}
	if model.LocalPath != "D:\\Photos" {
		t.Errorf("LocalPath = %q, want %q", model.LocalPath, "D:\\Photos")
	}
	if model.DrivePath != "Backup/Photos" {
		t.Errorf("DrivePath = %q, want %q", model.DrivePath, "Backup/Photos")
	}
	if model.FinishedAt != nil {
		t.Errorf("FinishedAt = %v, want nil", model.FinishedAt)
	}
	if model.TotalLocalOnly != 3 {
		t.Errorf("TotalLocalOnly = %d, want 3", model.TotalLocalOnly)
	}
}

func TestComparisonModel_TableName(t *testing.T) {
	model := ComparisonModel{}
	if model.TableName() != "comparisons" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "comparisons")
	}
}

func TestIgnorePatternModel_ToDomain(t *testing.T) {
	model := IgnorePatternModel{
		ID:      1,
		Pattern: "node_modules",
	}

	d := model.ToDomain()

	if d.ID != 1 {
		t.Errorf("ID = %d, want 1", d.ID)
	}
	if d.Pattern != "node_modules" {
		t.Errorf("Pattern = %q, want %q", d.Pattern, "node_modules")
	}
}

func TestIgnorePatternFromDomain(t *testing.T) {
	p := &domain.IgnorePattern{
		ID:      3,
		Pattern: ".git",
	}

	model := IgnorePatternFromDomain(p)

	if model.ID != 3 {
		t.Errorf("ID = %d, want 3", model.ID)
	}
	if model.Pattern != ".git" {
		t.Errorf("Pattern = %q, want %q", model.Pattern, ".git")
	}
}

func TestIgnorePatternModel_TableName(t *testing.T) {
	model := IgnorePatternModel{}
	if model.TableName() != "ignore_patterns" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "ignore_patterns")
	}
}

func TestComparisonFileModel_TableName(t *testing.T) {
	model := ComparisonFileModel{}
	if model.TableName() != "comparison_files" {
		t.Errorf("TableName() = %q, want %q", model.TableName(), "comparison_files")
	}
}

func TestComparisonModel_RoundTrip(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	finished := now.Add(5 * time.Minute)

	original := &domain.Comparison{
		LocalPath:      "/test/path",
		DrivePath:      "Drive/test",
		StartedAt:      now,
		FinishedAt:     &finished,
		TotalLocalOnly: 42,
		TotalDriveOnly: 13,
		TotalBoth:      999,
	}

	model := ComparisonFromDomain(original)
	result := model.ToDomain()

	if result.LocalPath != original.LocalPath {
		t.Errorf("LocalPath mismatch: got %q, want %q", result.LocalPath, original.LocalPath)
	}
	if result.DrivePath != original.DrivePath {
		t.Errorf("DrivePath mismatch: got %q, want %q", result.DrivePath, original.DrivePath)
	}
	if result.TotalLocalOnly != original.TotalLocalOnly {
		t.Errorf("TotalLocalOnly mismatch: got %d, want %d", result.TotalLocalOnly, original.TotalLocalOnly)
	}
	if result.TotalDriveOnly != original.TotalDriveOnly {
		t.Errorf("TotalDriveOnly mismatch: got %d, want %d", result.TotalDriveOnly, original.TotalDriveOnly)
	}
	if result.TotalBoth != original.TotalBoth {
		t.Errorf("TotalBoth mismatch: got %d, want %d", result.TotalBoth, original.TotalBoth)
	}
}

func TestIgnorePatternModel_RoundTrip(t *testing.T) {
	original := &domain.IgnorePattern{
		Pattern: "*.tmp",
	}

	model := IgnorePatternFromDomain(original)
	result := model.ToDomain()

	if result.Pattern != original.Pattern {
		t.Errorf("Pattern mismatch: got %q, want %q", result.Pattern, original.Pattern)
	}
}
