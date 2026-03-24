package protocol

// StatusLineInput is the JSON payload Claude Code pipes to the status line binary via stdin.
type StatusLineInput struct {
	Model         ModelInfo     `json:"model"`
	Workspace     WorkspaceInfo `json:"workspace"`
	ContextWindow ContextInfo   `json:"context_window"`
	RateLimits    *RateLimits   `json:"rate_limits"`
	Cost          CostInfo      `json:"cost"`
	SessionID     string        `json:"session_id"`
	Version       string        `json:"version"`
}

type ModelInfo struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type WorkspaceInfo struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type ContextInfo struct {
	UsedPercentage      float64 `json:"used_percentage"`
	RemainingPercentage float64 `json:"remaining_percentage"`
	TotalInputTokens    int     `json:"total_input_tokens"`
	TotalOutputTokens   int     `json:"total_output_tokens"`
	ContextWindowSize   int     `json:"context_window_size"`
}

type RateLimits struct {
	FiveHour *RateLimit `json:"five_hour"`
	SevenDay *RateLimit `json:"seven_day"`
}

type RateLimit struct {
	UsedPercentage float64 `json:"used_percentage"`
	ResetsAt       int64   `json:"resets_at"`
}

type CostInfo struct {
	TotalCostUSD      float64 `json:"total_cost_usd"`
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

// WidgetOutput is what each widget returns for rendering.
type WidgetOutput struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}
