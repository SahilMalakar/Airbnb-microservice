-- +goose Up
-- +goose StatementBegin
ALTER TABLE outbox ADD COLUMN claimed_by VARCHAR(255);
ALTER TABLE outbox ADD COLUMN claimed_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE outbox DROP COLUMN IF EXISTS claimed_by;
ALTER TABLE outbox DROP COLUMN IF EXISTS claimed_at;
-- +goose StatementEnd
