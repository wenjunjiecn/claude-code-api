// Package models defines OpenAI-compatible API types.
package models

import "strings"

// ChatMessage represents a message in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role" binding:"required,oneof=system user assistant"`
	Content any    `json:"content" binding:"required"`
	Name    string `json:"name,omitempty"`
}

// GetTextContent extracts text content from any format.
func (m *ChatMessage) GetTextContent() string {
	switch v := m.Content.(type) {
	case string:
		return v
	case []interface{}:
		var parts []string
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if text, exists := itemMap["text"]; exists {
					if s, ok := text.(string); ok {
						parts = append(parts, s)
					}
				} else if content, exists := itemMap["content"]; exists {
					if s, ok := content.(string); ok {
						parts = append(parts, s)
					}
				}
			}
		}
		return strings.Join(parts, "\n")
	default:
		return ""
	}
}

// ChatCompletionRequest is the request body for chat completions.
type ChatCompletionRequest struct {
	Model            string        `json:"model" binding:"required"`
	Messages         []ChatMessage `json:"messages" binding:"required,min=1"`
	Temperature      *float64      `json:"temperature,omitempty"`
	TopP             *float64      `json:"top_p,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	Stream           bool          `json:"stream,omitempty"`
	Stop             any           `json:"stop,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	User             string        `json:"user,omitempty"`

	// Extension fields for Claude Code
	ProjectID    string `json:"project_id,omitempty"`
	SessionID    string `json:"session_id,omitempty"`
	SystemPrompt string `json:"system_prompt,omitempty"`
}

// ChatCompletionChoice represents a single completion choice.
type ChatCompletionChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

// ChatCompletionUsage contains token usage information.
type ChatCompletionUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse is the response for chat completions.
type ChatCompletionResponse struct {
	ID        string                 `json:"id"`
	Object    string                 `json:"object"`
	Created   int64                  `json:"created"`
	Model     string                 `json:"model"`
	Choices   []ChatCompletionChoice `json:"choices"`
	Usage     ChatCompletionUsage    `json:"usage"`
	SessionID string                 `json:"session_id,omitempty"`
	ProjectID string                 `json:"project_id,omitempty"`
}

// ChatCompletionChunkDelta represents delta content in streaming.
type ChatCompletionChunkDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ChatCompletionChunkChoice represents a streaming choice.
type ChatCompletionChunkChoice struct {
	Index        int                      `json:"index"`
	Delta        ChatCompletionChunkDelta `json:"delta"`
	FinishReason *string                  `json:"finish_reason"`
}

// ChatCompletionChunk is a streaming response chunk.
type ChatCompletionChunk struct {
	ID      string                      `json:"id"`
	Object  string                      `json:"object"`
	Created int64                       `json:"created"`
	Model   string                      `json:"model"`
	Choices []ChatCompletionChunkChoice `json:"choices"`
}

// ModelObject represents a model in the models list.
type ModelObject struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelListResponse is the response for listing models.
type ModelListResponse struct {
	Object string        `json:"object"`
	Data   []ModelObject `json:"data"`
}

// ErrorDetail contains error information.
type ErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
}

// ErrorResponse wraps an error detail.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// HealthCheckResponse is the health endpoint response.
type HealthCheckResponse struct {
	Status         string `json:"status"`
	Version        string `json:"version"`
	ClaudeVersion  string `json:"claude_version,omitempty"`
	ActiveSessions int    `json:"active_sessions"`
}
