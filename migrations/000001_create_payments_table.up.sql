CREATE TABLE IF NOT EXISTS payments (
    id           UUID         PRIMARY KEY,
    amount_cents BIGINT       NOT NULL CHECK (amount_cents > 0),
    currency     CHAR(3)      NOT NULL,
    method       VARCHAR(20)  NOT NULL,
    status       VARCHAR(20)  NOT NULL,
    description  VARCHAR(255) NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_payments_created_at_id ON payments (created_at DESC, id DESC);
