package widgets

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// MemoryWidget displays the memory usage of the parent process.
type MemoryWidget struct{}

func (w *MemoryWidget) Name() string {
	return "memory"
}

func (w *MemoryWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	pid := os.Getppid()
	if cfg != nil {
		if v, ok := cfg["pid"]; ok {
			if p, ok := v.(int); ok {
				pid = p
			}
		}
	}

	cmd := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	rssStr := strings.TrimSpace(string(out))
	if rssStr == "" {
		return nil, nil
	}

	rssKB, err := strconv.Atoi(rssStr)
	if err != nil {
		return nil, nil
	}

	mb := rssKB / 1024
	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("%dMB", mb),
		Color: "dim",
	}, nil
}
