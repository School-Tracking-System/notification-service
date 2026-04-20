package main

import (
	"github.com/fercho/school-tracking/services/notification/internal/core/notification"
	notifgrpc "github.com/fercho/school-tracking/services/notification/internal/infrastructure/api/grpc"
	natsmsg "github.com/fercho/school-tracking/services/notification/internal/infrastructure/messaging/nats"
	"github.com/fercho/school-tracking/services/notification/internal/infrastructure/clients"
	"github.com/fercho/school-tracking/services/notification/internal/infrastructure/persistence/postgres"
	"github.com/fercho/school-tracking/services/notification/internal/infrastructure/providers/fcm"
	"github.com/fercho/school-tracking/services/notification/internal/infrastructure/providers/twilio"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	"github.com/fercho/school-tracking/services/notification/pkg/logger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func AppModule() fx.Option {
	return fx.Options(
		env.Module,
		logger.Module,
		postgres.Module,
		clients.Module,
		natsmsg.Module,
		// Providers
		fcm.Module,
		twilio.Module,
		// Service
		notification.NotificationModule,
		// gRPC handler + server
		fx.Provide(notifgrpc.NewNotificationHandler),
		notifgrpc.ServerModule,
		// Start NATS subscribers
		fx.Invoke(natsmsg.StartSubscribers),
		// Force gRPC server start
		fx.Invoke(func(*grpc.Server) {}),
		fx.Invoke(func(l *zap.Logger, cfg *env.Config) {
			l.Info("Notification service initialized",
				zap.String("grpc_port", cfg.GRPCPort),
				zap.String("env", cfg.Environment),
			)
		}),
	)
}
