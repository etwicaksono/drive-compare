package config

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	t.Setenv("PDI_GOOGLE_CLIENT_ID", "")
	t.Setenv("PDI_GOOGLE_CLIENT_SECRET", "")

	yaml := `
google_drive:
  client_id: "test-id"
  client_secret: "test-secret"
ignore:
  patterns:
    - node_modules
    - .git
  respect_gitignore: false
logging:
  level: debug
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GoogleDrive.ClientID != "test-id" {
		t.Errorf("ClientID = %q, want %q", cfg.GoogleDrive.ClientID, "test-id")
	}
	if len(cfg.Ignore.Patterns) != 2 {
		t.Errorf("Ignore.Patterns length = %d, want 2", len(cfg.Ignore.Patterns))
	}
	if *cfg.Ignore.RespectGitignore != false {
		t.Errorf("RespectGitignore = %v, want false", *cfg.Ignore.RespectGitignore)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Level = %q, want %q", cfg.Logging.Level, "debug")
	}
}

func TestLoad_DefaultLoggingLevel(t *testing.T) {
	yaml := `
google_drive:
  token_file: token.json
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Level = %q, want %q (default)", cfg.Logging.Level, "info")
	}
}

func TestLoad_EnvOverrideGoogleCreds(t *testing.T) {
	yaml := `
google_drive:
  client_id: "original"
  client_secret: "original"
`
	t.Setenv("PDI_GOOGLE_CLIENT_ID", "from_env_id")
	t.Setenv("PDI_GOOGLE_CLIENT_SECRET", "from_env_secret")

	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GoogleDrive.ClientID != "from_env_id" {
		t.Errorf("ClientID = %q, want %q", cfg.GoogleDrive.ClientID, "from_env_id")
	}
	if cfg.GoogleDrive.ClientSecret != "from_env_secret" {
		t.Errorf("ClientSecret = %q, want %q", cfg.GoogleDrive.ClientSecret, "from_env_secret")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("Load() should return error for missing file")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	_, err := Load(writeTemp(t, ":::invalid"))
	if err == nil {
		t.Fatal("Load() should return error for invalid YAML")
	}
}

func TestLoad_InvalidLoggingLevel(t *testing.T) {
	yaml := `
logging:
  level: invalid
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("Load() should return error for invalid logging level")
	}
}

func TestLoad_DefaultRespectGitignore(t *testing.T) {
	yaml := `
google_drive:
  token_file: token.json
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if *cfg.Ignore.RespectGitignore != true {
		t.Errorf("RespectGitignore = %v, want true (default)", *cfg.Ignore.RespectGitignore)
	}
}

func TestLoad_DefaultTokenFile(t *testing.T) {
	yaml := `
google_drive:
  client_id: test
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.GoogleDrive.TokenFile != "token.json" {
		t.Errorf("TokenFile = %q, want %q", cfg.GoogleDrive.TokenFile, "token.json")
	}
}
