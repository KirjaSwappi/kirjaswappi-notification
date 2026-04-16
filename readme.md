# Kirjaswappi Notification Microservice

This microservice handles notifications for Kirjaswappi.  
It receives notification data via gRPC calls from other services and broadcasts them in real-time to frontends via WebSockets.

## Features

- gRPC server to receive notifications from other microservices
- WebSocket server for real-time notifications to frontend clients
- JWT and API key authentication for WebSocket; API key authentication for gRPC
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
| `API_KEY` | *(empty)* | API key for gRPC, `/stats`, and WebSocket authentication |
| `JWT_SECRET` | *(empty)* | Secret for verifying JWT tokens on WebSocket connections |

> **Note:** If neither `API_KEY` nor `JWT_SECRET` is set, authentication is disabled (development mode). Setting only `JWT_SECRET` enables JWT verification for WebSocket connections, but leaves gRPC and `/stats` unauthenticated. In production, set `API_KEY` to protect gRPC and `/stats`; set `JWT_SECRET` as well if you want JWT-based authentication on WebSocket connections.

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
  -e API_KEY="your-api-key" \
  -e JWT_SECRET="your-jwt-secret" \
  kirjaswappi-notification
```

## API Endpoints

### WebSocket

- `GET /ws?token={token}[&userId={userId}]` - Real-time notifications (`userId` is required in development and API-key modes; ignored for JWT)

**Authentication (checked in order):**

1. **No auth configured** — `userId` query param is required and used directly (development mode)
2. **API key** — pass `API_KEY` as `token`; `userId` query param is required and trusted
3. **JWT** — pass a JWT as `token`; user ID is derived from the `sub` claim, `userId` query param is ignored

### HTTP

- `GET /healthz` - Health check with connection stats (public)
- `GET /stats` - Connection statistics (requires `X-API-Key` header when `API_KEY` is set)

### gRPC

- `SendNotification` - Send notification to a user (requires `x-api-key` metadata when `API_KEY` is set)

## WebSocket Usage

```javascript
// With JWT token
const token = 'eyJhbGciOiJIUzI1NiIs...';
const ws = new WebSocket(`wss://ans.kirjaswappi.fi/ws?token=${token}`);

ws.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    console.log('Notification:', notification);
    // { UserID: "123", Title: "New Message", Message: "...", Time: "..." }
};
```

```javascript
// With API key (server-to-server)
const apiKey = process.env.NOTIFICATION_API_KEY;
const wsApiKey = new WebSocket(`wss://ans.kirjaswappi.fi/ws?token=${apiKey}&userId=123`);

wsApiKey.onmessage = function(event) {
    const notification = JSON.parse(event.data);
    console.log('Notification:', notification);
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

© 2025 Kirjaswappi. All rights reserved. Unauthorized copying or distribution prohibited.
