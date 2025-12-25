// Package models defines Claude-specific types.
package models

// ClaudeModel represents available Claude models.
type ClaudeModel string

const (
	ClaudeOpus4    ClaudeModel = "claude-opus-4-20250514"
	ClaudeSonnet4  ClaudeModel = "claude-sonnet-4-20250514"
	ClaudeSonnet37 ClaudeModel = "claude-3-7-sonnet-20250219"
	ClaudeHaiku35  ClaudeModel = "claude-3-5-haiku-20241022"
)

// AllModels returns all available Claude models.
func AllModels() []ClaudeModel {
	return []ClaudeModel{
		ClaudeOpus4,
		ClaudeSonnet4,
		ClaudeSonnet37,
		ClaudeHaiku35,
	}
}

// ClaudeModelInfo contains model metadata.
type ClaudeModelInfo struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	MaxTokens         int     `json:"max_tokens"`
	InputCostPer1K    float64 `json:"input_cost_per_1k"`
	OutputCostPer1K   float64 `json:"output_cost_per_1k"`
	SupportsStreaming bool    `json:"supports_streaming"`
	SupportsTools     bool    `json:"supports_tools"`
}

// GetModelInfo returns metadata for a model.
func GetModelInfo(modelID string) ClaudeModelInfo {
	info := map[ClaudeModel]ClaudeModelInfo{
		ClaudeOpus4: {
			ID:                string(ClaudeOpus4),
			Name:              "Claude Opus 4",
			Description:       "Most powerful Claude model for complex reasoning",
			MaxTokens:         500000,
			InputCostPer1K:    15.0,
			OutputCostPer1K:   75.0,
			SupportsStreaming: true,
			SupportsTools:     true,
		},
		ClaudeSonnet4: {
			ID:                string(ClaudeSonnet4),
			Name:              "Claude Sonnet 4",
			Description:       "Latest Sonnet model with enhanced capabilities",
			MaxTokens:         500000,
			InputCostPer1K:    3.0,
			OutputCostPer1K:   15.0,
			SupportsStreaming: true,
			SupportsTools:     true,
		},
		ClaudeSonnet37: {
			ID:                string(ClaudeSonnet37),
			Name:              "Claude Sonnet 3.7",
			Description:       "Advanced Sonnet model for complex tasks",
			MaxTokens:         200000,
			InputCostPer1K:    3.0,
			OutputCostPer1K:   15.0,
			SupportsStreaming: true,
			SupportsTools:     true,
		},
		ClaudeHaiku35: {
			ID:                string(ClaudeHaiku35),
			Name:              "Claude Haiku 3.5",
			Description:       "Fast and cost-effective model for quick tasks",
			MaxTokens:         200000,
			InputCostPer1K:    0.25,
			OutputCostPer1K:   1.25,
			SupportsStreaming: true,
			SupportsTools:     true,
		},
	}

	if i, ok := info[ClaudeModel(modelID)]; ok {
		return i
	}
	return info[ClaudeHaiku35] // default
}

// ValidateModel checks if a model ID is valid and returns the normalized ID.
func ValidateModel(modelID string) string {
	for _, m := range AllModels() {
		if string(m) == modelID {
			return modelID
		}
	}
	return string(ClaudeHaiku35) // default to Haiku
}

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
