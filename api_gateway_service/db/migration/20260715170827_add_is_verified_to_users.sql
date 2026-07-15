-- +goose Up
ALTER TABLE users ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT false;
UPDATE users SET is_verified = true WHERE email IN ('admin@gmail.com', 'host@gmail.com');

-- +goose Down
ALTER TABLE users DROP COLUMN is_verified;
