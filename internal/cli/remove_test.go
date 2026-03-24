package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
)

func TestRemoveWidgetFromConfig(t *testing.T) {
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

	err := cli.RunRemove([]string{"effort"}, dir)
	if err != nil {
		t.Fatalf("RunRemove returned error: %v", err)
	}

	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// effort should be gone from row 1.
	if len(loaded.Lines[0].Widgets) != 1 {
		t.Fatalf("expected 1 widget in row 1, got %d", len(loaded.Lines[0].Widgets))
	}
	if loaded.Lines[0].Widgets[0] != "model" {
		t.Fatalf("expected remaining widget to be 'model', got %s", loaded.Lines[0].Widgets[0])
	}
	// Row 2 should still exist.
	if len(loaded.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(loaded.Lines))
	}
}

func TestRemoveWidgetNotEnabled(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	// Removing a widget that's not in the config should not error.
	err := cli.RunRemove([]string{"weather"}, dir)
	if err != nil {
		t.Fatalf("RunRemove returned error: %v", err)
	}

	// Config should be unchanged.
	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if len(loaded.Lines[0].Widgets) != 1 {
		t.Fatalf("expected 1 widget (unchanged), got %d", len(loaded.Lines[0].Widgets))
	}
}

func TestRemoveWidgetCleansEmptyLines(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model"}},
			{Widgets: []string{"weather"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	// Remove the only widget on row 2 — the empty row should be cleaned up.
	err := cli.RunRemove([]string{"weather"}, dir)
	if err != nil {
		t.Fatalf("RunRemove returned error: %v", err)
	}

	loaded, err := config.Load(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	if len(loaded.Lines) != 1 {
		t.Fatalf("expected 1 line after removing only widget from row 2, got %d", len(loaded.Lines))
	}
	if loaded.Lines[0].Widgets[0] != "model" {
		t.Fatalf("expected remaining line to contain 'model', got %v", loaded.Lines[0].Widgets)
	}
}

func TestRemoveNoArgs(t *testing.T) {
	dir := t.TempDir()
	err := cli.RunRemove([]string{}, dir)
	if err == nil {
		t.Fatal("expected error for no args, got nil")
	}
}
