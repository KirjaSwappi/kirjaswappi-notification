# Kirjaswappi Notification Microservice

This microservice handles notifications for Kirjaswappi.  
It receives notification data via gRPC calls from other services and broadcasts them in real-time to frontends via WebSockets.

## Features

- gRPC server to receive notifications from other microservices
- WebSocket server for real-time notifications to frontend clients
- Supports multiple subscribers per user
- Production-ready with structured logging, graceful shutdown, and health checks
- Configurable via environment variables
- Written in Go for high performance and easy deployment

## Architecture

- `cmd/server`: entrypoint with gRPC and WebSocket servers
- `internal/delivery/grpc`: gRPC handlers with validation
- `internal/delivery/websocket`: WebSocket handlers with security
- `internal/service`: core business logic (broadcasting)
- `internal/config`: configuration management
- `internal/logger`: structured logging
- `proto`: Protocol Buffers definitions and generated code

## Configuration

Configure the service using environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | 8080 | HTTP/WebSocket server port |
| `GRPC_PORT` | 50051 | gRPC server port |
| `LOG_LEVEL` | info | Log level (debug, info, warn, error) |
| `ALLOWED_ORIGINS` | * | Comma-separated list of allowed WebSocket origins |
| `SHUTDOWN_TIMEOUT` | 30 | Graceful shutdown timeout in seconds |

## Getting Started

### Prerequisites

- Go 1.24+ installed
- Protocol Buffers compiler (`protoc`) installed
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins installed and in your PATH

### Build & Run

```bash
# Development
go run ./cmd/server

# Production build
go build -o notification-service ./cmd/server
./notification-service
```

### Docker

```bash
# Build
docker build -t kirjaswappi-notification .

# Run
docker run -p 8080:8080 -p 50051:50051 \
  -e LOG_LEVEL=info \
  -e ALLOWED_ORIGINS="https://kirjaswappi.fi,https://app.kirjaswappi.fi" \
  kirjaswappi-notification
```

## API Endpoints

### WebSocket
- `GET /ws?userId={userId}` - WebSocket connection for real-time notifications

### HTTP
- `GET /healthz` - Health check with connection stats
- `GET /stats` - Connection statistics

### gRPC
- `SendNotification` - Send notification to user

## WebSocket Usage

```javascript
const ws = new WebSocket('wss://notify.kirjaswappi.fi/ws?userId=123');

ws.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    console.log('Notification:', notification);
    // { UserID: "123", Title: "New Message", Message: "...", Time: "..." }
};
```

## Development

### Testing
```bash
go test ./...
```

### Linting
```bash
make spotless
```

### Code Quality
- Structured logging with correlation
- Input validation and sanitization
- Graceful shutdown handling
- Connection health monitoring (ping/pong)
- CORS protection
- Non-root container execution

© 2025 Kirjaswappi. All rights reserved. Unauthorized copying or distribution prohibited.
