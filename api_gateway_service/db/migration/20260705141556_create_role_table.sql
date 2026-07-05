-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (name , description ) VALUES 
('admin' , 'Administrator with full access to system.'),
('user' , 'Regular users with access to view, book, and cancel their own bookings.'),
('host' , 'host members with access to manage bookings and view system activity.');

-- +goose Down
DROP TABLE IF EXISTS roles;
