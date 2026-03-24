package config

// Default returns the default configuration for ccstatuswidgets.
func Default() *Config {
	return &Config{
		TimeoutMs: 500,
		Separator: " ",
		Powerline: false,
		Lines: []LineConfig{
			{Widgets: []string{"model", "effort", "directory", "git-branch", "context-bar", "tokens", "session-time", "usage-5h", "usage-7d"}},
			{Widgets: []string{"lines-changed", "git-status", "cost", "memory"}},
		},
		Widgets: map[string]map[string]interface{}{
			"context-bar": {
				"bar_length":      float64(10),
				"show_percentage": true,
			},
			"usage-5h": {
				"bar_length":      float64(10),
				"show_percentage": true,
				"show_pace":       true,
			},
			"usage-7d": {
				"bar_length":      float64(10),
				"show_percentage": true,
				"show_pace":       true,
			},
			"cost": {
				"detect_max_plan": true,
			},
		},
	}
}
