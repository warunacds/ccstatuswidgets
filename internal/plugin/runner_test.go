package plugin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestDiscoverPlugins_FindsValidPlugins(t *testing.T) {
	dir := t.TempDir()

	// Create two valid plugin directories
	for _, name := range []string{"alpha", "beta"} {
		pdir := filepath.Join(dir, name)
		if err := os.MkdirAll(pdir, 0o755); err != nil {
			t.Fatal(err)
		}
		manifest := PluginManifest{
			Name:    name,
			Version: "1.0.0",
			Entry:   "main.sh",
		}
		data, _ := json.Marshal(manifest)
		if err := os.WriteFile(filepath.Join(pdir, "plugin.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	plugins, err := DiscoverPlugins(dir)
	if err != nil {
		t.Fatalf("DiscoverPlugins returned error: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	names := map[string]bool{}
	for _, p := range plugins {
		names[p.Name()] = true
	}
	if !names["alpha"] || !names["beta"] {
		t.Errorf("expected plugins alpha and beta, got %v", names)
	}
}

func TestDiscoverPlugins_SkipsWithoutPluginJSON(t *testing.T) {
	dir := t.TempDir()

	// Directory with plugin.json
	validDir := filepath.Join(dir, "valid")
	if err := os.MkdirAll(validDir, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := PluginManifest{Name: "valid", Version: "1.0.0", Entry: "run.sh"}
	data, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(validDir, "plugin.json"), data, 0o644)

	// Directory without plugin.json
	emptyDir := filepath.Join(dir, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	plugins, err := DiscoverPlugins(dir)
	if err != nil {
		t.Fatalf("DiscoverPlugins returned error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name() != "valid" {
		t.Errorf("expected plugin name 'valid', got %q", plugins[0].Name())
	}
}

func TestDiscoverPlugins_SkipsInvalidJSON(t *testing.T) {
	dir := t.TempDir()

	// Valid plugin
	validDir := filepath.Join(dir, "good")
	os.MkdirAll(validDir, 0o755)
	manifest := PluginManifest{Name: "good", Version: "1.0.0", Entry: "run.sh"}
	data, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(validDir, "plugin.json"), data, 0o644)

	// Invalid JSON
	badDir := filepath.Join(dir, "bad")
	os.MkdirAll(badDir, 0o755)
	os.WriteFile(filepath.Join(badDir, "plugin.json"), []byte("{invalid json"), 0o644)

	plugins, err := DiscoverPlugins(dir)
	if err != nil {
		t.Fatalf("DiscoverPlugins returned error: %v", err)
	}
	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name() != "good" {
		t.Errorf("expected plugin name 'good', got %q", plugins[0].Name())
	}
}

func TestExternalWidget_Name(t *testing.T) {
	w := &ExternalWidget{
		manifest: PluginManifest{Name: "my-widget"},
	}
	if got := w.Name(); got != "my-widget" {
		t.Errorf("Name() = %q, want %q", got, "my-widget")
	}
}

func TestExternalWidget_Render_PipesStdinReadsStdout(t *testing.T) {
	dir := t.TempDir()

	// Create a script that reads stdin and writes a fixed JSON response
	script := `#!/bin/sh
read input
echo '{"text":"hello from plugin","color":"green"}'
`
	scriptPath := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	w := &ExternalWidget{
		manifest: PluginManifest{
			Name:  "test-plugin",
			Entry: "run.sh",
		},
		dir: dir,
	}

	input := &protocol.StatusLineInput{
		SessionID: "test-session",
		Model:     protocol.ModelInfo{ID: "claude-3", DisplayName: "Claude 3"},
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if out.Text != "hello from plugin" {
		t.Errorf("Text = %q, want %q", out.Text, "hello from plugin")
	}
	if out.Color != "green" {
		t.Errorf("Color = %q, want %q", out.Color, "green")
	}
}

func TestExternalWidget_Render_ErrorOnNonZeroExit(t *testing.T) {
	dir := t.TempDir()

	script := `#!/bin/sh
exit 1
`
	scriptPath := filepath.Join(dir, "fail.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	w := &ExternalWidget{
		manifest: PluginManifest{
			Name:  "fail-plugin",
			Entry: "fail.sh",
		},
		dir: dir,
	}

	input := &protocol.StatusLineInput{}
	_, err := w.Render(input, nil)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}

func TestExternalWidget_Render_ErrorOnInvalidJSON(t *testing.T) {
	dir := t.TempDir()

	script := `#!/bin/sh
echo 'not json at all'
`
	scriptPath := filepath.Join(dir, "badjson.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	w := &ExternalWidget{
		manifest: PluginManifest{
			Name:  "badjson-plugin",
			Entry: "badjson.sh",
		},
		dir: dir,
	}

	input := &protocol.StatusLineInput{}
	_, err := w.Render(input, nil)
	if err == nil {
		t.Fatal("expected error for invalid JSON output, got nil")
	}
}

func TestExternalWidget_Render_UsesInterpreter(t *testing.T) {
	dir := t.TempDir()

	// Create a .sh file that will be run with sh as interpreter
	script := `read input
echo '{"text":"interpreted","color":"blue"}'
`
	scriptPath := filepath.Join(dir, "plugin.sh")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}

	w := &ExternalWidget{
		manifest: PluginManifest{
			Name:        "interp-plugin",
			Entry:       "plugin.sh",
			Interpreter: "sh",
		},
		dir: dir,
	}

	input := &protocol.StatusLineInput{
		SessionID: "test",
	}

	out, err := w.Render(input, nil)
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if out.Text != "interpreted" {
		t.Errorf("Text = %q, want %q", out.Text, "interpreted")
	}
	if out.Color != "blue" {
		t.Errorf("Color = %q, want %q", out.Color, "blue")
	}
}

func TestDiscoverPlugins_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	plugins, err := DiscoverPlugins(dir)
	if err != nil {
		t.Fatalf("DiscoverPlugins returned error: %v", err)
	}
	if len(plugins) != 0 {
		t.Fatalf("expected 0 plugins, got %d", len(plugins))
	}
}

func TestDiscoverPlugins_NonexistentDir(t *testing.T) {
	_, err := DiscoverPlugins("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for nonexistent directory, got nil")
	}
}
