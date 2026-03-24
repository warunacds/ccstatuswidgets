package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
)

func TestRunTrack_SetsFlightInConfig(t *testing.T) {
	configDir := t.TempDir()

	err := cli.RunTrack([]string{"UL504"}, configDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.json to exist: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON in config.json: %v", err)
	}

	flightCfg, ok := cfg.Widgets["flight"]
	if !ok {
		t.Fatal("expected widgets.flight to exist in config")
	}

	flight, ok := flightCfg["flight"].(string)
	if !ok || flight != "UL504" {
		t.Errorf("expected flight %q, got %q", "UL504", flight)
	}
}

func TestRunTrack_StopRemovesFlightFromConfig(t *testing.T) {
	configDir := t.TempDir()

	// First, track a flight.
	if err := cli.RunTrack([]string{"UL504"}, configDir); err != nil {
		t.Fatalf("track error: %v", err)
	}

	// Then, stop tracking.
	if err := cli.RunTrack([]string{"stop"}, configDir); err != nil {
		t.Fatalf("stop error: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.json to exist: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("invalid JSON in config.json: %v", err)
	}

	flightCfg, ok := cfg.Widgets["flight"]
	if !ok {
		t.Fatal("expected widgets.flight map to still exist")
	}

	if _, exists := flightCfg["flight"]; exists {
		t.Error("expected flight key to be removed after stop")
	}
}

func TestRunTrack_ConfigFileCreatedCorrectly(t *testing.T) {
	configDir := t.TempDir()

	// Track a flight (config.json doesn't exist yet).
	if err := cli.RunTrack([]string{"SQ321"}, configDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	// Verify file exists and is valid JSON.
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config.json not created: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json contains invalid JSON: %v", err)
	}

	// Update with a different flight.
	if err := cli.RunTrack([]string{"EK422"}, configDir); err != nil {
		t.Fatalf("unexpected error on update: %v", err)
	}

	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("config.json not found after update: %v", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json invalid after update: %v", err)
	}

	flight, ok := cfg.Widgets["flight"]["flight"].(string)
	if !ok || flight != "EK422" {
		t.Errorf("expected flight %q after update, got %q", "EK422", flight)
	}
}
