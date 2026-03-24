package widgets

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// SessionTimeWidget displays the elapsed time since the parent process (Claude Code) started.
type SessionTimeWidget struct{}

func (w *SessionTimeWidget) Name() string {
	return "session-time"
}

func (w *SessionTimeWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	var startTime time.Time

	// Check for a config override (used in tests).
	if cfg != nil {
		if v, ok := cfg["start_time"]; ok {
			ts, ok := v.(float64)
			if !ok {
				return nil, nil
			}
			startTime = time.Unix(int64(ts), 0)
		}
	}

	// If no override, read the parent process start time via ps.
	if startTime.IsZero() {
		pid := os.Getppid()
		t, err := getProcessStartTime(pid)
		if err != nil {
			return nil, nil
		}
		startTime = t
	}

	elapsed := time.Since(startTime)
	totalMinutes := int(elapsed.Minutes())
	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	var text string
	if hours == 0 {
		text = fmt.Sprintf("⏱ %dm", minutes)
	} else if minutes == 0 {
		text = fmt.Sprintf("⏱ %dh", hours)
	} else {
		text = fmt.Sprintf("⏱ %dh%dm", hours, minutes)
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "dim",
	}, nil
}

// getProcessStartTime reads the start time of a process by PID using the ps command.
func getProcessStartTime(pid int) (time.Time, error) {
	cmd := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	lstart := strings.TrimSpace(string(out))
	if lstart == "" {
		return time.Time{}, fmt.Errorf("empty lstart for pid %d", pid)
	}

	// macOS/Linux ps lstart format: "Mon Jan  2 15:04:05 2006"
	t, err := time.Parse("Mon Jan  2 15:04:05 2006", lstart)
	if err != nil {
		// Try single-digit day without extra space.
		t, err = time.Parse("Mon Jan 2 15:04:05 2006", lstart)
		if err != nil {
			return time.Time{}, err
		}
	}

	return t, nil
}
