package resources

import "context"

// EventPublisher defines the contract for publishing domain events.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// PushNotifier defines the contract for sending FCM push notifications.
type PushNotifier interface {
	Send(ctx context.Context, userID, title, body, data string) error
}

// SMSSender defines the contract for sending SMS via Twilio.
type SMSSender interface {
	Send(ctx context.Context, toPhone, body string) error
}
