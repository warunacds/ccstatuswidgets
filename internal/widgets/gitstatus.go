package widgets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// GitStatusWidget displays working tree status (clean/dirty with file counts).
type GitStatusWidget struct{}

func (w *GitStatusWidget) Name() string {
	return "git-status"
}

func (w *GitStatusWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	dir := input.Workspace.CurrentDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, nil
		}
		dir = cwd
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.Output()
	if err != nil {
		// Not in a git repo — return nil.
		return nil, nil
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return &protocol.WidgetOutput{
			Text:  "\u2713",
			Color: "green",
		}, nil
	}

	var modified, added, deleted, untracked int
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 2 {
			continue
		}
		// git status --porcelain format: XY filename
		// X = index status, Y = working tree status
		xy := line[:2]

		switch {
		case xy == "??":
			untracked++
		case xy[0] == 'D' || xy[1] == 'D':
			deleted++
		case xy[0] == 'A':
			added++
		case xy[0] == 'M' || xy[1] == 'M':
			modified++
		case xy[0] == 'R':
			// Renames count as modified.
			modified++
		}
	}

	var parts []string
	if modified > 0 {
		parts = append(parts, fmt.Sprintf("%dM", modified))
	}
	if added > 0 {
		parts = append(parts, fmt.Sprintf("%dA", added))
	}
	if deleted > 0 {
		parts = append(parts, fmt.Sprintf("%dD", deleted))
	}
	if untracked > 0 {
		parts = append(parts, fmt.Sprintf("%dU", untracked))
	}

	text := "\u270e " + strings.Join(parts, " ")

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "yellow",
	}, nil
}
