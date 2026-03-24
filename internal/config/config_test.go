package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := &config.Config{
		TimeoutMs: 1000,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "cost"}},
		},
		Widgets: map[string]map[string]interface{}{
			"cost": {"detect_max_plan": false},
		},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if loaded.TimeoutMs != 1000 {
		t.Fatalf("expected timeout 1000, got %d", loaded.TimeoutMs)
	}
	if len(loaded.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(loaded.Lines))
	}
	if len(loaded.Lines[0].Widgets) != 2 {
		t.Fatalf("expected 2 widgets, got %d", len(loaded.Lines[0].Widgets))
	}
	if loaded.Lines[0].Widgets[0] != "model" {
		t.Fatalf("expected first widget 'model', got %s", loaded.Lines[0].Widgets[0])
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent.json")

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := config.Default()
	if cfg.TimeoutMs != def.TimeoutMs {
		t.Fatalf("expected default timeout %d, got %d", def.TimeoutMs, cfg.TimeoutMs)
	}
	if len(cfg.Lines) != len(def.Lines) {
		t.Fatalf("expected %d lines, got %d", len(def.Lines), len(cfg.Lines))
	}
}

func TestLoadInvalidJSONReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")

	if err := os.WriteFile(path, []byte("{not valid json!!!"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestDefaultConfigHasExpectedWidgets(t *testing.T) {
	cfg := config.Default()

	if cfg.TimeoutMs != 500 {
		t.Fatalf("expected timeout 500, got %d", cfg.TimeoutMs)
	}

	if len(cfg.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(cfg.Lines))
	}

	// Line 1 widgets
	expectedLine1 := []string{"model", "effort", "directory", "git-branch", "context-bar", "usage-5h", "usage-7d"}
	if len(cfg.Lines[0].Widgets) != len(expectedLine1) {
		t.Fatalf("line 1: expected %d widgets, got %d", len(expectedLine1), len(cfg.Lines[0].Widgets))
	}
	for i, w := range expectedLine1 {
		if cfg.Lines[0].Widgets[i] != w {
			t.Fatalf("line 1 widget %d: expected %s, got %s", i, w, cfg.Lines[0].Widgets[i])
		}
	}

	// Line 2 widgets
	expectedLine2 := []string{"lines-changed", "cost", "memory"}
	if len(cfg.Lines[1].Widgets) != len(expectedLine2) {
		t.Fatalf("line 2: expected %d widgets, got %d", len(expectedLine2), len(cfg.Lines[1].Widgets))
	}
	for i, w := range expectedLine2 {
		if cfg.Lines[1].Widgets[i] != w {
			t.Fatalf("line 2 widget %d: expected %s, got %s", i, w, cfg.Lines[1].Widgets[i])
		}
	}
}

func TestDefaultConfigWidgetOverrides(t *testing.T) {
	cfg := config.Default()

	// context-bar
	cb, ok := cfg.Widgets["context-bar"]
	if !ok {
		t.Fatal("expected context-bar widget config")
	}
	if bl, ok := cb["bar_length"]; !ok || int(bl.(float64)) != 10 {
		t.Fatalf("context-bar bar_length: expected 10, got %v", cb["bar_length"])
	}
	if sp, ok := cb["show_percentage"]; !ok || sp != true {
		t.Fatalf("context-bar show_percentage: expected true, got %v", cb["show_percentage"])
	}

	// usage-5h
	u5, ok := cfg.Widgets["usage-5h"]
	if !ok {
		t.Fatal("expected usage-5h widget config")
	}
	if bl, ok := u5["bar_length"]; !ok || int(bl.(float64)) != 10 {
		t.Fatalf("usage-5h bar_length: expected 10, got %v", u5["bar_length"])
	}
	if sp, ok := u5["show_pace"]; !ok || sp != true {
		t.Fatalf("usage-5h show_pace: expected true, got %v", u5["show_pace"])
	}

	// usage-7d
	u7, ok := cfg.Widgets["usage-7d"]
	if !ok {
		t.Fatal("expected usage-7d widget config")
	}
	if bl, ok := u7["bar_length"]; !ok || int(bl.(float64)) != 10 {
		t.Fatalf("usage-7d bar_length: expected 10, got %v", u7["bar_length"])
	}

	// cost
	cost, ok := cfg.Widgets["cost"]
	if !ok {
		t.Fatal("expected cost widget config")
	}
	if dmp, ok := cost["detect_max_plan"]; !ok || dmp != true {
		t.Fatalf("cost detect_max_plan: expected true, got %v", cost["detect_max_plan"])
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := config.Default()
	if err := config.Save(path, cfg); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	var check config.Config
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatalf("saved file is not valid JSON: %v", err)
	}

	// Load it back
	loaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if loaded.TimeoutMs != cfg.TimeoutMs {
		t.Fatalf("round-trip timeout: expected %d, got %d", cfg.TimeoutMs, loaded.TimeoutMs)
	}
	if len(loaded.Lines) != len(cfg.Lines) {
		t.Fatalf("round-trip lines: expected %d, got %d", len(cfg.Lines), len(loaded.Lines))
	}
}

func TestConfigDir(t *testing.T) {
	dir := config.ConfigDir()
	if dir == "" {
		t.Fatal("ConfigDir returned empty string")
	}
	if filepath.Base(dir) != ".ccstatuswidgets" {
		t.Fatalf("expected .ccstatuswidgets dir, got %s", filepath.Base(dir))
	}
}
