package main

import (
	"log"
	"net"
	"net/http"

	handlergrpc "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/grpc"
	ws "github.com/kirjaswappi/kirjaswappi-notification/internal/delivery/websocket"
	"github.com/kirjaswappi/kirjaswappi-notification/internal/service"
	pb "github.com/kirjaswappi/kirjaswappi-notification/proto"
	"google.golang.org/grpc"
)

func main() {
	broadcaster := service.NewBroadcaster()

	// Start gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		grpcServer := grpc.NewServer()

		handler := handlergrpc.NewNotificationHandler(broadcaster)

		pb.RegisterNotificationServiceServer(grpcServer, handler)

		log.Println("gRPC server listening on :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Start WebSocket server
	http.HandleFunc("/ws", ws.NewHandler(broadcaster))
	log.Println("WebSocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
