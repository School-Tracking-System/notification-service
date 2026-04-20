package notification

import (
	"context"
	"fmt"

	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/repositories"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/resources"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/services"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type notificationService struct {
	repo    repositories.NotificationRepository
	push    resources.PushNotifier
	sms     resources.SMSSender
	log     *zap.Logger
}

// NewNotificationService creates the notification delivery service.
func NewNotificationService(
	repo repositories.NotificationRepository,
	push resources.PushNotifier,
	sms resources.SMSSender,
	log *zap.Logger,
) services.NotificationService {
	return &notificationService{
		repo: repo,
		push: push,
		sms:  sms,
		log:  log,
	}
}

func (s *notificationService) SendPush(ctx context.Context, req services.SendPushRequest) (*domain.Notification, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	n, err := domain.NewNotification(domain.NewNotificationParams{
		UserID:  userID,
		Type:    domain.NotificationTypePush,
		Channel: "fcm",
		Title:   req.Title,
		Body:    req.Body,
		Data:    req.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("failed to log notification: %w", err)
	}

	if err := s.push.Send(ctx, req.UserID, req.Title, req.Body, req.Data); err != nil {
		s.log.Warn("FCM delivery failed", zap.String("user_id", req.UserID), zap.Error(err))
		n.MarkFailed()
	} else {
		n.MarkSent()
	}

	_ = s.repo.Update(ctx, n)
	return n, nil
}

func (s *notificationService) SendSMS(ctx context.Context, req services.SendSMSRequest) (*domain.Notification, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	n, err := domain.NewNotification(domain.NewNotificationParams{
		UserID:  userID,
		Type:    domain.NotificationTypeSMS,
		Channel: "twilio",
		Body:    req.Body,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid notification: %w", err)
	}

	if err := s.repo.Create(ctx, n); err != nil {
		return nil, fmt.Errorf("failed to log notification: %w", err)
	}

	if err := s.sms.Send(ctx, req.Phone, req.Body); err != nil {
		s.log.Warn("Twilio SMS delivery failed", zap.String("phone", req.Phone), zap.Error(err))
		n.MarkFailed()
	} else {
		n.MarkSent()
	}

	_ = s.repo.Update(ctx, n)
	return n, nil
}

func (s *notificationService) GetNotification(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *notificationService) ListNotifications(ctx context.Context, req services.ListNotificationsRequest) ([]*domain.Notification, int, error) {
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user_id: %w", err)
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	return s.repo.ListByUser(ctx, userID, req.Status, req.Limit, req.Offset)
}

func (s *notificationService) RetryFailed(ctx context.Context) (int, error) {
	failed, err := s.repo.ListFailed(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to load failed notifications: %w", err)
	}

	retried := 0
	for _, n := range failed {
		var deliveryErr error
		switch n.Type {
		case domain.NotificationTypePush:
			deliveryErr = s.push.Send(ctx, n.UserID.String(), n.Title, n.Body, n.Data)
		case domain.NotificationTypeSMS:
			deliveryErr = s.sms.Send(ctx, "", n.Body)
		}

		if deliveryErr == nil {
			n.MarkSent()
			_ = s.repo.Update(ctx, n)
			retried++
		}
	}

	s.log.Info("Retried failed notifications", zap.Int("count", retried))
	return retried, nil
}

// NotificationModule provides the notification service to the fx dependency graph.
var NotificationModule = fx.Provide(NewNotificationService)
