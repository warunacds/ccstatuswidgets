package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type pomoState struct {
	Phase     string `json:"phase"`
	StartedAt int64  `json:"started_at"`
	WorkMins  int    `json:"work_mins"`
	BreakMins int    `json:"break_mins"`
}

// RunPomo handles the "pomo" subcommand. args should be the subcommand
// arguments (e.g., ["start"], ["stop"]). stateDir is the directory where
// pomodoro.json is stored.
func RunPomo(args []string, stateDir string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ccw pomo <start|stop|skip|status>")
	}

	statePath := filepath.Join(stateDir, "pomodoro.json")

	switch args[0] {
	case "start":
		return pomoStart(statePath)
	case "stop":
		return pomoStop(statePath)
	case "skip":
		return pomoSkip(statePath)
	case "status":
		return pomoStatus(statePath)
	default:
		return fmt.Errorf("unknown pomo command: %s", args[0])
	}
}

func pomoStart(statePath string) error {
	// Check if already running.
	if _, err := os.Stat(statePath); err == nil {
		return fmt.Errorf("pomodoro timer is already running (use 'ccw pomo stop' first)")
	}

	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	state := pomoState{
		Phase:     "work",
		StartedAt: time.Now().Unix(),
		WorkMins:  25,
		BreakMins: 5,
	}

	if err := writePomo(statePath, state); err != nil {
		return err
	}

	fmt.Println("🍅 Pomodoro started! 25 minutes of focused work.")
	return nil
}

func pomoStop(statePath string) error {
	err := os.Remove(statePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}
	fmt.Println("Pomodoro timer stopped.")
	return nil
}

func pomoSkip(statePath string) error {
	data, err := os.ReadFile(statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("no pomodoro timer running (use 'ccw pomo start' first)")
		}
		return fmt.Errorf("failed to read state: %w", err)
	}

	var state pomoState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("invalid state file: %w", err)
	}

	// Transition to the next phase.
	if state.Phase == "work" {
		state.Phase = "break"
		fmt.Printf("☕ Skipped to break! %d minutes.\n", state.BreakMins)
	} else {
		state.Phase = "work"
		fmt.Printf("🍅 Skipped to work! %d minutes.\n", state.WorkMins)
	}
	state.StartedAt = time.Now().Unix()

	return writePomo(statePath, state)
}

func pomoStatus(statePath string) error {
	data, err := os.ReadFile(statePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("No pomodoro timer running.")
			return nil
		}
		return fmt.Errorf("failed to read state: %w", err)
	}

	var state pomoState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("invalid state file: %w", err)
	}

	var durationMins int
	if state.Phase == "work" {
		durationMins = state.WorkMins
	} else {
		durationMins = state.BreakMins
	}

	startTime := time.Unix(state.StartedAt, 0)
	endTime := startTime.Add(time.Duration(durationMins) * time.Minute)
	remaining := time.Until(endTime)

	if remaining <= 0 {
		fmt.Printf("Phase: %s (expired — will auto-transition on next render)\n", state.Phase)
		return nil
	}

	mins := int(remaining.Minutes())
	secs := int(remaining.Seconds()) % 60
	fmt.Printf("Phase: %s | Remaining: %d:%02d\n", state.Phase, mins, secs)
	return nil
}

func writePomo(path string, state pomoState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}
