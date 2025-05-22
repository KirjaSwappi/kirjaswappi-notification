# Kirjaswappi Notification Microservice

This microservice handles notifications for Kirjaswappi.  
It receives notification data via gRPC calls from other services and broadcasts them in real-time to frontends via WebSockets.

## Features

- gRPC server to receive notifications from other microservices
- WebSocket server for real-time notifications to frontend clients
- Supports multiple subscribers per user
- Written in Go for high performance and easy deployment

## Architecture

- `cmd/server`: entrypoint with gRPC and WebSocket servers
- `internal/delivery/grpc`: gRPC handlers
- `internal/delivery/websocket`: WebSocket handlers
- `internal/service`: core business logic (broadcasting)
- `proto`: Protocol Buffers definitions and generated code

## Getting Started

### Prerequisites

- Go 1.20+ installed
- Protocol Buffers compiler (`protoc`) installed
- `protoc-gen-go` and `protoc-gen-go-grpc` plugins installed and in your PATH

### Build & Run

```bash
go run ./cmd/server
```

### To format the code and check linting issues, run

```bash
make spotless
```

© 2025 Kirjaswappi. All rights reserved. Unauthorized copying or distribution prohibited.
