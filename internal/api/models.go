// Package api provides HTTP handlers for the API.
package api

import (
	"net/http"
	"time"

	"claude-code-api/internal/claude"
	"claude-code-api/internal/models"

	"github.com/gin-gonic/gin"
)

// ModelsHandler handles models-related requests.
type ModelsHandler struct {
	manager *claude.Manager
}

// NewModelsHandler creates a new models handler.
func NewModelsHandler(manager *claude.Manager) *ModelsHandler {
	return &ModelsHandler{manager: manager}
}

// HandleListModels handles GET /v1/models
func (h *ModelsHandler) HandleListModels(c *gin.Context) {
	claudeVersion, _ := h.manager.GetVersion()
	ownedBy := "anthropic"
	if claudeVersion != "" {
		ownedBy = "anthropic-claude-" + claudeVersion
	}

	baseTimestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	var modelObjects []models.ModelObject
	for i, m := range models.AllModels() {
		modelObjects = append(modelObjects, models.ModelObject{
			ID:      string(m),
			Object:  "model",
			Created: baseTimestamp + int64(i),
			OwnedBy: ownedBy,
		})
	}

	c.JSON(http.StatusOK, models.ModelListResponse{
		Object: "list",
		Data:   modelObjects,
	})
}

// HandleGetModel handles GET /v1/models/:model_id
func (h *ModelsHandler) HandleGetModel(c *gin.Context) {
	modelID := c.Param("model_id")

	// Check if valid model
	for _, m := range models.AllModels() {
		if string(m) == modelID {
			claudeVersion, _ := h.manager.GetVersion()
			ownedBy := "anthropic"
			if claudeVersion != "" {
				ownedBy = "anthropic-claude-" + claudeVersion
			}

			c.JSON(http.StatusOK, models.ModelObject{
				ID:      modelID,
				Object:  "model",
				Created: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
				OwnedBy: ownedBy,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, models.ErrorResponse{
		Error: models.ErrorDetail{
			Message: "Model not found: " + modelID,
			Type:    "not_found",
			Code:    "model_not_found",
		},
	})
}

// HandleModelCapabilities handles GET /v1/models/capabilities
func (h *ModelsHandler) HandleModelCapabilities(c *gin.Context) {
	var capabilities []map[string]interface{}

	for _, m := range models.AllModels() {
		info := models.GetModelInfo(string(m))
		cap := map[string]interface{}{
			"id":                 info.ID,
			"name":               info.Name,
			"description":        info.Description,
			"max_tokens":         info.MaxTokens,
			"supports_streaming": info.SupportsStreaming,
			"supports_tools":     info.SupportsTools,
			"pricing": map[string]interface{}{
				"input_cost_per_1k_tokens":  info.InputCostPer1K,
				"output_cost_per_1k_tokens": info.OutputCostPer1K,
				"currency":                  "USD",
			},
			"features": []string{
				"text_generation",
				"conversation",
				"code_generation",
				"analysis",
				"reasoning",
			},
		}
		capabilities = append(capabilities, cap)
	}

	c.JSON(http.StatusOK, gin.H{
		"models":   capabilities,
		"total":    len(capabilities),
		"provider": "anthropic",
		"adapter":  "claude-code-api",
	})
}
