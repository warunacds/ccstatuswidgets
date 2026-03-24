package cli_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

func setupAddRegistry() *widget.Registry {
	r := widget.NewRegistry()
	r.Register(&fakeWidget{name: "model"})
	r.Register(&fakeWidget{name: "effort"})
	r.Register(&fakeWidget{name: "weather"})
	r.Register(&fakeWidget{name: "moon"})
	return r
}

func TestAddWidgetToLastRow(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "effort"}},
			{Widgets: []string{"weather"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupAddRegistry()
	err := cli.RunAdd([]string{"moon"}, dir, registry)
	if err != nil {
		t.Fatalf("RunAdd returned error: %v", err)
	}

	// Reload and check.
	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// Should be appended to last row (row 2).
	if len(loaded.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(loaded.Lines))
	}
	lastLine := loaded.Lines[1]
	if len(lastLine.Widgets) != 2 {
		t.Fatalf("expected 2 widgets in last line, got %d", len(lastLine.Widgets))
	}
	if lastLine.Widgets[1] != "moon" {
		t.Fatalf("expected last widget to be 'moon', got %s", lastLine.Widgets[1])
	}
}

func TestAddWidgetToSpecificRow(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model"}},
			{Widgets: []string{"effort"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupAddRegistry()
	err := cli.RunAdd([]string{"weather", "-r", "1"}, dir, registry)
	if err != nil {
		t.Fatalf("RunAdd returned error: %v", err)
	}

	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// weather should be appended to row 1.
	if len(loaded.Lines[0].Widgets) != 2 {
		t.Fatalf("expected 2 widgets in row 1, got %d", len(loaded.Lines[0].Widgets))
	}
	if loaded.Lines[0].Widgets[1] != "weather" {
		t.Fatalf("expected second widget to be 'weather', got %s", loaded.Lines[0].Widgets[1])
	}
}

func TestAddWidgetAtSpecificPosition(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "effort"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupAddRegistry()
	// Insert weather at row 1, col 2 (between model and effort).
	err := cli.RunAdd([]string{"weather", "-r", "1", "-c", "2"}, dir, registry)
	if err != nil {
		t.Fatalf("RunAdd returned error: %v", err)
	}

	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if len(loaded.Lines[0].Widgets) != 3 {
		t.Fatalf("expected 3 widgets in row 1, got %d", len(loaded.Lines[0].Widgets))
	}
	expected := []string{"model", "weather", "effort"}
	for i, w := range expected {
		if loaded.Lines[0].Widgets[i] != w {
			t.Fatalf("widget %d: expected %s, got %s", i, w, loaded.Lines[0].Widgets[i])
		}
	}
}

func TestAddWidgetAlreadyEnabled(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "weather"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupAddRegistry()
	// weather is already in the config — should print message and not error.
	err := cli.RunAdd([]string{"weather"}, dir, registry)
	if err != nil {
		t.Fatalf("RunAdd returned error: %v", err)
	}

	// Config should be unchanged.
	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if len(loaded.Lines[0].Widgets) != 2 {
		t.Fatalf("expected 2 widgets (unchanged), got %d", len(loaded.Lines[0].Widgets))
	}
}

func TestAddWidgetCreatesNewRow(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupAddRegistry()
	// Add to row 3 — should create rows 2 and 3.
	err := cli.RunAdd([]string{"weather", "-r", "3"}, dir, registry)
	if err != nil {
		t.Fatalf("RunAdd returned error: %v", err)
	}

	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if len(loaded.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(loaded.Lines))
	}
	if len(loaded.Lines[1].Widgets) != 0 {
		t.Fatalf("expected row 2 to be empty, got %v", loaded.Lines[1].Widgets)
	}
	if len(loaded.Lines[2].Widgets) != 1 || loaded.Lines[2].Widgets[0] != "weather" {
		t.Fatalf("expected row 3 to contain [weather], got %v", loaded.Lines[2].Widgets)
	}
}

func TestAddWidgetNoArgs(t *testing.T) {
	dir := t.TempDir()

	// Ensure config dir exists but no config file needed since we error before loading.
	os.MkdirAll(dir, 0o755)

	registry := setupAddRegistry()
	err := cli.RunAdd([]string{}, dir, registry)
	if err == nil {
		t.Fatal("expected error for no args, got nil")
	}
}
