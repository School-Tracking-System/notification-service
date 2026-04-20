package nats

import (
	"context"
	"encoding/json"

	"github.com/fercho/school-tracking/services/notification/internal/core/ports/services"
	"github.com/fercho/school-tracking/services/notification/internal/infrastructure/clients"
	fleetpb "github.com/fercho/school-tracking/proto/gen/fleet/v1"
	nats "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	SubjectTripStarted         = "trip.started"
	SubjectTripEnded           = "trip.ended"
	SubjectStudentBoarded      = "trip.student.boarded"
	SubjectStudentExited       = "trip.student.exited"
	SubjectSchoolReceptionDone = "trip.school.reception.done"
	SubjectStudentAssigned     = "fleet.student.assigned"
)

type tripStartedPayload struct {
	TripID   string `json:"trip_id"`
	DriverID string `json:"driver_id"`
}

type studentEventPayload struct {
	TripID    string `json:"trip_id"`
	StudentID string `json:"student_id"`
}

// StartSubscribers registers all NATS event subscriptions for the notification service.
func StartSubscribers(nc *nats.Conn, svc services.NotificationService, fleetClients clients.FleetClients, log *zap.Logger) error {
	subscriptions := []struct {
		subject string
		handler func([]byte)
	}{
		{
			SubjectTripStarted,
			func(data []byte) {
				var p tripStartedPayload
				if err := json.Unmarshal(data, &p); err != nil {
					log.Warn("Failed to parse trip.started", zap.Error(err))
					return
				}
				// Notify driver that their trip has started
				_, _ = svc.SendPush(context.Background(), services.SendPushRequest{
					UserID: p.DriverID,
					Title:  "Trip Started",
					Body:   "Your route has started. Drive safely!",
				})
			},
		},
		{
			SubjectStudentBoarded,
			func(data []byte) {
				var p studentEventPayload
				if err := json.Unmarshal(data, &p); err != nil {
					log.Warn("Failed to parse student.boarded", zap.Error(err))
					return
				}
				
				res, err := fleetClients.GuardianService.GetGuardiansByStudent(context.Background(), &fleetpb.GetGuardiansByStudentRequest{StudentId: p.StudentID})
				if err != nil {
					log.Error("Failed to fetch guardians", zap.Error(err), zap.String("student_id", p.StudentID))
					return
				}

				for _, g := range res.Guardians {
					_, _ = svc.SendPush(context.Background(), services.SendPushRequest{
						UserID: g.UserId,
						Title:  "Student Boarded",
						Body:   "Your child has boarded the bus",
					})
				}
				
				log.Info("Student boarded — guardian notification sent",
					zap.String("student_id", p.StudentID),
				)
			},
		},
		{
			SubjectStudentExited,
			func(data []byte) {
				var p studentEventPayload
				if err := json.Unmarshal(data, &p); err != nil {
					log.Warn("Failed to parse student.exited", zap.Error(err))
					return
				}
				
				res, err := fleetClients.GuardianService.GetGuardiansByStudent(context.Background(), &fleetpb.GetGuardiansByStudentRequest{StudentId: p.StudentID})
				if err != nil {
					log.Error("Failed to fetch guardians", zap.Error(err), zap.String("student_id", p.StudentID))
					return
				}

				for _, g := range res.Guardians {
					_, _ = svc.SendPush(context.Background(), services.SendPushRequest{
						UserID: g.UserId,
						Title:  "Student Exited",
						Body:   "Your child has exited the bus",
					})
				}

				log.Info("Student exited — guardian notification sent",
					zap.String("student_id", p.StudentID),
				)
			},
		},
		{
			SubjectSchoolReceptionDone,
			func(data []byte) {
				var p studentEventPayload
				if err := json.Unmarshal(data, &p); err != nil {
					log.Warn("Failed to parse school.reception.done", zap.Error(err))
					return
				}
				
				res, err := fleetClients.GuardianService.GetGuardiansByStudent(context.Background(), &fleetpb.GetGuardiansByStudentRequest{StudentId: p.StudentID})
				if err != nil {
					log.Error("Failed to fetch guardians", zap.Error(err), zap.String("student_id", p.StudentID))
					return
				}

				for _, g := range res.Guardians {
					_, _ = svc.SendPush(context.Background(), services.SendPushRequest{
						UserID: g.UserId,
						Title:  "School Reception",
						Body:   "Your child has safely arrived at the school",
					})
				}

				log.Info("School received student — guardian notification sent",
					zap.String("student_id", p.StudentID),
				)
			},
		},
	}

	for _, s := range subscriptions {
		s := s
		_, err := nc.Subscribe(s.subject, func(msg *nats.Msg) { s.handler(msg.Data) })
		if err != nil {
			return err
		}
	}

	log.Info("Notification NATS subscribers started")
	return nil
}
