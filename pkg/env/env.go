package env

import (
	"github.com/caarlos0/env/v10"
	"go.uber.org/fx"
)

type Config struct {
	Port              string `env:"PORT" envDefault:"8085"`
	GRPCPort          string `env:"GRPC_PORT" envDefault:"9095"`
	Environment       string `env:"ENVIRONMENT" envDefault:"development"`
	DatabaseURL       string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/notification_db?sslmode=disable"`
	NATSUrl           string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	TwilioAccountSID  string `env:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken   string `env:"TWILIO_AUTH_TOKEN"`
	TwilioFromPhone   string `env:"TWILIO_FROM_PHONE"`
	FCMCredentialPath string `env:"FCM_CREDENTIAL_PATH"`
	FleetServiceURL   string `env:"FLEET_SERVICE_URL" envDefault:"localhost:9090"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

var Module = fx.Provide(NewConfig)
