// Package api provides HTTP handlers for the API.
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claude-code-api/internal/claude"
	"claude-code-api/internal/config"
	"claude-code-api/internal/models"
	"claude-code-api/internal/streaming"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ChatHandler handles chat completion requests.
type ChatHandler struct {
	cfg     *config.Config
	manager *claude.Manager
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(cfg *config.Config, manager *claude.Manager) *ChatHandler {
	return &ChatHandler{cfg: cfg, manager: manager}
}

// HandleChatCompletion handles POST /v1/chat/completions
func (h *ChatHandler) HandleChatCompletion(c *gin.Context) {
	var req models.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Message: fmt.Sprintf("Invalid request: %v", err),
				Type:    "invalid_request_error",
			},
		})
		return
	}

	// Use model directly - Claude CLI will validate
	claudeModel := req.Model
	if claudeModel == "" {
		claudeModel = h.cfg.DefaultModel
	}

	// Extract user prompt
	var userPrompt string
	var systemPrompt string
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			userPrompt = msg.GetTextContent()
		} else if msg.Role == "system" {
			systemPrompt = msg.GetTextContent()
		}
	}

	if userPrompt == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: models.ErrorDetail{
				Message: "At least one user message is required",
				Type:    "invalid_request_error",
				Code:    "missing_user_message",
			},
		})
		return
	}

	if req.SystemPrompt != "" {
		systemPrompt = req.SystemPrompt
	}

	// Setup project directory
	projectID := req.ProjectID
	if projectID == "" {
		projectID = "default"
	}
	projectPath := filepath.Join(h.cfg.ProjectRoot, projectID)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		log.Error().Err(err).Msg("Failed to create project directory")
	}

	// Create Claude session
	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(h.cfg.StreamingTimeoutSecs)*time.Second)
	defer cancel()

	proc, err := h.manager.CreateSession(ctx, projectPath, userPrompt, claudeModel, systemPrompt)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Claude session")
		c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
			Error: models.ErrorDetail{
				Message: fmt.Sprintf("Failed to start Claude: %v", err),
				Type:    "service_unavailable",
				Code:    "claude_unavailable",
			},
		})
		return
	}

	sessionID := proc.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	if req.Stream {
		h.handleStreamingResponse(c, proc, claudeModel, sessionID, projectID)
	} else {
		h.handleNonStreamingResponse(c, proc, claudeModel, sessionID, projectID)
	}
}

func (h *ChatHandler) handleStreamingResponse(c *gin.Context, proc *claude.Process, model, sessionID, projectID string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Session-ID", sessionID)
	c.Header("X-Project-ID", projectID)

	formatter := &streaming.SSEFormatter{}
	converter := streaming.NewConverter(model, sessionID)

	// Send initial chunk
	c.Writer.WriteString(formatter.FormatEvent(converter.CreateInitialChunk()))
	c.Writer.Flush()

	// Stream Claude output
	for msg := range proc.Output {
		if msg.Type == "assistant" && msg.Message != nil {
			content := streaming.ExtractTextContent(msg.Message.Content)
			if content != "" {
				c.Writer.WriteString(formatter.FormatEvent(converter.CreateContentChunk(content)))
				c.Writer.Flush()
			}
		}
		if msg.Type == "result" {
			break
		}
	}

	// Send final chunk and done
	c.Writer.WriteString(formatter.FormatEvent(converter.CreateFinalChunk()))
	c.Writer.WriteString(formatter.FormatDone())
	c.Writer.Flush()
}

func (h *ChatHandler) handleNonStreamingResponse(c *gin.Context, proc *claude.Process, model, sessionID, projectID string) {
	var contentParts []string

	// Collect all output
	for msg := range proc.Output {
		if msg.Type == "assistant" && msg.Message != nil {
			content := streaming.ExtractTextContent(msg.Message.Content)
			if content != "" {
				contentParts = append(contentParts, content)
			}
		}
		if msg.Type == "result" {
			break
		}
	}

	completeContent := strings.Join(contentParts, "\n")
	if completeContent == "" {
		completeContent = "Hello! I'm Claude, ready to help."
	}

	completionID := fmt.Sprintf("chatcmpl-%s", uuid.New().String()[:29])
	wordCount := len(strings.Fields(completeContent))

	response := models.ChatCompletionResponse{
		ID:      completionID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []models.ChatCompletionChoice{{
			Index: 0,
			Message: models.ChatMessage{
				Role:    "assistant",
				Content: completeContent,
			},
			FinishReason: "stop",
		}},
		Usage: models.ChatCompletionUsage{
			PromptTokens:     10,
			CompletionTokens: wordCount,
			TotalTokens:      10 + wordCount,
		},
		SessionID: sessionID,
		ProjectID: projectID,
	}

	c.JSON(http.StatusOK, response)
}
