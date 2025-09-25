-- internal/db/migrations/001_create_subscriptions.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL, -- store as YYYY-MM-01 (first day of month)
    end_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
-- CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
-- CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
-- CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);
