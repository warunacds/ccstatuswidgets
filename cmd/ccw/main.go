package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/cache"
	"github.com/warunacds/ccstatuswidgets/internal/cli"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/engine"
	"github.com/warunacds/ccstatuswidgets/internal/plugin"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
	"github.com/warunacds/ccstatuswidgets/internal/renderer"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
	"github.com/warunacds/ccstatuswidgets/internal/widgets"
)

func main() {
	// If a subcommand is provided, route to it.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := cli.RunInit(); err != nil {
				fmt.Fprintf(os.Stderr, "ccw init: %v\n", err)
				os.Exit(1)
			}
			return
		case "doctor":
			if err := cli.RunDoctor(); err != nil {
				fmt.Fprintf(os.Stderr, "ccw doctor: %v\n", err)
				os.Exit(1)
			}
			return
		case "preview":
			if err := cli.RunPreview(); err != nil {
				fmt.Fprintf(os.Stderr, "ccw preview: %v\n", err)
				os.Exit(1)
			}
			return
		case "version":
			cli.RunVersion()
			return
		case "plugin":
			pluginsDir := filepath.Join(config.ConfigDir(), "plugins")
			registry := widget.NewRegistry()
			widgets.RegisterAll(registry)
			if err := cli.RunPlugin(os.Args[2:], pluginsDir, registry.Names()); err != nil {
				fmt.Fprintf(os.Stderr, "ccw plugin: %v\n", err)
				os.Exit(1)
			}
			return
		case "pomo":
			if err := cli.RunPomo(os.Args[2:], config.ConfigDir()); err != nil {
				fmt.Fprintf(os.Stderr, "ccw pomo: %v\n", err)
				os.Exit(1)
			}
			return
		case "track":
			if err := cli.RunTrack(os.Args[2:], config.ConfigDir()); err != nil {
				fmt.Fprintf(os.Stderr, "ccw track: %v\n", err)
				os.Exit(1)
			}
			return
		case "hn":
			if err := cli.RunHN(""); err != nil {
				fmt.Fprintf(os.Stderr, "ccw hn: %v\n", err)
				os.Exit(1)
			}
			return
		default:
			printUsage()
			os.Exit(1)
		}
	}

	// No args: stdin pipeline mode (called by Claude Code).
	runPipeline()
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: ccw <command>\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  init      Initialize configuration and directories\n")
	fmt.Fprintf(os.Stderr, "  doctor    Check installation health\n")
	fmt.Fprintf(os.Stderr, "  preview   Show a sample status line\n")
	fmt.Fprintf(os.Stderr, "  version   Print version information\n")
	fmt.Fprintf(os.Stderr, "  plugin    Manage external plugins (add, list, remove, update)\n")
	fmt.Fprintf(os.Stderr, "  pomo      Pomodoro timer (start, stop, skip, status)\n")
	fmt.Fprintf(os.Stderr, "  track     Track a flight number (e.g., ccw track UL504)\n")
	fmt.Fprintf(os.Stderr, "  hn        Show top Hacker News stories\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "When called without arguments, reads JSON from stdin (Claude Code pipeline mode).\n")
}

func runPipeline() {
	// Read stdin with 1-second timeout.
	done := make(chan []byte, 1)
	go func() {
		data, _ := io.ReadAll(os.Stdin)
		done <- data
	}()

	var data []byte
	select {
	case data = <-done:
	case <-time.After(1 * time.Second):
		os.Exit(0)
	}

	if len(data) == 0 {
		os.Exit(0)
	}

	// Parse JSON input from Claude Code.
	var input protocol.StatusLineInput
	if err := json.Unmarshal(data, &input); err != nil {
		os.Exit(0)
	}

	// Load config (falls back to defaults if no config file exists).
	configDir := config.ConfigDir()
	cfg, err := config.Load(filepath.Join(configDir, "config.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ccw: config error: %v\n", err)
		os.Exit(1)
	}

	// Create widget registry and register all built-in widgets.
	registry := widget.NewRegistry()
	widgets.RegisterAll(registry)

	// Discover and register external plugins.
	externalWidgets, _ := plugin.DiscoverPlugins(filepath.Join(configDir, "plugins"))
	for _, ew := range externalWidgets {
		registry.Register(ew)
	}

	// Create file-based cache for widget fallback.
	cacheDir := filepath.Join(configDir, "cache")
	cacheInstance := cache.New(cacheDir)

	// Create engine with configured timeout.
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	eng := engine.New(registry, cacheInstance, timeout)

	// Run all widgets concurrently.
	results := eng.Run(&input, cfg)

	// Render and print output.
	output := renderer.Render(results)
	fmt.Print(output)
}
