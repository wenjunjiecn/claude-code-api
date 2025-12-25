# Claude Code API Gateway

A high-performance, OpenAI-compatible API gateway for Claude Code.

## Quick Start

```bash
# Clone
git clone https://github.com/codingworkflow/claude-code-api
cd claude-code-api

# Run with Docker
make docker-run
```

The API will be available at http://localhost:8000

## Features

- **High Performance**: 17k+ QPS for health checks, 22k+ QPS for model listings
- **Dynamic Models**: Supports ANY model name (pass-through to Claude CLI)
- **Configurable**: Customizable model list via `config.yaml`
- **Streaming**: Full SSE support for real-time responses
- **OpenAI Compatible**: Drop-in replacement for OpenAI API

## Configuration

### Custom Model List

Mount your own `config.yaml` to customize the models shown in `/v1/models`:

```bash
# Run with custom config
make docker-run-config
```

Or manually:
```bash
docker run -p 8000:8000 -v $(pwd)/config.yaml:/app/config.yaml claude-code-api
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server host |
| `PORT` | `8000` | Server port |
| `CLAUDE_BINARY_PATH` | auto-detect | Path to Claude CLI |
| `CONFIG_FILE` | `config.yaml` | Path to config file |
| `DEFAULT_MODEL` | `claude-sonnet-4-5-20250929` | Default model if none specified |
| `MAX_CONCURRENT_SESSIONS` | `10` | Max concurrent sessions |
| `REQUIRE_AUTH` | `false` | Require API key auth |
| `API_KEYS` | - | Comma-separated API keys |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check (µs latency) |
| `/v1/models` | GET | List available models (from config) |
| `/v1/chat/completions` | POST | Chat completion (supports any model) |

## Supported Models

You can use **any** model name supported by Claude Code CLI. Common ones included in default config:

- `claude-opus-4-20250514`
- `claude-sonnet-4-5-20250929` (Default)
- `claude-sonnet-4-20250514`
- `claude-3-7-sonnet-20250219`
- `claude-3-5-haiku-20241022`

## Usage Examples

```bash
# Chat with specific model
curl -X POST http://localhost:8000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5-20250929",
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

## License

GNU General Public License v3.0