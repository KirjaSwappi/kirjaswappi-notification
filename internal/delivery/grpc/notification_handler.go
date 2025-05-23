package grpc

import (
	"context"

	"github.com/kirjaswappi/kirjaswappi-notification/internal/domain"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
	pb "github.com/kirjaswappi/kirjaswappi-notification/proto"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	Broadcaster *service.Broadcaster
}

func NewNotificationHandler(b *service.Broadcaster) *NotificationHandler {
	return &NotificationHandler{Broadcaster: b}
}

func (h *NotificationHandler) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	// Override Time with current time
	now := timestamppb.Now()

	notification := domain.Notification{
		UserID:  req.UserId,
		Title:   req.Title,
		Message: req.Message,
		Time:    now.AsTime(), // use current time instead of req.Time
	}

	h.Broadcaster.Broadcast(notification)

	return &pb.NotificationResponse{Success: true}, nil
}
