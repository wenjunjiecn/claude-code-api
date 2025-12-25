// Claude Code API Gateway - Go Implementation
//
// A simple, focused OpenAI-compatible API gateway for Claude Code.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"claude-code-api/internal/api"
	"claude-code-api/internal/claude"
	"claude-code-api/internal/config"
	"claude-code-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const version = "1.0.0"

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Set log level
	switch cfg.LogLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Initialize Claude manager
	manager := claude.NewManager(cfg)

	// Verify Claude availability
	claudeVersion, err := manager.GetVersion()
	if err != nil {
		log.Fatal().Err(err).Msg("Claude Code CLI not available")
	}
	log.Info().Str("claude_version", claudeVersion).Msg("Claude Code available")

	// Setup Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(api.LoggingMiddleware())
	router.Use(api.CORSMiddleware(cfg))
	router.Use(api.AuthMiddleware(cfg))

	// Create handlers
	chatHandler := api.NewChatHandler(cfg, manager)
	modelsHandler := api.NewModelsHandler(manager)

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "Claude Code API Gateway",
			"version":     version,
			"description": "OpenAI-compatible API for Claude Code",
			"endpoints": gin.H{
				"chat":   "/v1/chat/completions",
				"models": "/v1/models",
			},
			"docs":   "/docs",
			"health": "/health",
		})
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		cv, err := manager.GetVersion()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, models.HealthCheckResponse{
				Status:  "unhealthy",
				Version: version,
			})
			return
		}

		c.JSON(http.StatusOK, models.HealthCheckResponse{
			Status:         "healthy",
			Version:        version,
			ClaudeVersion:  cv,
			ActiveSessions: manager.ActiveSessionCount(),
		})
	})

	// API routes
	v1 := router.Group("/v1")
	{
		v1.POST("/chat/completions", chatHandler.HandleChatCompletion)
		v1.GET("/models", modelsHandler.HandleListModels)
		v1.GET("/models/capabilities", modelsHandler.HandleModelCapabilities)
		v1.GET("/models/:model_id", modelsHandler.HandleGetModel)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info().Str("address", addr).Msg("Starting Claude Code API Gateway")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed")
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager.CleanupAll()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
