package grpc

import (
	"context"
	"log/slog"
	"strings"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/domain"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
	pb "github.com/kirjaswappi/kirjaswappi-notification/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	broadcaster *service.Broadcaster
	logger      *slog.Logger
}

func NewNotificationHandler(b *service.Broadcaster, logger *slog.Logger) *NotificationHandler {
	return &NotificationHandler{
		broadcaster: b,
		logger:      logger,
	}
}

func (h *NotificationHandler) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	// Validate request
	if err := h.validateRequest(req); err != nil {
		h.logger.Warn("Invalid notification request",
			slog.String("error", err.Error()),
			slog.String("user_id", req.GetUserId()))
		return nil, err
	}

	notification := domain.Notification{
		UserID:  req.UserId,
		Title:   req.Title,
		Message: req.Message,
		Time:    req.GetTime().AsTime(),
	}

	delivered := h.broadcaster.Broadcast(notification)
	if delivered == 0 {
		h.logger.Warn("notification delivered to zero subscribers",
			slog.String("userId", req.UserId),
			slog.String("title", req.Title))
	}

	h.logger.Info("Notification sent",
		slog.String("user_id", req.UserId),
		slog.String("title", req.Title))

	return &pb.NotificationResponse{Success: delivered > 0}, nil
}

func (h *NotificationHandler) validateRequest(req *pb.NotificationRequest) error {
	if req.UserId == "" {
		return status.Error(codes.InvalidArgument, "userId is required")
	}

	if len(req.UserId) > 100 {
		return status.Error(codes.InvalidArgument, "userId too long")
	}

	if strings.ContainsAny(req.UserId, "\n\r\t") {
		return status.Error(codes.InvalidArgument, "userId contains invalid characters")
	}

	if req.Title == "" {
		return status.Error(codes.InvalidArgument, "title is required")
	}

	if len(req.Title) > 200 {
		return status.Error(codes.InvalidArgument, "title too long")
	}

	if req.Message == "" {
		return status.Error(codes.InvalidArgument, "message is required")
	}

	if len(req.Message) > 1000 {
		return status.Error(codes.InvalidArgument, "message too long")
	}

	return nil
}
