package clients

import (
	"context"

	authpb "github.com/fercho/school-tracking/proto/gen/auth/v1"
	fleetpb "github.com/fercho/school-tracking/proto/gen/fleet/v1"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FleetClients struct {
	GuardianService fleetpb.GuardianServiceClient
	StudentService  fleetpb.StudentServiceClient
	AuthService     authpb.AuthServiceClient // Maybe needed later or just reuse the pattern
	Conn            *grpc.ClientConn
}

func NewFleetClient(lc fx.Lifecycle, cfg *env.Config, log *zap.Logger) (FleetClients, error) {
	conn, err := grpc.Dial(
		cfg.FleetServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("Failed to connect to Fleet API", zap.Error(err))
		return FleetClients{}, err
	}

	clients := FleetClients{
		GuardianService: fleetpb.NewGuardianServiceClient(conn),
		StudentService:  fleetpb.NewStudentServiceClient(conn),
		Conn:            conn,
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("Closing Fleet gRPC connection")
			return conn.Close()
		},
	})

	return clients, nil
}
