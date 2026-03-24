package widgets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// GitDiffWidget displays the count of staged files.
type GitDiffWidget struct{}

func (w *GitDiffWidget) Name() string {
	return "git-diff"
}

func (w *GitDiffWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	dir := input.Workspace.CurrentDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil
		}
		dir = cwd
	}

	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.Output()
	if err != nil {
		// Not in a git repo — return nil.
		return nil, nil
	}

	// Count non-empty lines.
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	if count == 0 {
		return nil, nil
	}

	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("staged: %d files", count),
		Color: "green",
	}, nil
}
