package bank

import (
	"context"
	"os"

	notificationpb "github.com/RAF-SI-2025/Banka-3-Backend/gen/notification"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Server) sendCardCreatedEmail(ctx context.Context, email string) error {
	addr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if addr == "" {
		addr = defaultNotificationURL
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := notificationpb.NewNotificationServiceClient(conn)
	_, err = client.SendCardCreatedEmail(ctx, &notificationpb.CardCreatedMailRequest{
		ToAddr: email,
	})
	return err
}

func (s *Server) sendCardConfirmationEmail(ctx context.Context, email string, link string) error {
	addr := os.Getenv("NOTIFICATION_GRPC_ADDR")
	if addr == "" {
		addr = defaultNotificationURL
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := notificationpb.NewNotificationServiceClient(conn)
	_, err = client.SendCardConfirmationEmail(ctx, &notificationpb.CardConfirmationMailRequest{
		ToAddr: email,
		Link:   link,
	})
	return err
}
