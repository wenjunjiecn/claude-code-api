// Package models defines Claude-specific types.
package models

// ClaudeMessage represents a message from Claude CLI JSONL output.
type ClaudeMessage struct {
	Type       string                 `json:"type"`
	Subtype    string                 `json:"subtype,omitempty"`
	Message    *ClaudeMessageContent  `json:"message,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	Model      string                 `json:"model,omitempty"`
	CWD        string                 `json:"cwd,omitempty"`
	Tools      []string               `json:"tools,omitempty"`
	Result     string                 `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Usage      map[string]interface{} `json:"usage,omitempty"`
	CostUSD    float64                `json:"cost_usd,omitempty"`
	DurationMs int                    `json:"duration_ms,omitempty"`
}

// ClaudeMessageContent represents the content of a Claude message.
type ClaudeMessageContent struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}
