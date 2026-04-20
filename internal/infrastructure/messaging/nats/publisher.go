package nats

import (
	"context"
	"fmt"

	"github.com/fercho/school-tracking/services/notification/internal/core/ports/resources"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	nats "github.com/nats-io/nats.go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type publisher struct{ js nats.JetStreamContext }

func NewPublisher(nc *nats.Conn) (resources.EventPublisher, error) {
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}
	return &publisher{js: js}, nil
}

func (p *publisher) Publish(_ context.Context, subject string, payload []byte) error {
	_, err := p.js.Publish(subject, payload)
	return err
}

// NewNATSConnection opens a connection to the NATS server.
func NewNATSConnection(lc fx.Lifecycle, cfg *env.Config, log *zap.Logger) (*nats.Conn, error) {
	nc, err := nats.Connect(cfg.NATSUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("Connected to NATS", zap.String("url", cfg.NATSUrl))
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("Closing NATS connection")
			nc.Close()
			return nil
		},
	})
	return nc, nil
}

// Module provides NATS dependencies to the fx graph.
var Module = fx.Options(
	fx.Provide(NewNATSConnection),
	fx.Provide(NewPublisher),
)
