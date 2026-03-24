package widgets

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// pomodoroState is the JSON structure stored in the pomodoro state file.
type pomodoroState struct {
	Phase     string `json:"phase"`
	StartedAt int64  `json:"started_at"`
	WorkMins  int    `json:"work_mins"`
	BreakMins int    `json:"break_mins"`
}

// PomodoroWidget displays a pomodoro timer countdown.
type PomodoroWidget struct{}

func (w *PomodoroWidget) Name() string {
	return "pomodoro"
}

func (w *PomodoroWidget) Render(_ *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	stateDir := pomodoroStateDir(cfg)
	statePath := filepath.Join(stateDir, "pomodoro.json")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("pomodoro: read state: %w", err)
	}

	var state pomodoroState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("pomodoro: parse state: %w", err)
	}

	// Apply config overrides for durations.
	workMins := state.WorkMins
	breakMins := state.BreakMins
	if v, ok := cfg["work_mins"]; ok {
		if f, ok := v.(float64); ok {
			workMins = int(f)
		}
	}
	if v, ok := cfg["break_mins"]; ok {
		if f, ok := v.(float64); ok {
			breakMins = int(f)
		}
	}
	if workMins <= 0 {
		workMins = 25
	}
	if breakMins <= 0 {
		breakMins = 5
	}

	// Determine duration for the current phase.
	var durationMins int
	if state.Phase == "work" {
		durationMins = workMins
	} else {
		durationMins = breakMins
	}

	startTime := time.Unix(state.StartedAt, 0)
	endTime := startTime.Add(time.Duration(durationMins) * time.Minute)
	remaining := time.Until(endTime)

	// Auto-transition if the timer has expired.
	if remaining <= 0 {
		if state.Phase == "work" {
			state.Phase = "break"
			durationMins = breakMins
		} else {
			state.Phase = "work"
			durationMins = workMins
		}
		state.StartedAt = time.Now().Unix()
		state.WorkMins = workMins
		state.BreakMins = breakMins

		if err := writePomoState(statePath, state); err != nil {
			return nil, fmt.Errorf("pomodoro: write state: %w", err)
		}

		// Recalculate remaining time for the new phase.
		startTime = time.Unix(state.StartedAt, 0)
		endTime = startTime.Add(time.Duration(durationMins) * time.Minute)
		remaining = time.Until(endTime)
	}

	// Format the countdown.
	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60

	var emoji, color string
	if state.Phase == "work" {
		emoji = "🍅"
		color = "red"
	} else {
		emoji = "☕"
		color = "green"
	}

	text := fmt.Sprintf("%s %d:%02d", emoji, mins, secs)

	return &protocol.WidgetOutput{
		Text:  text,
		Color: color,
	}, nil
}

// pomodoroStateDir returns the directory for the pomodoro state file.
// It checks cfg["state_dir"] first, then falls back to ~/.ccstatuswidgets.
func pomodoroStateDir(cfg map[string]interface{}) string {
	if cfg != nil {
		if v, ok := cfg["state_dir"]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".ccstatuswidgets")
	}
	return filepath.Join(home, ".ccstatuswidgets")
}

// writePomoState writes the pomodoro state to the given path.
func writePomoState(path string, state pomodoroState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
