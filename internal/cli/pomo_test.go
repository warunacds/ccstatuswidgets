package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/warunacds/ccstatuswidgets/internal/cli"
)

type pomoState struct {
	Phase     string `json:"phase"`
	StartedAt int64  `json:"started_at"`
	WorkMins  int    `json:"work_mins"`
	BreakMins int    `json:"break_mins"`
}

func TestRunPomo_StartCreatesWorkPhase(t *testing.T) {
	stateDir := t.TempDir()

	err := cli.RunPomo([]string{"start"}, stateDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	statePath := filepath.Join(stateDir, "pomodoro.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("expected pomodoro.json to exist: %v", err)
	}

	var state pomoState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("invalid JSON in pomodoro.json: %v", err)
	}

	if state.Phase != "work" {
		t.Errorf("expected phase %q, got %q", "work", state.Phase)
	}
	if state.StartedAt == 0 {
		t.Error("expected started_at to be set")
	}
	if state.WorkMins != 25 {
		t.Errorf("expected work_mins %d, got %d", 25, state.WorkMins)
	}
	if state.BreakMins != 5 {
		t.Errorf("expected break_mins %d, got %d", 5, state.BreakMins)
	}
}

func TestRunPomo_StopRemovesStateFile(t *testing.T) {
	stateDir := t.TempDir()

	// Start a timer first.
	if err := cli.RunPomo([]string{"start"}, stateDir); err != nil {
		t.Fatalf("start error: %v", err)
	}

	// Stop the timer.
	if err := cli.RunPomo([]string{"stop"}, stateDir); err != nil {
		t.Fatalf("stop error: %v", err)
	}

	statePath := filepath.Join(stateDir, "pomodoro.json")
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected pomodoro.json to be removed after stop")
	}
}

func TestRunPomo_SkipTransitionsWorkToBreak(t *testing.T) {
	stateDir := t.TempDir()

	// Start a work timer.
	if err := cli.RunPomo([]string{"start"}, stateDir); err != nil {
		t.Fatalf("start error: %v", err)
	}

	// Skip to break.
	if err := cli.RunPomo([]string{"skip"}, stateDir); err != nil {
		t.Fatalf("skip error: %v", err)
	}

	statePath := filepath.Join(stateDir, "pomodoro.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("expected pomodoro.json to exist: %v", err)
	}

	var state pomoState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if state.Phase != "break" {
		t.Errorf("expected phase %q after skip from work, got %q", "break", state.Phase)
	}
}

func TestRunPomo_SkipTransitionsBreakToWork(t *testing.T) {
	stateDir := t.TempDir()

	// Start and skip to break.
	if err := cli.RunPomo([]string{"start"}, stateDir); err != nil {
		t.Fatalf("start error: %v", err)
	}
	if err := cli.RunPomo([]string{"skip"}, stateDir); err != nil {
		t.Fatalf("skip error: %v", err)
	}

	// Skip again to go back to work.
	if err := cli.RunPomo([]string{"skip"}, stateDir); err != nil {
		t.Fatalf("skip error: %v", err)
	}

	statePath := filepath.Join(stateDir, "pomodoro.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("expected pomodoro.json to exist: %v", err)
	}

	var state pomoState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if state.Phase != "work" {
		t.Errorf("expected phase %q after skip from break, got %q", "work", state.Phase)
	}
}

func TestRunPomo_StatusPrintsPhaseAndTime(t *testing.T) {
	stateDir := t.TempDir()

	// Start a timer.
	if err := cli.RunPomo([]string{"start"}, stateDir); err != nil {
		t.Fatalf("start error: %v", err)
	}

	// Status should not error.
	err := cli.RunPomo([]string{"status"}, stateDir)
	if err != nil {
		t.Fatalf("status error: %v", err)
	}
}

func TestRunPomo_StartWhenAlreadyRunning(t *testing.T) {
	stateDir := t.TempDir()

	// Start a timer.
	if err := cli.RunPomo([]string{"start"}, stateDir); err != nil {
		t.Fatalf("start error: %v", err)
	}

	// Start again — should return an error (warning).
	err := cli.RunPomo([]string{"start"}, stateDir)
	if err == nil {
		t.Fatal("expected error when starting an already running timer")
	}
}
