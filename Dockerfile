FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kirjaswappi-notification ./cmd/server

FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /

COPY --from=builder /app/kirjaswappi-notification .

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

USER appuser

EXPOSE 50051
EXPOSE 8080

ENTRYPOINT ["/kirjaswappi-notification"]
