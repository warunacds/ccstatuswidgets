package cli

import (
	"fmt"
	"path/filepath"

	"github.com/warunacds/ccstatuswidgets/internal/config"
)

// RunTrack sets or removes a flight number in the config for real-time tracking.
//
//	ccw track UL504  — sets widgets.flight.flight to "UL504"
//	ccw track stop   — removes the flight key from widgets.flight
func RunTrack(args []string, configDir string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: ccw track <flight_number|stop>")
	}

	configPath := filepath.Join(configDir, "config.json")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Widgets == nil {
		cfg.Widgets = make(map[string]map[string]interface{})
	}
	if cfg.Widgets["flight"] == nil {
		cfg.Widgets["flight"] = make(map[string]interface{})
	}

	sub := args[0]

	if sub == "stop" {
		delete(cfg.Widgets["flight"], "flight")
		if err := config.Save(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Flight tracking stopped.")
		return nil
	}

	cfg.Widgets["flight"]["flight"] = sub
	if err := config.Save(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Now tracking flight %s.\n", sub)
	return nil
}
