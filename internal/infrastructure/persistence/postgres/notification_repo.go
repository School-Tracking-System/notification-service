package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/fercho/school-tracking/services/notification/internal/core/domain"
	"github.com/fercho/school-tracking/services/notification/internal/core/ports/repositories"
	"github.com/fercho/school-tracking/services/notification/pkg/env"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NewDatabase opens a PostgreSQL connection pool.
func NewDatabase(lc fx.Lifecycle, cfg *env.Config, log *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open notification database: %w", err)
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("notification database ping failed: %w", err)
			}
			log.Info("Connected to notification database")
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("Closing notification database connection")
			return db.Close()
		},
	})
	return db, nil
}

type notificationRepo struct {
	db  *sql.DB
	log *zap.Logger
}

func NewNotificationRepository(db *sql.DB, log *zap.Logger) repositories.NotificationRepository {
	return &notificationRepo{db: db, log: log}
}

func (r *notificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, channel, title, body, data, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		n.ID, n.UserID, n.Type, n.Channel, n.Title, n.Body, n.Data, n.Status, n.CreatedAt, n.UpdatedAt,
	)
	return err
}

func (r *notificationRepo) Update(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET status=$1, updated_at=$2 WHERE id=$3`,
		n.Status, n.UpdatedAt, n.ID,
	)
	return err
}

func (r *notificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, type, channel, title, body, data, status, created_at, updated_at
		FROM notifications WHERE id = $1`, id,
	)
	return scanNotification(row)
}

func (r *notificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, status *domain.NotificationStatus, limit, offset int) ([]*domain.Notification, int, error) {
	query := `SELECT id, user_id, type, channel, title, body, data, status, created_at, updated_at FROM notifications WHERE user_id = $1`
	args := []any{userID}
	i := 2
	if status != nil {
		query += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, *status)
		i++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ns []*domain.Notification
	for rows.Next() {
		n, err := scanNotificationRow(rows)
		if err != nil {
			return nil, 0, err
		}
		ns = append(ns, n)
	}

	var total int
	_ = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notifications WHERE user_id=$1", userID).Scan(&total)
	return ns, total, nil
}

func (r *notificationRepo) ListFailed(ctx context.Context) ([]*domain.Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, type, channel, title, body, data, status, created_at, updated_at
		FROM notifications WHERE status = 'failed' ORDER BY created_at ASC LIMIT 100`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ns []*domain.Notification
	for rows.Next() {
		n, err := scanNotificationRow(rows)
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	return ns, nil
}

func scanNotification(row *sql.Row) (*domain.Notification, error) {
	n := &domain.Notification{}
	err := row.Scan(&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Title, &n.Body, &n.Data, &n.Status, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("failed to scan notification: %w", err)
	}
	return n, nil
}

func scanNotificationRow(rows *sql.Rows) (*domain.Notification, error) {
	n := &domain.Notification{}
	err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Channel, &n.Title, &n.Body, &n.Data, &n.Status, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan notification row: %w", err)
	}
	return n, nil
}

// Module provides all PostgreSQL dependencies to the fx graph.
var Module = fx.Options(
	fx.Provide(NewDatabase),
	fx.Provide(NewNotificationRepository),
)
