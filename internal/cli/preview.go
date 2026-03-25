package cli

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/warunacds/ccstatuswidgets/internal/cache"
	"github.com/warunacds/ccstatuswidgets/internal/config"
	"github.com/warunacds/ccstatuswidgets/internal/engine"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
	"github.com/warunacds/ccstatuswidgets/internal/renderer"
	"github.com/warunacds/ccstatuswidgets/internal/widget"
	"github.com/warunacds/ccstatuswidgets/internal/widgets"
)

// RunPreview creates a sample StatusLineInput with realistic data
// and runs it through the full pipeline, printing the rendered output.
func RunPreview() error {
	input := &protocol.StatusLineInput{
		Model:     protocol.ModelInfo{DisplayName: "Opus 4.6"},
		Workspace: protocol.WorkspaceInfo{CurrentDir: "/Users/demo/myproject"},
		ContextWindow: protocol.ContextInfo{
			UsedPercentage:      35,
			RemainingPercentage: 65,
		},
		RateLimits: &protocol.RateLimits{
			FiveHour: &protocol.RateLimit{
				UsedPercentage: 22,
				ResetsAt:       time.Now().Add(2*time.Hour + 30*time.Minute).Unix(),
			},
			SevenDay: &protocol.RateLimit{
				UsedPercentage: 8,
				ResetsAt:       time.Now().Add(5 * 24 * time.Hour).Unix(),
			},
		},
		Cost: protocol.CostInfo{
			TotalCostUSD:      1.47,
			TotalLinesAdded:   342,
			TotalLinesRemoved: 87,
		},
	}

	// Load config (falls back to defaults if no config file exists).
	configDir := config.ConfigDir()
	cfg, err := config.Load(filepath.Join(configDir, "config.json"))
	if err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	// Create widget registry and register all built-in widgets.
	registry := widget.NewRegistry()
	widgets.RegisterAll(registry)

	// Create file-based cache for widget fallback.
	cacheDir := filepath.Join(configDir, "cache")
	cacheInstance := cache.New(cacheDir)

	// Create engine with configured timeout.
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	eng := engine.New(registry, cacheInstance, timeout)

	// Run all widgets concurrently.
	results := eng.Run(input, cfg)

	// Render and print output.
	output := renderer.Render(results, cfg)

	fmt.Println("ccstatuswidgets preview:")
	fmt.Println()
	fmt.Println(output)

	return nil
}
