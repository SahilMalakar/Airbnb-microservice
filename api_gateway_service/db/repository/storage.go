package db

import "database/sql"

// Storage aggregates all repository interfaces the application depends on,
// giving a single injection point for data access.
type Storage struct {
	Repo UserRepository
}

// NewStorage wires up and returns a Storage with concrete repository
// implementations.
func NewStorage(conn *sql.DB) *Storage {
	return &Storage{
		Repo: &UserRepositoryImpl{db: conn},
	}
}
