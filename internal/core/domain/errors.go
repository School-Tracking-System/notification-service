package domain

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrInvalidNotification  = errors.New("invalid notification data")
	ErrDeliveryFailed       = errors.New("notification delivery failed")
)
