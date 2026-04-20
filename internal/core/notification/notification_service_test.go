package notification

import (
	"context"
	"errors"
	"testing"

	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/mocks"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/services"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newSvc(t *testing.T) (services.NotificationService, *mocks.MockNotificationRepository, *mocks.MockPushNotifier, *mocks.MockSMSSender) {
	repo := mocks.NewMockNotificationRepository(t)
	push := mocks.NewMockPushNotifier(t)
	sms := mocks.NewMockSMSSender(t)
	svc := NewNotificationService(repo, push, sms, zap.NewNop())
	return svc, repo, push, sms
}

func TestSendPush(t *testing.T) {
	ctx := context.Background()
	validUserID := uuid.New()

	tests := []struct {
		name       string
		req        services.SendPushRequest
		setupRepo  func(*mocks.MockNotificationRepository)
		setupPush  func(*mocks.MockPushNotifier)
		wantErr    bool
		errMsg     string
		wantStatus domain.NotificationStatus
	}{
		{
			name: "success — FCM delivers and status is sent",
			req:  services.SendPushRequest{UserID: validUserID.String(), Title: "Trip Started", Body: "Drive safely!"},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("Create", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Type == domain.NotificationTypePush && n.Channel == "fcm" && n.Status == domain.NotificationStatusPending
				})).Return(nil)
				r.On("Update", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Status == domain.NotificationStatusSent
				})).Return(nil)
			},
			setupPush: func(p *mocks.MockPushNotifier) {
				p.On("Send", ctx, validUserID.String(), "Trip Started", "Drive safely!", "").Return(nil)
			},
			wantStatus: domain.NotificationStatusSent,
		},
		{
			name: "delivery failure — status marked as failed, no error returned",
			req:  services.SendPushRequest{UserID: validUserID.String(), Title: "Hi", Body: "Body"},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("Create", ctx, mock.Anything).Return(nil)
				r.On("Update", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Status == domain.NotificationStatusFailed
				})).Return(nil)
			},
			setupPush: func(p *mocks.MockPushNotifier) {
				p.On("Send", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("fcm error"))
			},
			wantStatus: domain.NotificationStatusFailed,
		},
		{
			name:      "invalid user_id rejected before any IO",
			req:       services.SendPushRequest{UserID: "not-a-uuid", Body: "Hello"},
			setupRepo: func(r *mocks.MockNotificationRepository) {},
			setupPush: func(p *mocks.MockPushNotifier) {},
			wantErr:   true,
			errMsg:    "invalid user_id",
		},
		{
			name: "repo Create failure propagated",
			req:  services.SendPushRequest{UserID: validUserID.String(), Body: "Hello"},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("Create", ctx, mock.Anything).Return(errors.New("db error"))
			},
			setupPush: func(p *mocks.MockPushNotifier) {},
			wantErr:   true,
			errMsg:    "failed to log notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, push, _ := newSvc(t)
			tt.setupRepo(repo)
			tt.setupPush(push)

			n, err := svc.SendPush(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, n.Status)
		})
	}
}

func TestSendSMS(t *testing.T) {
	ctx := context.Background()
	validUserID := uuid.New()
	phone := "+573001234567"

	tests := []struct {
		name       string
		req        services.SendSMSRequest
		setupRepo  func(*mocks.MockNotificationRepository)
		setupSMS   func(*mocks.MockSMSSender)
		wantErr    bool
		errMsg     string
		wantStatus domain.NotificationStatus
	}{
		{
			name: "success — Twilio delivers and status is sent",
			req:  services.SendSMSRequest{UserID: validUserID.String(), Phone: phone, Body: "Your student has boarded."},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("Create", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Type == domain.NotificationTypeSMS && n.Channel == "twilio"
				})).Return(nil)
				r.On("Update", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Status == domain.NotificationStatusSent
				})).Return(nil)
			},
			setupSMS: func(s *mocks.MockSMSSender) {
				s.On("Send", ctx, phone, "Your student has boarded.").Return(nil)
			},
			wantStatus: domain.NotificationStatusSent,
		},
		{
			name: "delivery failure — status marked as failed, no error returned",
			req:  services.SendSMSRequest{UserID: validUserID.String(), Phone: phone, Body: "Hi"},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("Create", ctx, mock.Anything).Return(nil)
				r.On("Update", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.Status == domain.NotificationStatusFailed
				})).Return(nil)
			},
			setupSMS: func(s *mocks.MockSMSSender) {
				s.On("Send", ctx, phone, mock.Anything).Return(errors.New("twilio error"))
			},
			wantStatus: domain.NotificationStatusFailed,
		},
		{
			name:      "invalid user_id rejected before any IO",
			req:       services.SendSMSRequest{UserID: "bad-id", Phone: phone, Body: "hi"},
			setupRepo: func(r *mocks.MockNotificationRepository) {},
			setupSMS:  func(s *mocks.MockSMSSender) {},
			wantErr:   true,
			errMsg:    "invalid user_id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, sms := newSvc(t)
			tt.setupRepo(repo)
			tt.setupSMS(sms)

			n, err := svc.SendSMS(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, n.Status)
		})
	}
}

