package db

import (
	"database/sql"
	"fmt"
)

// User represents a single row from the users table.
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}

// UserRepository defines the data-access operations available for users.
type UserRepository interface {
	Create() error
	// GetUsers() ([]User, error)
	// GetUserByID(id int64) (*User, error)
	// Update(id int64, user *User) error
	// Delete(id int64) error
}

// UserRepositoryImpl is the concrete, database-backed implementation of
// UserRepository.
type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(conn *sql.DB) UserRepository {
	return &UserRepositoryImpl{db: conn}
}

// Create persists a new user record.
// TODO: not yet implemented — currently a no-op.
func (u *UserRepositoryImpl) Create() error {
	//TODO: lets make one dummy db query
	query := `SELECT 1;`

	rows, err := u.db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("successfully queried database", rows)
	return nil
}