-- +goose Up
ALTER TABLE users DROP COLUMN role;
DROP TYPE IF EXISTS user_role;

-- +goose Down
CREATE TYPE user_role AS ENUM ('user', 'admin', 'host');
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'user';