// Package api provides HTTP handlers for the API.
package api

import (
	"net/http"
	"time"

	"claude-code-api/internal/claude"
	"claude-code-api/internal/config"
	"claude-code-api/internal/models"

	"github.com/gin-gonic/gin"
)

// ModelsHandler handles models-related requests.
type ModelsHandler struct {
	cfg     *config.Config
	manager *claude.Manager
}

// NewModelsHandler creates a new models handler.
func NewModelsHandler(cfg *config.Config, manager *claude.Manager) *ModelsHandler {
	return &ModelsHandler{cfg: cfg, manager: manager}
}

// HandleListModels handles GET /v1/models
// Returns models from config file
func (h *ModelsHandler) HandleListModels(c *gin.Context) {
	claudeVersion, _ := h.manager.GetVersion()
	ownedBy := "anthropic"
	if claudeVersion != "" {
		ownedBy = "anthropic-claude-" + claudeVersion
	}

	baseTimestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	var modelObjects []models.ModelObject
	for i, m := range h.cfg.Models {
		modelObjects = append(modelObjects, models.ModelObject{
			ID:      m.ID,
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
// Returns info for any model - doesn't validate since any model name can be used
func (h *ModelsHandler) HandleGetModel(c *gin.Context) {
	modelID := c.Param("model_id")

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
}

// HandleModelCapabilities handles GET /v1/models/capabilities
func (h *ModelsHandler) HandleModelCapabilities(c *gin.Context) {
	var capabilities []map[string]interface{}

	for _, m := range h.cfg.Models {
		cap := map[string]interface{}{
			"id":                 m.ID,
			"name":               m.Name,
			"description":        m.Description,
			"supports_streaming": true,
			"supports_tools":     true,
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
