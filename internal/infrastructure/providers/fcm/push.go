package fcm

import (
	"context"
	"fmt"

	"github.com/fercho/school-tracking/services/notification/internal/core/ports/resources"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// pushNotifier is a stub FCM push implementation.
// In production, replace with firebase-admin-go SDK using a service account credential.
type pushNotifier struct {
	credentialPath string
	log            *zap.Logger
}

// NewPushNotifier creates an FCM PushNotifier.
// When FCM_CREDENTIAL_PATH is empty, sends are logged and skipped (dev mode).
func NewPushNotifier(cfg *env.Config, log *zap.Logger) resources.PushNotifier {
	if cfg.FCMCredentialPath == "" {
		log.Warn("FCM_CREDENTIAL_PATH not set — push notifications are disabled (dev mode)")
	}
	return &pushNotifier{credentialPath: cfg.FCMCredentialPath, log: log}
}

func (p *pushNotifier) Send(_ context.Context, userID, title, body, data string) error {
	if p.credentialPath == "" {
		p.log.Debug("FCM stub: skipping push notification",
			zap.String("user_id", userID),
			zap.String("title", title),
		)
		return nil
	}
	// TODO: integrate firebase-admin-go when credentials are available.
	return fmt.Errorf("FCM integration not yet implemented: %s / %s", userID, title)
}

var Module = fx.Provide(NewPushNotifier)
