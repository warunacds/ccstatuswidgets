package widgets

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// EffortWidget displays the configured effort level from Claude settings.
type EffortWidget struct{}

func (w *EffortWidget) Name() string {
	return "effort"
}

func (w *EffortWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	settingsDir := w.settingsDir(cfg)

	// Read settings.json first, then overlay settings.local.json.
	effort := ""
	if e := readEffortLevel(filepath.Join(settingsDir, "settings.json")); e != "" {
		effort = e
	}
	if e := readEffortLevel(filepath.Join(settingsDir, "settings.local.json")); e != "" {
		effort = e
	}

	if effort == "" {
		return nil, nil
	}

	return &protocol.WidgetOutput{
		Text:  "(" + effort + ")",
		Color: "dim",
	}, nil
}

func (w *EffortWidget) settingsDir(cfg map[string]interface{}) string {
	if cfg != nil {
		if dir, ok := cfg["settings_dir"].(string); ok && dir != "" {
			return dir
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude")
}

// readEffortLevel reads a JSON file and extracts the effortLevel string field.
func readEffortLevel(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return ""
	}
	level, ok := m["effortLevel"].(string)
	if !ok {
		return ""
	}
	return level
}
