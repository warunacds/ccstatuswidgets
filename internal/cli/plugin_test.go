package cli_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
)

func TestPluginList_EmptyDir(t *testing.T) {
	pluginsDir := t.TempDir()

	err := cli.RunPlugin([]string{"list"}, pluginsDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPluginList_ShowsInstalledPlugin(t *testing.T) {
	pluginsDir := t.TempDir()

	// Create a fake plugin directory with a plugin.json.
	pluginDir := filepath.Join(pluginsDir, "my-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := map[string]interface{}{
		"name":        "my-plugin",
		"version":     "1.0.0",
		"description": "A test plugin",
		"entry":       "main.py",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	err := cli.RunPlugin([]string{"list"}, pluginsDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPluginRemove_DeletesPluginDir(t *testing.T) {
	pluginsDir := t.TempDir()

	// Create a fake plugin directory with a plugin.json.
	pluginDir := filepath.Join(pluginsDir, "removable-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	manifest := map[string]interface{}{
		"name":    "removable",
		"version": "1.0.0",
		"entry":   "main.py",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	err := cli.RunPlugin([]string{"remove", "removable"}, pluginsDir, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Directory should be gone.
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Error("expected plugin directory to be removed")
	}
}

func TestPluginRemove_ErrorsOnUnknownPlugin(t *testing.T) {
	pluginsDir := t.TempDir()

	err := cli.RunPlugin([]string{"remove", "nonexistent"}, pluginsDir, nil)
	if err == nil {
		t.Fatal("expected error when removing unknown plugin")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("expected error to mention plugin name, got: %v", err)
	}
}

func TestPluginAdd_ValidatesPluginJSON(t *testing.T) {
	pluginsDir := t.TempDir()

	// Create a bare git repo without a plugin.json.
	bareDir := t.TempDir()
	workDir := t.TempDir()

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	run(bareDir, "git", "init", "--bare", bareDir)
	run(workDir, "git", "clone", bareDir, workDir)
	run(workDir, "git", "config", "user.email", "test@test.com")
	run(workDir, "git", "config", "user.name", "Test")

	// Commit a dummy file (no plugin.json).
	dummyPath := filepath.Join(workDir, "README.md")
	if err := os.WriteFile(dummyPath, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	run(workDir, "git", "add", ".")
	run(workDir, "git", "commit", "-m", "initial")
	run(workDir, "git", "push", "origin", "HEAD")

	// Try to add — should fail because no plugin.json.
	err := cli.RunPlugin([]string{"add", bareDir}, pluginsDir, nil)
	if err == nil {
		t.Fatal("expected error when plugin.json is missing")
	}
	if !strings.Contains(err.Error(), "plugin.json") {
		t.Errorf("expected error to mention plugin.json, got: %v", err)
	}
}

func TestPluginAdd_RejectsBuiltinNameConflict(t *testing.T) {
	pluginsDir := t.TempDir()

	// Create a bare git repo with a plugin.json whose name conflicts.
	bareDir := t.TempDir()
	workDir := t.TempDir()

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command %v failed: %v\n%s", args, err, out)
		}
	}

	run(bareDir, "git", "init", "--bare", bareDir)
	run(workDir, "git", "clone", bareDir, workDir)
	run(workDir, "git", "config", "user.email", "test@test.com")
	run(workDir, "git", "config", "user.name", "Test")

	// Create plugin.json with a name that conflicts with a built-in.
	manifest := map[string]interface{}{
		"name":    "weather",
		"version": "1.0.0",
		"entry":   "main.py",
	}
	data, _ := json.Marshal(manifest)
	if err := os.WriteFile(filepath.Join(workDir, "plugin.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	run(workDir, "git", "add", ".")
	run(workDir, "git", "commit", "-m", "initial")
	run(workDir, "git", "push", "origin", "HEAD")

	builtins := []string{"model", "weather", "effort"}

	err := cli.RunPlugin([]string{"add", bareDir}, pluginsDir, builtins)
	if err == nil {
		t.Fatal("expected error when plugin name conflicts with built-in")
	}
	if !strings.Contains(err.Error(), "conflicts") {
		t.Errorf("expected error to mention conflict, got: %v", err)
	}
}
