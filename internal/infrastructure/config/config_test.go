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
	yaml := `
database:
  host: localhost
  port: 5432
  user: inventory
  password: secret
  name: pdi
  sslmode: disable
ignore:
  - node_modules
  - .git
logging:
  level: debug
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Database.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Database.Host, "localhost")
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Port = %d, want %d", cfg.Database.Port, 5432)
	}
	if cfg.Database.User != "inventory" {
		t.Errorf("User = %q, want %q", cfg.Database.User, "inventory")
	}
	if cfg.Database.Password != "secret" {
		t.Errorf("Password = %q, want %q", cfg.Database.Password, "secret")
	}
	if cfg.Database.Name != "pdi" {
		t.Errorf("Name = %q, want %q", cfg.Database.Name, "pdi")
	}
	if cfg.Database.SSLMode != "disable" {
		t.Errorf("SSLMode = %q, want %q", cfg.Database.SSLMode, "disable")
	}
	if len(cfg.Ignore) != 2 {
		t.Errorf("Ignore length = %d, want 2", len(cfg.Ignore))
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Level = %q, want %q", cfg.Logging.Level, "debug")
	}
}

func TestLoad_DefaultLoggingLevel(t *testing.T) {
	yaml := `
database:
  host: localhost
  port: 5432
  user: inventory
  name: pdi
  sslmode: disable
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Level = %q, want %q (default)", cfg.Logging.Level, "info")
	}
}

func TestLoad_EnvOverridePassword(t *testing.T) {
	yaml := `
database:
  host: localhost
  port: 5432
  user: inventory
  password: original
  name: pdi
  sslmode: disable
`
	t.Setenv("PDI_DB_PASSWORD", "from_env")

	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Database.Password != "from_env" {
		t.Errorf("Password = %q, want %q", cfg.Database.Password, "from_env")
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

func TestLoad_MissingRequiredFields(t *testing.T) {
	yaml := `
database:
  host: ""
  port: 0
  user: ""
  name: ""
  sslmode: ""
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("Load() should return error for missing required fields")
	}
}

func TestLoad_InvalidLoggingLevel(t *testing.T) {
	yaml := `
database:
  host: localhost
  port: 5432
  user: inventory
  name: pdi
  sslmode: disable
logging:
  level: invalid
`
	_, err := Load(writeTemp(t, yaml))
	if err == nil {
		t.Fatal("Load() should return error for invalid logging level")
	}
}

func TestLoad_EmptyIgnorePatterns(t *testing.T) {
	yaml := `
database:
  host: localhost
  port: 5432
  user: inventory
  name: pdi
  sslmode: disable
`
	cfg, err := Load(writeTemp(t, yaml))
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Ignore != nil && len(cfg.Ignore) != 0 {
		t.Errorf("Ignore = %v, want empty", cfg.Ignore)
	}
}
