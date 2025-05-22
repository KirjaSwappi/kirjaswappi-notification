# Use an official Go image for building
FROM golang:1.24 as builder

WORKDIR /app

# Copy go mod and sum files first, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o kirjaswappi-notification ./cmd/server

# Final stage: minimal image
FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/kirjaswappi-notification .

# Expose gRPC and WebSocket ports
EXPOSE 50051
EXPOSE 8080

# Run the service
ENTRYPOINT ["/kirjaswappi-notification"]
