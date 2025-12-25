# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git for go mod (if needed)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o claude-api ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates npm

# Install Claude Code CLI
RUN npm install -g @anthropic-ai/claude-code

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/claude-api .

# Expose port
EXPOSE 8000

# Run
CMD ["./claude-api"]
