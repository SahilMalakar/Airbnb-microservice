package db

// Storage aggregates all repository interfaces the application depends on,
// giving a single injection point for data access.
type Storage struct {
	UserRepository UserRepository
}

// NewStorage wires up and returns a Storage with concrete repository
// implementations.
func NewStorage() *Storage {
	return &Storage{
		UserRepository: &UserRepositoryImpl{},
	}
}