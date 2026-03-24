package widgets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

func TestPomodoroWidget_Name(t *testing.T) {
	w := &PomodoroWidget{}
	if w.Name() != "pomodoro" {
		t.Errorf("expected name %q, got %q", "pomodoro", w.Name())
	}
}

func TestPomodoroWidget_ReturnsNilWhenNoStateFile(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()
	cfg := map[string]interface{}{
		"state_dir": stateDir,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != nil {
		t.Errorf("expected nil output, got %+v", out)
	}
}

func TestPomodoroWidget_WorkPhaseCountdownRed(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()

	// Start work phase 6 minutes and 18 seconds ago (25 min total → 18:42 remaining).
	startedAt := time.Now().Add(-6*time.Minute - 18*time.Second)
	state := pomodoroState{
		Phase:     "work",
		StartedAt: startedAt.Unix(),
		WorkMins:  25,
		BreakMins: 5,
	}
	writeState(t, stateDir, state)

	cfg := map[string]interface{}{
		"state_dir": stateDir,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Color != "red" {
		t.Errorf("expected color %q during work phase, got %q", "red", out.Color)
	}
	// The text should start with the tomato emoji.
	if len(out.Text) < 2 || out.Text[:4] != "🍅" {
		t.Errorf("expected text to start with tomato emoji, got %q", out.Text)
	}
	// Should contain a time like "18:4" (we allow slight variance from timing).
	if len(out.Text) < 5 {
		t.Errorf("expected text to contain a countdown, got %q", out.Text)
	}
}

func TestPomodoroWidget_BreakPhaseCountdownGreen(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()

	// Start break phase 1 minute and 39 seconds ago (5 min total → 3:21 remaining).
	startedAt := time.Now().Add(-1*time.Minute - 39*time.Second)
	state := pomodoroState{
		Phase:     "break",
		StartedAt: startedAt.Unix(),
		WorkMins:  25,
		BreakMins: 5,
	}
	writeState(t, stateDir, state)

	cfg := map[string]interface{}{
		"state_dir": stateDir,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	if out.Color != "green" {
		t.Errorf("expected color %q during break phase, got %q", "green", out.Color)
	}
	// The text should start with the coffee emoji.
	if len(out.Text) < 2 || out.Text[:3] != "☕" {
		t.Errorf("expected text to start with coffee emoji, got %q", out.Text)
	}
}

func TestPomodoroWidget_AutoTransitionsWorkToBreak(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()

	// Work phase started 26 minutes ago (25 min work → expired 1 min ago).
	startedAt := time.Now().Add(-26 * time.Minute)
	state := pomodoroState{
		Phase:     "work",
		StartedAt: startedAt.Unix(),
		WorkMins:  25,
		BreakMins: 5,
	}
	writeState(t, stateDir, state)

	cfg := map[string]interface{}{
		"state_dir": stateDir,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output after auto-transition")
	}
	// Should now be in break phase.
	if out.Color != "green" {
		t.Errorf("expected green (break phase) after work→break transition, got %q", out.Color)
	}

	// Verify the state file was updated.
	updated := readState(t, stateDir)
	if updated.Phase != "break" {
		t.Errorf("expected state file phase to be %q, got %q", "break", updated.Phase)
	}
}

func TestPomodoroWidget_AutoTransitionsBreakToWork(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()

	// Break phase started 6 minutes ago (5 min break → expired 1 min ago).
	startedAt := time.Now().Add(-6 * time.Minute)
	state := pomodoroState{
		Phase:     "break",
		StartedAt: startedAt.Unix(),
		WorkMins:  25,
		BreakMins: 5,
	}
	writeState(t, stateDir, state)

	cfg := map[string]interface{}{
		"state_dir": stateDir,
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output after auto-transition")
	}
	// Should now be in work phase.
	if out.Color != "red" {
		t.Errorf("expected red (work phase) after break→work transition, got %q", out.Color)
	}

	// Verify the state file was updated.
	updated := readState(t, stateDir)
	if updated.Phase != "work" {
		t.Errorf("expected state file phase to be %q, got %q", "work", updated.Phase)
	}
}

func TestPomodoroWidget_RespectsConfigDurations(t *testing.T) {
	w := &PomodoroWidget{}
	stateDir := t.TempDir()

	// Work phase started 10 minutes ago. With work_mins=15, should have 5 min left.
	startedAt := time.Now().Add(-10 * time.Minute)
	state := pomodoroState{
		Phase:     "work",
		StartedAt: startedAt.Unix(),
		WorkMins:  15,
		BreakMins: 3,
	}
	writeState(t, stateDir, state)

	cfg := map[string]interface{}{
		"state_dir": stateDir,
		"work_mins": float64(15),
		"break_mins": float64(3),
	}

	out, err := w.Render(&protocol.StatusLineInput{}, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil output")
	}
	// Should still be in work phase (10 < 15 mins).
	if out.Color != "red" {
		t.Errorf("expected red (work still active with 15 min config), got %q", out.Color)
	}

	// With default 25 min, 10 min ago would also still be active, but with 8 min work,
	// it should have transitioned. Let's verify that too.
	stateDir2 := t.TempDir()
	state2 := pomodoroState{
		Phase:     "work",
		StartedAt: startedAt.Unix(),
		WorkMins:  8,
		BreakMins: 3,
	}
	writeState(t, stateDir2, state2)

	cfg2 := map[string]interface{}{
		"state_dir":  stateDir2,
		"work_mins":  float64(8),
		"break_mins": float64(3),
	}

	out2, err := w.Render(&protocol.StatusLineInput{}, cfg2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out2 == nil {
		t.Fatal("expected non-nil output after transition")
	}
	// 10 min > 8 min work → should have transitioned to break.
	if out2.Color != "green" {
		t.Errorf("expected green (break) after 8 min work expired, got %q", out2.Color)
	}
}

// Helper to write a pomodoro state file.
func writeState(t *testing.T, stateDir string, state pomodoroState) {
	t.Helper()
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("failed to marshal state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "pomodoro.json"), data, 0644); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}
}

// Helper to read a pomodoro state file.
func readState(t *testing.T, stateDir string) pomodoroState {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(stateDir, "pomodoro.json"))
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}
	var state pomodoroState
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}
	return state
}
