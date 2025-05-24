FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kirjaswappi-notification ./cmd/server

FROM alpine:latest

WORKDIR /

COPY --from=builder /app/kirjaswappi-notification .

EXPOSE 50051
EXPOSE 8080

ENTRYPOINT ["/kirjaswappi-notification"]
