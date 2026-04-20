package twilio

import (
	"context"
	"fmt"

	"github.com/fercho/school-tracking/services/notification/internal/core/ports/resources"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
	twilioClient "github.com/twilio/twilio-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type smsSender struct {
	client    *twilioClient.RestClient
	fromPhone string
	log       *zap.Logger
}

// NewSMSSender creates a Twilio-backed SMSSender.
func NewSMSSender(cfg *env.Config, log *zap.Logger) resources.SMSSender {
	client := twilioClient.NewRestClientWithParams(twilioClient.ClientParams{
		Username: cfg.TwilioAccountSID,
		Password: cfg.TwilioAuthToken,
	})
	return &smsSender{client: client, fromPhone: cfg.TwilioFromPhone, log: log}
}

func (s *smsSender) Send(_ context.Context, toPhone, body string) error {
	if toPhone == "" {
		return fmt.Errorf("toPhone is required for SMS")
	}
	params := &openapi.CreateMessageParams{}
	params.SetTo(toPhone)
	params.SetFrom(s.fromPhone)
	params.SetBody(body)

	_, err := s.client.Api.CreateMessage(params)
	if err != nil {
		return fmt.Errorf("twilio SMS failed: %w", err)
	}
	s.log.Info("SMS sent", zap.String("to", toPhone))
	return nil
}

var Module = fx.Provide(NewSMSSender)
