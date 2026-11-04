package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readStatusLine(t *testing.T, path string) (map[string]interface{}, map[string]interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("parse settings: %v", err)
	}
	sl, _ := settings["statusLine"].(map[string]interface{})
	if sl == nil {
		t.Fatalf("statusLine missing from %s", string(data))
	}
	return settings, sl
}

func TestPatchSettingsFile_AddsStatusLineWithRefreshInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"model":"opus"}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := patchSettingsFile(path, "/usr/local/bin/ccw"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings, sl := readStatusLine(t, path)
	if settings["model"] != "opus" {
		t.Errorf("expected other settings preserved, got %v", settings["model"])
	}
	if sl["type"] != "command" || sl["command"] != "/usr/local/bin/ccw" {
		t.Errorf("unexpected statusLine: %v", sl)
	}
	if sl["refreshInterval"] != float64(defaultRefreshInterval) {
		t.Errorf("expected refreshInterval %d, got %v", defaultRefreshInterval, sl["refreshInterval"])
	}
}

func TestPatchSettingsFile_PreservesExistingStatusLineKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	existing := `{"statusLine":{"type":"command","command":"/old/ccw","refreshInterval":2,"padding":1}}`
	if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := patchSettingsFile(path, "/new/ccw"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, sl := readStatusLine(t, path)
	if sl["command"] != "/new/ccw" {
		t.Errorf("expected command updated, got %v", sl["command"])
	}
	if sl["refreshInterval"] != float64(2) {
		t.Errorf("expected user refreshInterval 2 kept, got %v", sl["refreshInterval"])
	}
	if sl["padding"] != float64(1) {
		t.Errorf("expected padding kept, got %v", sl["padding"])
	}
}

func TestPatchSettingsFile_RejectsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{not json`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := patchSettingsFile(path, "/usr/local/bin/ccw"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
