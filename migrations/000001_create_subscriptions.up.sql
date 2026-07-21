CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT NOT NULL,
    from_iata CHAR(3) NOT NULL,
    to_iata CHAR(3) NOT NULL,
    date_from DATE NOT NULL,
    date_to DATE NOT NULL,
    threshold NUMERIC(10, 2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT date_range_valid CHECK (date_to >= date_from)
);

CREATE INDEX idx_subscriptions_telegram_id ON subscriptions (telegram_id);
