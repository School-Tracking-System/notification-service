package repositories

import (
	"context"

	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/google/uuid"
)

// NotificationRepository defines the persistence contract for Notification log records.
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	Update(ctx context.Context, n *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	ListByUser(ctx context.Context, userID uuid.UUID, status *domain.NotificationStatus, limit, offset int) ([]*domain.Notification, int, error)
	ListFailed(ctx context.Context) ([]*domain.Notification, error)
}
