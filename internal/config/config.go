package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the full ccstatuswidgets configuration.
type Config struct {
	TimeoutMs int                                `json:"timeout_ms"`
	Lines     []LineConfig                       `json:"lines"`
	Widgets   map[string]map[string]interface{} `json:"widgets"`
}

// LineConfig defines which widgets appear on a single status line.
type LineConfig struct {
	Widgets []string `json:"widgets"`
}

// Load reads a JSON config file from path. If the file does not exist,
// it returns Default(). If the file exists but contains invalid JSON,
// it returns an error.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to path as pretty-printed JSON.
func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ConfigDir returns the default config directory: ~/.ccstatuswidgets.
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ccstatuswidgets")
	}
	return filepath.Join(home, ".ccstatuswidgets")
}
