package widgets

import (
	"os"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// DirectoryWidget displays the basename of the current working directory.
type DirectoryWidget struct{}

func (w *DirectoryWidget) Name() string {
	return "directory"
}

func (w *DirectoryWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	dir := input.Workspace.CurrentDir
	if dir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dir = cwd
	}

	return &protocol.WidgetOutput{
		Text:  filepath.Base(dir),
		Color: "cyan",
	}, nil
}
