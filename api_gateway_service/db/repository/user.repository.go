package db

import "database/sql"

// UserRepository defines the data-access operations available for users.
type UserRepository interface {
	Create() error
}

// UserRepositoryImpl is the concrete, database-backed implementation of
// UserRepository.
type UserRepositoryImpl struct {
	db *sql.DB
}

// Create persists a new user record.
// TODO: not yet implemented — currently a no-op.
func (u *UserRepositoryImpl) Create() error {

	return nil
}