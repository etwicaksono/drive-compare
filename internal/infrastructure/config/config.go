package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GoogleDrive GoogleDriveConfig `yaml:"google_drive"`
	Ignore      IgnoreConfig      `yaml:"ignore"`
	Logging     LoggingConfig     `yaml:"logging"`
}

type GoogleDriveConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	TokenFile    string `yaml:"token_file"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type IgnoreConfig struct {
	Patterns         []string `yaml:"patterns"`
	RespectGitignore *bool    `yaml:"respect_gitignore"`
}

func Load(path string) (*Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = defaultConfig()
			applyEnvOverrides(&cfg)
			if verr := validate(&cfg); verr != nil {
				return nil, fmt.Errorf("config validation failed: %w", verr)
			}
			if werr := writeDefault(path, &cfg); werr != nil {
				// non-fatal: config works in memory even if write fails
				fmt.Fprintf(os.Stderr, "warning: could not write default config to %s: %v\n", path, werr)
			}
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	applyEnvOverrides(&cfg)

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func defaultConfig() Config {
	respectGitignore := true
	return Config{
		GoogleDrive: GoogleDriveConfig{
			TokenFile: "token.json",
		},
		Ignore: IgnoreConfig{
			Patterns:         []string{"node_modules", ".git", "Thumbs.db", "desktop.ini", "*.tmp", "*.cache"},
			RespectGitignore: &respectGitignore,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

func writeDefault(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func applyEnvOverrides(cfg *Config) {
	if id := os.Getenv("PDI_GOOGLE_CLIENT_ID"); id != "" {
		cfg.GoogleDrive.ClientID = id
	}
	if secret := os.Getenv("PDI_GOOGLE_CLIENT_SECRET"); secret != "" {
		cfg.GoogleDrive.ClientSecret = secret
	}
}

func validate(cfg *Config) error {
	var errs []string

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
