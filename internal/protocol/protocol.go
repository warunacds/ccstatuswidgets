package protocol

// StatusLineInput is the JSON payload Claude Code pipes to the status line binary via stdin.
type StatusLineInput struct {
	Model         ModelInfo     `json:"model"`
	Workspace     WorkspaceInfo `json:"workspace"`
	ContextWindow ContextInfo   `json:"context_window"`
	RateLimits    *RateLimits   `json:"rate_limits"`
	Cost          CostInfo      `json:"cost"`
	SessionID     string        `json:"session_id"`
	SessionName   string        `json:"session_name"` // absent unless named via --name, /rename, or AI title
	Version       string        `json:"version"`
	Effort        *EffortInfo   `json:"effort"` // absent when the model has no effort parameter
	FastMode      bool          `json:"fast_mode"`
	PromptCache   *PromptCache  `json:"prompt_cache"` // absent until the first API response
	PR            *PRInfo       `json:"pr"`           // absent unless an open PR/MR exists for the branch
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
	TotalDurationMs   int64   `json:"total_duration_ms"` // wall-clock session time, accumulates across resumes
	TotalLinesAdded   int     `json:"total_lines_added"`
	TotalLinesRemoved int     `json:"total_lines_removed"`
}

// EffortInfo is the reasoning effort level for the session.
type EffortInfo struct {
	Level string `json:"level"` // low, medium, high, xhigh, max
}

// PromptCache summarizes the main conversation's prompt cache usage.
type PromptCache struct {
	Warm            bool    `json:"warm"`
	CachingObserved bool    `json:"caching_observed"`
	TTL             string  `json:"ttl"` // "5m" or "1h"
	ExpiresAt       int64   `json:"expires_at"`
	Requests        int     `json:"requests"`
	Misses          int     `json:"misses"`
	HitRatio        float64 `json:"hit_ratio"` // 0..1, null (decoded as 0) until any input tokens are counted
}

// PRInfo describes the open pull request (or GitLab merge request) for the current branch.
type PRInfo struct {
	Number      int    `json:"number"`
	URL         string `json:"url"`
	ReviewState string `json:"review_state"` // approved, pending, changes_requested, draft; may be absent
	Kind        string `json:"kind"`         // "mr" for a GitLab merge request, absent for GitHub
}

// WidgetOutput is what each widget returns for rendering.
type WidgetOutput struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}
