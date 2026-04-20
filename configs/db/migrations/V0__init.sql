-- Notification DB — Extensiones y tipos base
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE notification_type   AS ENUM ('push', 'sms');
CREATE TYPE notification_status AS ENUM ('pending', 'sent', 'failed');
