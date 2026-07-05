package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database    DatabaseConfig    `yaml:"database"`
	GoogleDrive GoogleDriveConfig `yaml:"google_drive"`
	Ignore      IgnoreConfig      `yaml:"ignore"`
	Logging     LoggingConfig     `yaml:"logging"`
}

type GoogleDriveConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	TokenFile    string `yaml:"token_file"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type IgnoreConfig struct {
	Patterns         []string `yaml:"patterns"`
	RespectGitignore *bool    `yaml:"respect_gitignore"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	applyEnvOverrides(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if pw := os.Getenv("PDI_DB_PASSWORD"); pw != "" {
		cfg.Database.Password = pw
	}
	if id := os.Getenv("PDI_GOOGLE_CLIENT_ID"); id != "" {
		cfg.GoogleDrive.ClientID = id
	}
	if secret := os.Getenv("PDI_GOOGLE_CLIENT_SECRET"); secret != "" {
		cfg.GoogleDrive.ClientSecret = secret
	}
}

func validate(cfg *Config) error {
	var errs []string

	if cfg.Database.Host == "" {
		errs = append(errs, "database.host is required")
	}
	if cfg.Database.Port == 0 {
		errs = append(errs, "database.port is required")
	}
	if cfg.Database.User == "" {
		errs = append(errs, "database.user is required")
	}
	if cfg.Database.Name == "" {
		errs = append(errs, "database.name is required")
	}
	if cfg.Database.SSLMode == "" {
		errs = append(errs, "database.sslmode is required")
	}

	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if !validLevels[cfg.Logging.Level] {
		errs = append(errs, fmt.Sprintf("logging.level must be one of: debug, info, warn, error (got %q)", cfg.Logging.Level))
	}

	if cfg.GoogleDrive.TokenFile == "" {
		cfg.GoogleDrive.TokenFile = "token.json"
	}

	if cfg.Ignore.RespectGitignore == nil {
		defaultTrue := true
		cfg.Ignore.RespectGitignore = &defaultTrue
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}

	return nil
}
