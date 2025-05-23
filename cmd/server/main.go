package main

import (
	"context"
	"log"
	"net"
	"net/http"

	handlergrpc "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/grpc"
	ws "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/websocket"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
	pb "github.com/kirjaswappi/kirjaswappi-notification/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// --- gRPC Interceptors ---
func unaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("[gRPC] Unary call: %s | Payload: %+v", info.FullMethod, req)
	return handler(ctx, req)
}

func streamLoggingInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	log.Printf("[gRPC] Stream call: %s", info.FullMethod)
	return handler(srv, ss)
}

// --- HTTP Logging Middleware ---
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	broadcaster := service.NewBroadcaster()

	// --- Start gRPC server ---
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(unaryLoggingInterceptor),
			grpc.StreamInterceptor(streamLoggingInterceptor),
		)

		handler := handlergrpc.NewNotificationHandler(broadcaster)
		pb.RegisterNotificationServiceServer(grpcServer, handler)

		reflection.Register(grpcServer)

		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// --- Start WebSocket server ---
	http.Handle("/ws", loggingMiddleware(http.HandlerFunc(ws.NewHandler(broadcaster))))

	// --- Health Check ---
	http.Handle("/healthz", loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})))

	log.Println("HTTP/WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
