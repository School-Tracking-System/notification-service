package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// NotificationType classifies the delivery channel.
type NotificationType string

const (
	NotificationTypePush NotificationType = "push"
	NotificationTypeSMS  NotificationType = "sms"
)

// NotificationStatus reflects the delivery lifecycle.
type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusFailed  NotificationStatus = "failed"
)

// Notification is the log record for every attempted delivery.
type Notification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Type      NotificationType
	Channel   string // "fcm" | "twilio"
	Title     string
	Body      string
	Data      string // JSON payload for FCM
	Status    NotificationStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewNotificationParams holds the data required to create a Notification record.
type NewNotificationParams struct {
	UserID  uuid.UUID
	Type    NotificationType
	Channel string
	Title   string
	Body    string
	Data    string
}

// NewNotification creates a Notification in the pending state.
func NewNotification(p NewNotificationParams) (*Notification, error) {
	if p.UserID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
	if p.Body == "" {
		return nil, errors.New("notification body is required")
	}
	now := time.Now().UTC()
	return &Notification{
		ID:        uuid.New(),
		UserID:    p.UserID,
		Type:      p.Type,
		Channel:   p.Channel,
		Title:     p.Title,
		Body:      p.Body,
		Data:      p.Data,
		Status:    NotificationStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// MarkSent transitions the notification to sent.
func (n *Notification) MarkSent() {
	n.Status = NotificationStatusSent
	n.UpdatedAt = time.Now().UTC()
}

// MarkFailed transitions the notification to failed.
func (n *Notification) MarkFailed() {
	n.Status = NotificationStatusFailed
	n.UpdatedAt = time.Now().UTC()
}
