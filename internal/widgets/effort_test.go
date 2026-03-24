package widgets

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestEffortWidget_Name(t *testing.T) {
	w := &EffortWidget{}
	if w.Name() != "effort" {
		t.Errorf("expected name %q, got %q", "effort", w.Name())
	}
}

func TestEffortWidget_ReturnsEffortFromSettingsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"effortLevel":"high"}`), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	w := &EffortWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"settings_dir": tmpDir,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "(high)" {
		t.Errorf("expected text %q, got %q", "(high)", out.Text)
	}
	if out.Color != "dim" {
		t.Errorf("expected color %q, got %q", "dim", out.Color)
	}
}

func TestEffortWidget_LocalSettingsOverridesGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	localSettingsPath := filepath.Join(tmpDir, "settings.local.json")
	if err := os.WriteFile(settingsPath, []byte(`{"effortLevel":"low"}`), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}
	if err := os.WriteFile(localSettingsPath, []byte(`{"effortLevel":"max"}`), 0644); err != nil {
		t.Fatalf("failed to write settings.local.json: %v", err)
	}

	w := &EffortWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"settings_dir": tmpDir,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Text != "(max)" {
		t.Errorf("expected text %q, got %q", "(max)", out.Text)
	}
}

func TestEffortWidget_ReturnsNilWhenNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	w := &EffortWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"settings_dir": tmpDir,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestEffortWidget_ReturnsNilWhenFieldMissing(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"someOtherField":"value"}`), 0644); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	w := &EffortWidget{}
	input := &protocol.StatusLineInput{}
	cfg := map[string]interface{}{
		"settings_dir": tmpDir,
	}

	out, err := w.Render(input, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}
