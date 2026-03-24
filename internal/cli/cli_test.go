package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
)

func TestRunInitCreatesDirsAndConfig(t *testing.T) {
	// Use a temp dir as a fake home to isolate from real system.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	err := cli.RunInit()
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}

	configDir := filepath.Join(tmpHome, ".ccstatuswidgets")

	// Check directories were created.
	for _, sub := range []string{"", "cache", "plugins"} {
		dir := filepath.Join(configDir, sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}

	// Check config file was created and is valid JSON.
	configPath := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config.json to exist: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("config.json is not valid JSON: %v", err)
	}

	// Verify it has default values.
	def := config.Default()
	if cfg.TimeoutMs != def.TimeoutMs {
		t.Fatalf("expected timeout %d, got %d", def.TimeoutMs, cfg.TimeoutMs)
	}
	if len(cfg.Lines) != len(def.Lines) {
		t.Fatalf("expected %d lines, got %d", len(def.Lines), len(cfg.Lines))
	}
}

func TestRunInitSkipsClaudeSettingsWhenMissing(t *testing.T) {
	// Use a temp dir as a fake home with no .claude directory.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Should not error even when ~/.claude/settings.json doesn't exist.
	err := cli.RunInit()
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}
}

func TestRunInitPatchesClaudeSettings(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create a pre-existing settings.json.
	claudeDir := filepath.Join(tmpHome, ".claude")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}

	existing := map[string]interface{}{
		"theme": "dark",
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	err := cli.RunInit()
	if err != nil {
		t.Fatalf("RunInit returned error: %v", err)
	}

	// Verify settings.json was patched.
	patched, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(patched, &settings); err != nil {
		t.Fatalf("patched settings.json is not valid JSON: %v", err)
	}

	// Original key should still be there.
	if settings["theme"] != "dark" {
		t.Fatal("expected existing 'theme' key to be preserved")
	}

	// statusLine should be added.
	sl, ok := settings["statusLine"]
	if !ok {
		t.Fatal("expected statusLine key in settings")
	}

	slMap, ok := sl.(map[string]interface{})
	if !ok {
		t.Fatal("expected statusLine to be a map")
	}

	cmd, ok := slMap["command"]
	if !ok || cmd == "" {
		t.Fatal("expected statusLine.command to be set")
	}
}

func TestDoctorReportsConfigStatus(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Doctor should not error even without config.
	err := cli.RunDoctor()
	if err != nil {
		t.Fatalf("RunDoctor returned error: %v", err)
	}
}

func TestDoctorAfterInit(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Init first, then doctor.
	if err := cli.RunInit(); err != nil {
		t.Fatalf("RunInit error: %v", err)
	}

	err := cli.RunDoctor()
	if err != nil {
		t.Fatalf("RunDoctor returned error: %v", err)
	}
}

func TestRunVersionDoesNotPanic(t *testing.T) {
	// Simply verify it doesn't panic.
	cli.RunVersion()
}
