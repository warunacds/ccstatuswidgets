package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunDoctor checks the health of the ccstatuswidgets installation
// and reports the status of each component.
func RunDoctor() error {
	fmt.Println("ccstatuswidgets doctor")
	fmt.Println()

	allGood := true

	// 1. Binary found (always true since we're running).
	printCheck(true, "ccw binary found")

	// 2. Config file exists and is valid JSON.
	configOK := checkConfig()
	if !configOK {
		allGood = false
	}

	// 3. Claude Code settings has statusLine configured.
	settingsOK := checkClaudeSettings()
	if !settingsOK {
		allGood = false
	}

	// 4. git available.
	gitOK := checkCommand("git")
	if !gitOK {
		allGood = false
	}

	// 5. python3 available.
	pythonOK := checkCommand("python3")
	if !pythonOK {
		allGood = false
	}

	fmt.Println()
	if allGood {
		fmt.Println("All checks passed!")
	} else {
		fmt.Println("Some checks failed. Run 'ccw init' to fix configuration issues.")
	}

	return nil
}

func checkConfig() bool {
	configDir := config.ConfigDir()
	configPath := filepath.Join(configDir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		printCheck(false, "config.json not found (run ccw init)")
		return false
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		printCheck(false, "config.json is not valid JSON")
		return false
	}

	printCheck(true, "config.json valid")
	return true
}

func checkClaudeSettings() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		printCheck(false, "Claude Code settings not configured (run ccw init)")
		return false
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		printCheck(false, "Claude Code settings not configured (run ccw init)")
		return false
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		printCheck(false, "Claude Code settings.json is not valid JSON")
		return false
	}

	sl, ok := settings["statusLine"]
	if !ok {
		printCheck(false, "Claude Code settings not configured (run ccw init)")
		return false
	}

	slMap, ok := sl.(map[string]interface{})
	if !ok {
		printCheck(false, "Claude Code statusLine config is malformed")
		return false
	}

	cmd, ok := slMap["command"]
	if !ok || cmd == "" {
		printCheck(false, "Claude Code statusLine command not set (run ccw init)")
		return false
	}

	printCheck(true, "Claude Code settings configured")
	return true
}

func checkCommand(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		printCheck(false, name+" not available")
		return false
	}
	printCheck(true, name+" available")
	return true
}

func printCheck(ok bool, msg string) {
	if ok {
		fmt.Printf("  \u2713 %s\n", msg)
	} else {
		fmt.Printf("  \u2717 %s\n", msg)
	}
}
