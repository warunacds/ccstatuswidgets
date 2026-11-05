package widgets

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// PlanWidget displays the Claude subscription plan (Free, Pro, Max, Ultra, ...).
// Claude Code doesn't send this on stdin; it's read from oauthAccount.planDisplayName
// in ~/.claude.json, which is absent for API key, Bedrock and Vertex users.
type PlanWidget struct{}

func (w *PlanWidget) Name() string {
	return "plan"
}

func (w *PlanWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	plan := readPlanDisplayName(w.statePath(cfg))
	if plan == "" {
		return nil, nil
	}
	return &protocol.WidgetOutput{
		Text:  plan,
		Color: "magenta",
	}, nil
}

func (w *PlanWidget) statePath(cfg map[string]interface{}) string {
	if cfg != nil {
		if p, ok := cfg["state_file"].(string); ok && p != "" {
			return p
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude.json")
}

// readPlanDisplayName extracts oauthAccount.planDisplayName from Claude Code's state file.
func readPlanDisplayName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var state struct {
		OAuthAccount struct {
			PlanDisplayName string `json:"planDisplayName"`
		} `json:"oauthAccount"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return ""
	}
	return state.OAuthAccount.PlanDisplayName
}
