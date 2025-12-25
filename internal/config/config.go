// Package config provides configuration management for the Claude Code API Gateway.
package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all application configuration.
type Config struct {
	// Server settings
	Host string `envconfig:"HOST" default:"0.0.0.0"`
	Port int    `envconfig:"PORT" default:"8000"`

	// Claude settings
	ClaudeBinaryPath      string `envconfig:"CLAUDE_BINARY_PATH"`
	DefaultModel          string `envconfig:"DEFAULT_MODEL" default:"claude-3-5-sonnet-20241022"`
	MaxConcurrentSessions int    `envconfig:"MAX_CONCURRENT_SESSIONS" default:"10"`
	SessionTimeoutMinutes int    `envconfig:"SESSION_TIMEOUT_MINUTES" default:"30"`
	StreamingTimeoutSecs  int    `envconfig:"STREAMING_TIMEOUT_SECONDS" default:"300"`

	// Project settings
	ProjectRoot string `envconfig:"PROJECT_ROOT" default:"/tmp/claude_projects"`

	// Auth settings
	APIKeys     []string `envconfig:"API_KEYS"`
	RequireAuth bool     `envconfig:"REQUIRE_AUTH" default:"false"`

	// CORS settings
	AllowedOrigins []string `envconfig:"ALLOWED_ORIGINS" default:"*"`

	// Logging
	LogLevel string `envconfig:"LOG_LEVEL" default:"info"`
}

// Load loads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	// Auto-detect Claude binary if not set
	if cfg.ClaudeBinaryPath == "" {
		cfg.ClaudeBinaryPath = findClaudeBinary()
	}

	// Ensure project root exists
	if err := os.MkdirAll(cfg.ProjectRoot, 0755); err != nil {
		return nil, err
	}

	return cfg, nil
}

// findClaudeBinary attempts to find the Claude CLI binary.
func findClaudeBinary() string {
	// Check PATH first
	if path, err := exec.LookPath("claude"); err == nil {
		return path
	}

	// Try npm global bin
	if out, err := exec.Command("npm", "bin", "-g").Output(); err == nil {
		npmPath := filepath.Join(strings.TrimSpace(string(out)), "claude")
		if _, err := os.Stat(npmPath); err == nil {
			return npmPath
		}
	}

	// Common locations
	locations := []string{
		"/usr/local/bin/claude",
	}
	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return "claude" // fallback
}
