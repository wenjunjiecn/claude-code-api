// Package streaming provides SSE streaming utilities.
package streaming

import (
	"encoding/json"
	"fmt"
	"time"

	"claude-code-api/internal/models"

	"github.com/google/uuid"
)

// SSEFormatter formats data for Server-Sent Events.
type SSEFormatter struct{}

// FormatEvent formats a data object as an SSE event.
func (f *SSEFormatter) FormatEvent(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf("data: %s\n\n", jsonData)
}

// FormatDone returns the SSE completion signal.
func (f *SSEFormatter) FormatDone() string {
	return "data: [DONE]\n\n"
}

// FormatError formats an error as an SSE event.
func (f *SSEFormatter) FormatError(errMsg, errType string) string {
	errData := models.ErrorResponse{
		Error: models.ErrorDetail{
			Message: errMsg,
			Type:    errType,
			Code:    "stream_error",
		},
	}
	return f.FormatEvent(errData)
}

// Converter converts Claude output to OpenAI streaming format.
type Converter struct {
	Model        string
	SessionID    string
	CompletionID string
	Created      int64
}

// NewConverter creates a new streaming converter.
func NewConverter(model, sessionID string) *Converter {
	return &Converter{
		Model:        model,
		SessionID:    sessionID,
		CompletionID: fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:29]),
		Created:      time.Now().Unix(),
	}
}

// CreateInitialChunk creates the initial streaming chunk.
func (c *Converter) CreateInitialChunk() models.ChatCompletionChunk {
	return models.ChatCompletionChunk{
		ID:      c.CompletionID,
		Object:  "chat.completion.chunk",
		Created: c.Created,
		Model:   c.Model,
		Choices: []models.ChatCompletionChunkChoice{{
			Index: 0,
			Delta: models.ChatCompletionChunkDelta{
				Role:    "assistant",
				Content: "",
			},
			FinishReason: nil,
		}},
	}
}

// CreateContentChunk creates a content chunk.
func (c *Converter) CreateContentChunk(content string) models.ChatCompletionChunk {
	return models.ChatCompletionChunk{
		ID:      c.CompletionID,
		Object:  "chat.completion.chunk",
		Created: c.Created,
		Model:   c.Model,
		Choices: []models.ChatCompletionChunkChoice{{
			Index: 0,
			Delta: models.ChatCompletionChunkDelta{
				Content: content,
			},
			FinishReason: nil,
		}},
	}
}

// CreateFinalChunk creates the final chunk with finish_reason.
func (c *Converter) CreateFinalChunk() models.ChatCompletionChunk {
	stopReason := "stop"
	return models.ChatCompletionChunk{
		ID:      c.CompletionID,
		Object:  "chat.completion.chunk",
		Created: c.Created,
		Model:   c.Model,
		Choices: []models.ChatCompletionChunkChoice{{
			Index:        0,
			Delta:        models.ChatCompletionChunkDelta{},
			FinishReason: &stopReason,
		}},
	}
}

// ExtractTextContent extracts text from Claude message content.
func ExtractTextContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "text" {
					if text, ok := itemMap["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}
