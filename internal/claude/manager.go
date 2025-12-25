// Package claude provides Claude CLI process management.
package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"claude-code-api/internal/config"
	"claude-code-api/internal/models"

	"github.com/rs/zerolog/log"
)

// Process represents a single Claude CLI process.
type Process struct {
	SessionID   string
	ProjectPath string
	cmd         *exec.Cmd
	IsRunning   bool
	Output      chan models.ClaudeMessage
	mu          sync.Mutex
}

// Start executes the Claude CLI and captures output.
func (p *Process) Start(ctx context.Context, cfg *config.Config, prompt, model, systemPrompt string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	args := []string{"-p", prompt}

	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}
	if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args,
		"--output-format", "stream-json",
		"--verbose",
		"--dangerously-skip-permissions",
	)

	p.cmd = exec.CommandContext(ctx, cfg.ClaudeBinaryPath, args...)
	p.cmd.Dir = p.ProjectPath

	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := p.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start claude: %w", err)
	}

	p.IsRunning = true
	p.Output = make(chan models.ClaudeMessage, 100)

	// Read stdout JSONL
	go func() {
		defer close(p.Output)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var msg models.ClaudeMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				log.Warn().Err(err).Str("line", line).Msg("Failed to parse JSONL")
				continue
			}

			// Extract session ID from first message
			if p.SessionID == "" && msg.SessionID != "" {
				p.SessionID = msg.SessionID
			}

			p.Output <- msg
		}
		p.mu.Lock()
		p.IsRunning = false
		p.mu.Unlock()
	}()

	// Log stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Debug().Str("stderr", scanner.Text()).Msg("Claude stderr")
		}
	}()

	// Wait for completion
	go func() {
		if err := p.cmd.Wait(); err != nil {
			log.Error().Err(err).Msg("Claude process exited with error")
		}
	}()

	return nil
}

// Stop terminates the Claude process.
func (p *Process) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
	}
	p.IsRunning = false
}

// Manager manages multiple Claude processes.
type Manager struct {
	cfg         *config.Config
	processes   map[string]*Process
	mu          sync.RWMutex
	version     string
	versionOnce sync.Once
}

// NewManager creates a new Claude manager.
func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		cfg:       cfg,
		processes: make(map[string]*Process),
	}
}

// GetVersion returns the Claude CLI version.
// Caches the result to avoid repeated subprocess calls.
func (m *Manager) GetVersion() (string, error) {
	var err error
	m.versionOnce.Do(func() {
		cmd := exec.Command(m.cfg.ClaudeBinaryPath, "--version")
		var out []byte
		out, err = cmd.Output()
		if err == nil {
			m.version = strings.TrimSpace(string(out))
		}
	})

	if m.version == "" && err != nil {
		return "", fmt.Errorf("failed to get claude version: %w", err)
	}
	return m.version, nil
}

// CreateSession creates and starts a new Claude session.
func (m *Manager) CreateSession(ctx context.Context, projectPath, prompt, model, systemPrompt string) (*Process, error) {
	m.mu.Lock()
	if len(m.processes) >= m.cfg.MaxConcurrentSessions {
		m.mu.Unlock()
		return nil, fmt.Errorf("max concurrent sessions (%d) reached", m.cfg.MaxConcurrentSessions)
	}
	m.mu.Unlock()

	proc := &Process{
		ProjectPath: projectPath,
	}

	if err := proc.Start(ctx, m.cfg, prompt, model, systemPrompt); err != nil {
		return nil, err
	}

	// Don't store - Claude CLI completes immediately
	log.Info().Str("session_id", proc.SessionID).Msg("Claude session created")

	return proc, nil
}

// ActiveSessionCount returns the number of active sessions.
func (m *Manager) ActiveSessionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.processes)
}

// CleanupAll stops all sessions.
func (m *Manager) CleanupAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, proc := range m.processes {
		proc.Stop()
		delete(m.processes, id)
	}
}
