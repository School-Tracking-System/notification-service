-- Notification table — ENUMs defined in V0__init.sql

CREATE TABLE notifications (
    id         UUID                 PRIMARY KEY,
    user_id    UUID                 NOT NULL,
    type       notification_type    NOT NULL,
    channel    VARCHAR(20)          NOT NULL,
    title      TEXT                 NOT NULL DEFAULT '',
    body       TEXT                 NOT NULL,
    data       TEXT                 NOT NULL DEFAULT '',
    status     notification_status  NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id  ON notifications (user_id);
CREATE INDEX idx_notifications_status   ON notifications (status);
CREATE INDEX idx_notifications_created  ON notifications (created_at DESC);
