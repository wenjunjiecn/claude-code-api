# Claude Code API - Go Implementation

# Build
build:
	CGO_ENABLED=0 go build -o bin/claude-api ./cmd/server

# Run
run: build
	./bin/claude-api

# Test
test:
	go test ./...

# Docker
docker-build:
	docker build -t claude-code-api .

docker-run:
	docker run -p 8000:8000 claude-code-api

# Clean
clean:
	rm -rf bin/

# Kill process on specific port
kill:
	@if [ -z "$(PORT)" ]; then \
		echo "Usage: make kill PORT=8000"; \
	else \
		lsof -iTCP:$(PORT) -sTCP:LISTEN -t | xargs -r kill -9 2>/dev/null || true; \
		echo "Killed processes on port $(PORT)"; \
	fi

help:
	@echo "Claude Code API Commands:"
	@echo ""
	@echo "  make build        - Build binary"
	@echo "  make run          - Build and run"
	@echo "  make test         - Run tests"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make clean        - Remove build artifacts"