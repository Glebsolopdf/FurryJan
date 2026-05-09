package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LogLevel string

const (
	LogDebug LogLevel = "DEBUG"
	LogInfo  LogLevel = "INFO"
	LogWarn  LogLevel = "WARN"
	LogError LogLevel = "ERROR"
)

// Config holds application configuration
type Config struct {
	Username          string   `json:"username"`
	APIKey            string   `json:"api_key"`
	DownloadDir       string   `json:"download_dir"`
	DBPath            string   `json:"db_path"`
	RateLimitMS       int      `json:"rate_limit_ms"`
	AllowedTypes      []string `json:"allowed_types"`
	MaxSizeMB         int      `json:"max_size_mb"`
	BlobWriterEnabled bool     `json:"blob_writer_enabled"`
	BlobBufferMB      int      `json:"blob_buffer_mb"`
	BlobAutoCleanup   bool     `json:"blob_auto_cleanup"`
	LogLevel          LogLevel `json:"log_level"`
	Language          string   `json:"language"`
}

func getConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err == nil && configDir != "" {
		return filepath.Join(configDir, "furryjan")
	}

	homeDir, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "AppData", "Roaming", "furryjan")
	}
	return filepath.Join(homeDir, ".config", "furryjan")
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := getConfigDir()
	downloadDir := filepath.Join(homeDir, "Downloads", "Furryjan")
	if homeDir == "" {
		downloadDir = filepath.Join(".", "Downloads", "Furryjan")
	}
	return &Config{
		DownloadDir:       downloadDir,
		DBPath:            filepath.Join(configDir, "history.db"),
		RateLimitMS:       500,
		AllowedTypes:      []string{"image", "animation", "video"},
		MaxSizeMB:         0,
		BlobWriterEnabled: true,
		BlobBufferMB:      700,
		BlobAutoCleanup:   true,
		LogLevel:          LogInfo,
		Language:          "en",
	}
}

// ConfigPath returns the path to config file
func ConfigPath() (string, error) {
	configDir := getConfigDir()
	return filepath.Join(configDir, "config.json"), nil
}

// Exists checks if config file exists
func Exists() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Load loads config from file
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, fmt.Errorf("cannot determine config path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file '%s': %w", path, err)
	}

	cfg := Default()
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot parse config file: %w", err)
	}

	// Apply sensible defaults for new fields if they were not in the JSON
	// (for backward compatibility with old config files)
	if cfg.BlobBufferMB == 0 {
		cfg.BlobBufferMB = 700
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = LogInfo
	}

	// Expand ~ in paths
	cfg.DownloadDir = expandPath(cfg.DownloadDir)
	cfg.DBPath = expandPath(cfg.DBPath)

	return cfg, nil
}

// Save saves config to file with restricted permissions
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return fmt.Errorf("cannot determine config path: %w", err)
	}

	dir := filepath.Dir(path)
	err = os.MkdirAll(dir, 0700)
	if err != nil {
		return fmt.Errorf("cannot create config directory '%s': %w", dir, err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	// Write to temporary file first, then rename (atomic operation)
	tmpPath := path + ".tmp"
	err = os.WriteFile(tmpPath, data, 0600)
	if err != nil {
		return fmt.Errorf("cannot write temporary config file '%s': %w", tmpPath, err)
	}

	// Rename temp file to actual config file
	err = os.Rename(tmpPath, path)
	if err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("cannot save config file '%s': %w", path, err)
	}

	return nil
}

// IsComplete checks if configuration has all required fields
func (c *Config) IsComplete() bool {
	return c.Username != "" && c.APIKey != "" && c.DownloadDir != ""
}

// expandPath expands ~ to home directory
func expandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(homeDir, p[1:])
	}
	return p
}

// MaskAPIKey returns masked API key for display
func (c *Config) MaskAPIKey() string {
	if len(c.APIKey) <= 4 {
		return "****"
	}
	return "****" + c.APIKey[len(c.APIKey)-4:]
}
