package grpc

import (
	"context"
	"fmt"
	"net"

	pb "github.com/fercho/school-tracking/proto/gen/notification/v1"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer(
	lc fx.Lifecycle,
	cfg *env.Config,
	log *zap.Logger,
	handler pb.NotificationServiceServer,
) *grpc.Server {
	srv := grpc.NewServer()
	pb.RegisterNotificationServiceServer(srv, handler)
	reflection.Register(srv)

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
			if err != nil {
				return fmt.Errorf("notification gRPC listen failed: %w", err)
			}
			log.Info("Notification gRPC server listening", zap.String("port", cfg.GRPCPort))
			go func() {
				if err := srv.Serve(lis); err != nil {
					log.Fatal("notification gRPC serve failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("Stopping Notification gRPC server")
			srv.GracefulStop()
			return nil
		},
	})
	return srv
}

var ServerModule = fx.Provide(NewGRPCServer)
