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
	DefaultModel          string `envconfig:"DEFAULT_MODEL" default:"claude-sonnet-4-5-20250929"`
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

	// Config file path
	ConfigFile string `envconfig:"CONFIG_FILE" default:"config.yaml"`

	// Models loaded from config file
	Models []ModelConfig `ignored:"true"`
}

// ModelConfig represents a model entry in config file.
type ModelConfig struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// Load loads configuration from environment variables and config file.
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

	// Load models from config file
	cfg.Models = loadModelsFromFile(cfg.ConfigFile)

	return cfg, nil
}

// ConfigFile represents the structure of config.yaml
type ConfigFile struct {
	Models []ModelConfig `yaml:"models"`
}

// loadModelsFromFile loads models from YAML config file
func loadModelsFromFile(path string) []ModelConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default models if file doesn't exist
		return defaultModels()
	}

	var cf ConfigFile
	if err := parseYAML(data, &cf); err != nil {
		return defaultModels()
	}

	if len(cf.Models) == 0 {
		return defaultModels()
	}

	return cf.Models
}

// parseYAML is a simple YAML parser for our config format
func parseYAML(data []byte, cf *ConfigFile) error {
	lines := strings.Split(string(data), "\n")
	var currentModel *ModelConfig

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "- id:") {
			if currentModel != nil {
				cf.Models = append(cf.Models, *currentModel)
			}
			currentModel = &ModelConfig{
				ID: strings.TrimSpace(strings.TrimPrefix(line, "- id:")),
			}
		} else if currentModel != nil {
			if strings.HasPrefix(line, "name:") {
				currentModel.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
			} else if strings.HasPrefix(line, "description:") {
				currentModel.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
			}
		}
	}

	if currentModel != nil {
		cf.Models = append(cf.Models, *currentModel)
	}

	return nil
}

// defaultModels returns the default model list
func defaultModels() []ModelConfig {
	return []ModelConfig{
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", Description: "Most powerful Claude model"},
		{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5", Description: "Latest Sonnet - best for coding"},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Description: "Balanced performance and cost"},
		{ID: "claude-3-7-sonnet-20250219", Name: "Claude Sonnet 3.7", Description: "Hybrid reasoning model"},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude Haiku 3.5", Description: "Fast and cost-effective"},
	}
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
