-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO users (name, email, password) VALUES
('Admin User', 'admin@gmail.com', crypt('admin1234', gen_salt('bf'))),
('Host User', 'host@gmail.com', crypt('host1234', gen_salt('bf')));

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'admin@gmail.com' AND r.name = 'admin';

INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u, roles r
WHERE u.email = 'host@gmail.com' AND r.name = 'host';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM user_roles
WHERE user_id IN (
    SELECT id FROM users WHERE email IN ('admin@gmail.com', 'host@gmail.com')
);

DELETE FROM users WHERE email IN ('admin@gmail.com', 'host@gmail.com');
-- +goose StatementEnd