# Claude Code API Gateway

A simple, focused OpenAI-compatible API gateway for Claude Code with streaming support.

## Quick Start

```bash
# Clone
git clone https://github.com/codingworkflow/claude-code-api
cd claude-code-api

# Build and run
make run
```

The API will be available at http://localhost:8000

## Prerequisites

- Go 1.22+
- Claude Code CLI installed (`npm install -g @anthropic-ai/claude-code`)

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/v1/models` | GET | List available models |
| `/v1/chat/completions` | POST | Chat completion |

## Supported Models

- `claude-opus-4-20250514` - Claude Opus 4
- `claude-sonnet-4-20250514` - Claude Sonnet 4
- `claude-3-7-sonnet-20250219` - Claude Sonnet 3.7
- `claude-3-5-haiku-20241022` - Claude Haiku 3.5

## Usage Examples

```bash
# Health check
curl http://localhost:8000/health

# List models
curl http://localhost:8000/v1/models

# Chat completion
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-haiku-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Streaming
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-haiku-20241022",
    "messages": [{"role": "user", "content": "Tell me a joke"}],
    "stream": true
  }'
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server host |
| `PORT` | `8000` | Server port |
| `CLAUDE_BINARY_PATH` | auto-detect | Path to Claude CLI |
| `MAX_CONCURRENT_SESSIONS` | `10` | Max concurrent sessions |
| `REQUIRE_AUTH` | `false` | Require API key auth |
| `API_KEYS` | - | Comma-separated API keys |

## Makefile Commands

```bash
make build    # Build binary
make run      # Build and run
make test     # Run tests
make clean    # Remove build artifacts
```

## License

GNU General Public License v3.0