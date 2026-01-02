package websocket

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
)

type Handler struct {
	broadcaster    *service.Broadcaster
	logger         *slog.Logger
	allowedOrigins []string
	upgrader       websocket.Upgrader
}

func NewHandler(b *service.Broadcaster, logger *slog.Logger, allowedOrigins []string) *Handler {
	h := &Handler{
		broadcaster:    b,
		logger:         logger,
		allowedOrigins: allowedOrigins,
	}

	h.upgrader = websocket.Upgrader{
		CheckOrigin:     h.checkOrigin,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return h
}

func (h *Handler) checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Allow same origin
	if origin == "" {
		return true
	}

	// Check against allowed origins
	for _, allowed := range h.allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	h.logger.Warn("WebSocket connection rejected",
		slog.String("origin", origin),
		slog.String("remote_addr", r.RemoteAddr))
	return false
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		h.logger.Warn("WebSocket connection missing userId",
			slog.String("remote_addr", r.RemoteAddr))
		http.Error(w, "userId query param required", http.StatusBadRequest)
		return
	}

	// Basic userID validation
	if len(userID) == 0 || len(userID) > 100 || strings.ContainsAny(userID, "\n\r\t") {
		h.logger.Warn("WebSocket connection invalid userId",
			slog.String("user_id", userID),
			slog.String("remote_addr", r.RemoteAddr))
		http.Error(w, "invalid userId", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed",
			slog.String("error", err.Error()),
			slog.String("user_id", userID),
			slog.String("remote_addr", r.RemoteAddr))
		return
	}

	h.logger.Info("WebSocket connection established",
		slog.String("user_id", userID),
		slog.String("remote_addr", r.RemoteAddr))

	h.handleConnection(conn, userID)
}

func (h *Handler) handleConnection(conn *websocket.Conn, userID string) {
	defer func() {
		if err := conn.Close(); err != nil {
			h.logger.Debug("WebSocket connection close error",
				slog.String("error", err.Error()),
				slog.String("user_id", userID))
		}
		h.logger.Info("WebSocket connection closed", slog.String("user_id", userID))
	}()

	// Set connection timeouts
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		h.logger.Debug("Failed to set read deadline", slog.String("error", err.Error()))
	}
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		h.logger.Debug("Failed to set write deadline", slog.String("error", err.Error()))
	}

	// Handle ping/pong for connection health
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			h.logger.Debug("Failed to set read deadline in pong handler", slog.String("error", err.Error()))
		}
		return nil
	})

	ch := h.broadcaster.Subscribe(userID)
	defer h.broadcaster.Unsubscribe(userID, ch)

	// Start ping ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle incoming messages (for ping/pong)
	go func() {
		defer cancel()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket read error",
						slog.String("error", err.Error()),
						slog.String("user_id", userID))
				}
				return
			}
		}
	}()

	// Main message loop
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}

			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				h.logger.Debug("Failed to set write deadline", slog.String("error", err.Error()))
			}
			if err := conn.WriteJSON(msg); err != nil {
				h.logger.Error("WebSocket write error",
					slog.String("error", err.Error()),
					slog.String("user_id", userID))
				return
			}

		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				h.logger.Debug("Failed to set write deadline for ping", slog.String("error", err.Error()))
			}
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Debug("WebSocket ping failed",
					slog.String("error", err.Error()),
					slog.String("user_id", userID))
				return
			}

		case <-ctx.Done():
			return
		}
	}
}
