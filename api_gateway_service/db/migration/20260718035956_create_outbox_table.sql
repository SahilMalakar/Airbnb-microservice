-- +goose Up
-- +goose StatementBegin
CREATE TABLE outbox (
    id                BIGSERIAL PRIMARY KEY,
    event_id          UUID NOT NULL DEFAULT gen_random_uuid(),
    event_type        VARCHAR(255) NOT NULL,
    aggregate_type    VARCHAR(255) NOT NULL,
    aggregate_id      VARCHAR(255) NOT NULL,
    aggregate_version BIGINT NOT NULL,
    schema_version    INT NOT NULL DEFAULT 1,
    payload           JSONB NOT NULL,
    correlation_id    UUID,
    occurred_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_outbox_unprocessed ON outbox (created_at) WHERE processed_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_outbox_unprocessed;
DROP TABLE IF EXISTS outbox;
-- +goose StatementEnd