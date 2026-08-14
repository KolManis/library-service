-- +goose Up
CREATE TABLE outbox (
    id           uuid PRIMARY KEY,
    event_type   text NOT NULL,
    payload      jsonb NOT NULL,
    created_at   timestamptz NOT NULL,
    published_at timestamptz
);
CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS outbox;