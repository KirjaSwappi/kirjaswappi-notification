package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/config"
	handlergrpc "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/grpc"
	ws "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/websocket"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/logger"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
	pb "github.com/kirjaswappi/kirjaswappi-notification/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	config      *config.Config
	logger      *slog.Logger
	broadcaster *service.Broadcaster
	grpcServer  *grpc.Server
	httpServer  *http.Server
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New(cfg.LogLevel)

	// Create server
	server := &Server{
		config: cfg,
		logger: log,
	}

	// Run server
	if err := server.Run(); err != nil {
		log.Error("Server failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func (s *Server) Run() error {
	s.logger.Info("Starting notification service",
		slog.Int("http_port", s.config.HTTPPort),
		slog.Int("grpc_port", s.config.GRPCPort),
		slog.String("log_level", s.config.LogLevel))

	// Initialize broadcaster
	s.broadcaster = service.NewBroadcaster(s.logger)

	// Start gRPC server
	if err := s.startGRPCServer(); err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}

	// Start HTTP server
	if err := s.startHTTPServer(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Wait for shutdown signal
	s.waitForShutdown()

	return nil
}

func (s *Server) startGRPCServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.GRPCPort, err)
	}

	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryLoggingInterceptor),
		grpc.StreamInterceptor(s.streamLoggingInterceptor),
	)

	handler := handlergrpc.NewNotificationHandler(s.broadcaster, s.logger)
	pb.RegisterNotificationServiceServer(s.grpcServer, handler)
	reflection.Register(s.grpcServer)

	go func() {
		s.logger.Info("gRPC server started", slog.Int("port", s.config.GRPCPort))
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("gRPC server failed", slog.String("error", err.Error()))
		}
	}()

	return nil
}

func (s *Server) startHTTPServer() error {
	mux := http.NewServeMux()

	// WebSocket handler
	wsHandler := ws.NewHandler(s.broadcaster, s.logger, s.config.AllowedOrigins)
	mux.Handle("/ws", s.loggingMiddleware(wsHandler))

	// Health check
	mux.Handle("/healthz", s.loggingMiddleware(http.HandlerFunc(s.healthCheck)))

	// Stats endpoint (for monitoring)
	mux.Handle("/stats", s.loggingMiddleware(http.HandlerFunc(s.statsHandler)))

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		s.logger.Info("HTTP server started", slog.Int("port", s.config.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server failed", slog.String("error", err.Error()))
		}
	}()

	return nil
}

func (s *Server) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	s.logger.Info("Shutdown signal received", slog.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.ShutdownTimeout)*time.Second)
	defer cancel()

	s.shutdown(ctx)
}

func (s *Server) shutdown(ctx context.Context) {
	s.logger.Info("Shutting down servers...")

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("HTTP server shutdown failed", slog.String("error", err.Error()))
		} else {
			s.logger.Info("HTTP server stopped")
		}
	}

	// Shutdown gRPC server
	if s.grpcServer != nil {
		done := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			s.logger.Info("gRPC server stopped")
		case <-ctx.Done():
			s.logger.Warn("gRPC server force stopped")
			s.grpcServer.Stop()
		}
	}

	// Close broadcaster
	if s.broadcaster != nil {
		s.broadcaster.Close()
	}

	s.logger.Info("Shutdown complete")
}

// Middleware
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		if r.URL.Path != "/healthz" {
			s.logger.Info("HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Duration("duration", duration))
		}
	})
}

// gRPC Interceptors
func (s *Server) unaryLoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	s.logger.Info("gRPC request",
		slog.String("method", info.FullMethod),
		slog.Duration("duration", duration),
		slog.Bool("error", err != nil))

	return resp, err
}

func (s *Server) streamLoggingInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	start := time.Now()
	err := handler(srv, ss)
	duration := time.Since(start)

	s.logger.Info("gRPC stream",
		slog.String("method", info.FullMethod),
		slog.Duration("duration", duration),
		slog.Bool("error", err != nil))

	return err
}

// Handlers
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	users, subs := s.broadcaster.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{"status":"ok","users":%d,"subscribers":%d}`, users, subs)
	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.Error("Failed to write health check response", slog.String("error", err.Error()))
	}
}

func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	users, subs := s.broadcaster.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := fmt.Sprintf(`{"users":%d,"subscribers":%d}`, users, subs)
	if _, err := w.Write([]byte(response)); err != nil {
		s.logger.Error("Failed to write stats response", slog.String("error", err.Error()))
	}
}
