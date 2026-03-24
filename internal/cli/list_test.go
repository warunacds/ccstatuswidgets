package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
)

func setupRegistry() *widget.Registry {
	r := widget.NewRegistry()
	r.Register(&fakeWidget{name: "model"})
	r.Register(&fakeWidget{name: "effort"})
	r.Register(&fakeWidget{name: "weather"})
	return r
}

func writeConfig(t *testing.T, dir string, cfg *config.Config) {
	t.Helper()
	configPath := filepath.Join(dir, "config.json")
	if err := config.Save(configPath, cfg); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func TestListShowsBuiltinWidgetsWithStatus(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	os.MkdirAll(pluginsDir, 0o755)

	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "effort"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupRegistry()

	// Should not error.
	err := cli.RunList(dir, registry, pluginsDir)
	if err != nil {
		t.Fatalf("RunList returned error: %v", err)
	}
}

func TestListShowsPositionForEnabledWidgets(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	os.MkdirAll(pluginsDir, 0o755)

	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model", "effort"}},
			{Widgets: []string{"weather"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupRegistry()

	// Should not error — weather is on row 2 col 1.
	err := cli.RunList(dir, registry, pluginsDir)
	if err != nil {
		t.Fatalf("RunList returned error: %v", err)
	}
}

func TestListShowsDisabledForWidgetsNotInConfig(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")
	os.MkdirAll(pluginsDir, 0o755)

	cfg := &config.Config{
		TimeoutMs: 500,
		Lines: []config.LineConfig{
			{Widgets: []string{"model"}},
		},
		Widgets: map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupRegistry()

	// weather and effort should be shown as disabled.
	err := cli.RunList(dir, registry, pluginsDir)
	if err != nil {
		t.Fatalf("RunList returned error: %v", err)
	}
}

func TestListShowsPluginsSection(t *testing.T) {
	dir := t.TempDir()
	pluginsDir := filepath.Join(dir, "plugins")

	// Create a fake plugin.
	pluginDir := filepath.Join(pluginsDir, "my-plugin")
	os.MkdirAll(pluginDir, 0o755)
	manifest := map[string]string{"name": "my-plugin", "version": "1.0.0"}
	data, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644)

	cfg := &config.Config{
		TimeoutMs: 500,
		Lines:     []config.LineConfig{},
		Widgets:   map[string]map[string]interface{}{},
	}
	writeConfig(t, dir, cfg)

	registry := setupRegistry()

	err := cli.RunList(dir, registry, pluginsDir)
	if err != nil {
		t.Fatalf("RunList returned error: %v", err)
	}
}