func TestGetNotification(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	expected := &domain.Notification{ID: id, Status: domain.NotificationStatusSent}

	tests := []struct {
		name      string
		id        uuid.UUID
		setupRepo func(*mocks.MockNotificationRepository)
		wantErr   bool
		errTarget error
	}{
		{
			name: "found — returns notification",
			id:   id,
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("GetByID", ctx, id).Return(expected, nil)
			},
		},
		{
			name: "not found — propagates domain error",
			id:   uuid.New(),
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("GetByID", ctx, mock.Anything).Return(nil, domain.ErrNotificationNotFound)
			},
			wantErr:   true,
			errTarget: domain.ErrNotificationNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _ := newSvc(t)
			tt.setupRepo(repo)

			n, err := svc.GetNotification(ctx, tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errTarget != nil {
					assert.Equal(t, tt.errTarget, err)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, expected, n)
		})
	}
}

func TestListNotifications(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	statusFilter := domain.NotificationStatusFailed

	tests := []struct {
		name      string
		req       services.ListNotificationsRequest
		setupRepo func(*mocks.MockNotificationRepository)
		wantLen   int
		wantTotal int
		wantErr   bool
	}{
		{
			name: "all notifications for user",
			req:  services.ListNotificationsRequest{UserID: userID.String()},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("ListByUser", ctx, userID, (*domain.NotificationStatus)(nil), 20, 0).
					Return([]*domain.Notification{
						{ID: uuid.New(), UserID: userID, Status: domain.NotificationStatusSent},
						{ID: uuid.New(), UserID: userID, Status: domain.NotificationStatusFailed},
					}, 2, nil)
			},
			wantLen:   2,
			wantTotal: 2,
		},
		{
			name: "filtered by status",
			req:  services.ListNotificationsRequest{UserID: userID.String(), Status: &statusFilter},
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("ListByUser", ctx, userID, &statusFilter, 20, 0).
					Return([]*domain.Notification{{ID: uuid.New(), Status: statusFilter}}, 1, nil)
			},
			wantLen:   1,
			wantTotal: 1,
		},
		{
			name:      "invalid user_id rejected",
			req:       services.ListNotificationsRequest{UserID: "bad-id"},
			setupRepo: func(r *mocks.MockNotificationRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _ := newSvc(t)
			tt.setupRepo(repo)

			result, total, err := svc.ListNotifications(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result, tt.wantLen)
			assert.Equal(t, tt.wantTotal, total)
		})
	}
}

func TestRetryFailed(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	notifID := uuid.New()

	failedPush := &domain.Notification{
		ID: notifID, UserID: userID,
		Type: domain.NotificationTypePush, Title: "Retry", Body: "Retry body",
		Status: domain.NotificationStatusFailed,
	}

	tests := []struct {
		name      string
		setupRepo func(*mocks.MockNotificationRepository)
		setupPush func(*mocks.MockPushNotifier)
		wantCount int
		wantErr   bool
	}{
		{
			name: "retries failed push and marks as sent",
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("ListFailed", ctx).Return([]*domain.Notification{failedPush}, nil)
				r.On("Update", ctx, mock.MatchedBy(func(n *domain.Notification) bool {
					return n.ID == notifID && n.Status == domain.NotificationStatusSent
				})).Return(nil)
			},
			setupPush: func(p *mocks.MockPushNotifier) {
				p.On("Send", ctx, userID.String(), "Retry", "Retry body", "").Return(nil)
			},
			wantCount: 1,
		},
		{
			name: "repo ListFailed error propagated",
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("ListFailed", ctx).Return(nil, errors.New("db error"))
			},
			setupPush: func(p *mocks.MockPushNotifier) {},
			wantErr:   true,
		},
		{
			name: "no failed notifications returns zero count",
			setupRepo: func(r *mocks.MockNotificationRepository) {
				r.On("ListFailed", ctx).Return([]*domain.Notification{}, nil)
			},
			setupPush: func(p *mocks.MockPushNotifier) {},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, push, _ := newSvc(t)
			tt.setupRepo(repo)
			tt.setupPush(push)

			count, err := svc.RetryFailed(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}
