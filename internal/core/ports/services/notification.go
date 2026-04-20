package services

import (
	"context"

	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/google/uuid"
)

// NotificationService defines business logic for delivering notifications.
type NotificationService interface {
	SendPush(ctx context.Context, req SendPushRequest) (*domain.Notification, error)
	SendSMS(ctx context.Context, req SendSMSRequest) (*domain.Notification, error)
	GetNotification(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListNotifications(ctx context.Context, req ListNotificationsRequest) ([]*domain.Notification, int, error)
	RetryFailed(ctx context.Context) (int, error)
}

// SendPushRequest holds data for a push notification delivery.
type SendPushRequest struct {
	UserID string
	Title  string
	Body   string
	Data   string
}

// SendSMSRequest holds data for an SMS delivery.
type SendSMSRequest struct {
	UserID string
	Phone  string
	Body   string
}

// ListNotificationsRequest holds filters for listing notifications.
type ListNotificationsRequest struct {
	UserID string
	Status *domain.NotificationStatus
	Limit  int
	Offset int
}
