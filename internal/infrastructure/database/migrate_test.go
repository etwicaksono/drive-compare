package database

import (
	"testing"
)

func TestMigrations_EmbedExists(t *testing.T) {
	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		t.Fatalf("failed to read embedded migrations: %v", err)
	}

	if len(entries) == 0 {
		t.Fatal("no migration files found in embedded FS")
	}

	found := false
	for _, e := range entries {
		if e.Name() == "00001_create_initial_tables.sql" {
			found = true
			break
		}
	}

	if !found {
		t.Error("initial migration file not found in embedded FS")
	}
}

func TestMigrations_SQLContent(t *testing.T) {
	content, err := migrations.ReadFile("migrations/00001_create_initial_tables.sql")
	if err != nil {
		t.Fatalf("failed to read migration file: %v", err)
	}

	sql := string(content)

	required := []string{
		"CREATE TABLE comparisons",
		"CREATE TABLE comparison_files",
		"CREATE TABLE ignore_patterns",
		"-- +goose Up",
		"-- +goose Down",
	}

	for _, r := range required {
		if !contains(sql, r) {
			t.Errorf("migration SQL missing required content: %q", r)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRunMigrations_NilDB(t *testing.T) {
	err := RunMigrations(nil)
	if err == nil {
		t.Fatal("RunMigrations(nil) should return error")
	}
}

func TestRollbackMigration_NilDB(t *testing.T) {
	err := RollbackMigration(nil)
	if err == nil {
		t.Fatal("RollbackMigration(nil) should return error")
	}
}
