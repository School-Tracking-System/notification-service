package grpc

import (
	"context"
	"errors"

	pb "github.com/fercho/school-tracking/proto/gen/notification/v1"
	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type notificationHandler struct {
	pb.UnimplementedNotificationServiceServer
	service services.NotificationService
	log     *zap.Logger
}

func NewNotificationHandler(service services.NotificationService, log *zap.Logger) pb.NotificationServiceServer {
	return &notificationHandler{service: service, log: log}
}

func (h *notificationHandler) SendPush(ctx context.Context, req *pb.SendPushRequest) (*pb.NotificationResponse, error) {
	n, err := h.service.SendPush(ctx, services.SendPushRequest{
		UserID: req.UserId,
		Title:  req.Title,
		Body:   req.Body,
		Data:   req.Data,
	})
	if err != nil {
		return nil, h.mapError(err)
	}
	return &pb.NotificationResponse{Notification: domainToProto(n)}, nil
}

func (h *notificationHandler) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.NotificationResponse, error) {
	n, err := h.service.SendSMS(ctx, services.SendSMSRequest{
		UserID: req.UserId,
		Phone:  req.Phone,
		Body:   req.Body,
	})
	if err != nil {
		return nil, h.mapError(err)
	}
	return &pb.NotificationResponse{Notification: domainToProto(n)}, nil
}

func (h *notificationHandler) GetNotification(ctx context.Context, req *pb.GetNotificationRequest) (*pb.NotificationResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid notification id")
	}
	n, err := h.service.GetNotification(ctx, id)
	if err != nil {
		return nil, h.mapError(err)
	}
	return &pb.NotificationResponse{Notification: domainToProto(n)}, nil
}

func (h *notificationHandler) ListNotifications(ctx context.Context, req *pb.ListNotificationsRequest) (*pb.ListNotificationsResponse, error) {
	listReq := services.ListNotificationsRequest{
		UserID: req.UserId,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}
	if req.Status != "" {
		s := domain.NotificationStatus(req.Status)
		listReq.Status = &s
	}

	ns, total, err := h.service.ListNotifications(ctx, listReq)
	if err != nil {
		return nil, h.mapError(err)
	}

	var pbNs []*pb.Notification
	for _, n := range ns {
		pbNs = append(pbNs, domainToProto(n))
	}
	return &pb.ListNotificationsResponse{Notifications: pbNs, TotalCount: int32(total)}, nil
}

func (h *notificationHandler) mapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotificationNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidNotification):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		h.log.Error("Unexpected notification error", zap.Error(err))
		return status.Error(codes.Internal, "internal server error")
	}
}

func domainToProto(n *domain.Notification) *pb.Notification {
	return &pb.Notification{
		Id:        n.ID.String(),
		UserId:    n.UserID.String(),
		Type:      string(n.Type),
		Channel:   n.Channel,
		Title:     n.Title,
		Body:      n.Body,
		Status:    string(n.Status),
		CreatedAt: timestamppb.New(n.CreatedAt),
		UpdatedAt: timestamppb.New(n.UpdatedAt),
	}
}
