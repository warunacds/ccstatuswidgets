package widgets

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

const nowPlayingMaxLen = 30

// NowPlayingWidget displays the currently playing track.
// On macOS it tries Spotify first, then Music.app via AppleScript.
// On Linux it uses playerctl.
// For testability, cfg["command"] and cfg["args"] override the default command.
type NowPlayingWidget struct{}

func (w *NowPlayingWidget) Name() string {
	return "now-playing"
}

func (w *NowPlayingWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	text := w.getCurrentTrack(cfg)
	if text == "" {
		return nil, nil
	}

	if len(text) > nowPlayingMaxLen {
		text = text[:nowPlayingMaxLen] + "..."
	}

	return &protocol.WidgetOutput{
		Text:  fmt.Sprintf("\u266a %s", text),
		Color: "magenta",
	}, nil
}

func (w *NowPlayingWidget) getCurrentTrack(cfg map[string]interface{}) string {
	// Allow command override for testing
	if cmd, ok := cfg["command"]; ok {
		if cmdStr, ok := cmd.(string); ok {
			args := ""
			if a, ok := cfg["args"]; ok {
				if aStr, ok := a.(string); ok {
					args = aStr
				}
			}
			return w.runCommand(cmdStr, args)
		}
	}

	switch runtime.GOOS {
	case "darwin":
		return w.macOSTrack()
	case "linux":
		return w.linuxTrack()
	default:
		return ""
	}
}

func (w *NowPlayingWidget) macOSTrack() string {
	// Try Spotify first (only if running)
	script := `tell application "System Events"
	if exists (processes whose name is "Spotify") then
		tell application "Spotify"
			if player state is playing then
				return artist of current track & " - " & name of current track
			end if
		end tell
	end if
end tell
return ""`
	text := w.runCommand("osascript", "-e", script)
	if text != "" {
		return text
	}

	// Fall back to Music.app (only if running)
	script = `tell application "System Events"
	if exists (processes whose name is "Music") then
		tell application "Music"
			if player state is playing then
				return artist of current track & " - " & name of current track
			end if
		end tell
	end if
end tell
return ""`
	return w.runCommand("osascript", "-e", script)
}

func (w *NowPlayingWidget) linuxTrack() string {
	return w.runCommand("playerctl", "metadata", "--format", "{{artist}} - {{title}}")
}

func (w *NowPlayingWidget) runCommand(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
