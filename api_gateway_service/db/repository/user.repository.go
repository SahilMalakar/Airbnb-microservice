package db

import (
	"database/sql"
	"fmt"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

// UserRepository defines the data-access operations available for users.
type UserRepository interface {
	Create(name, email, hashPassword string,
		role models.Role) (*models.User, error)
	// GetUsers() ([]User, error)
	// GetUserByID(id int64) (*models.User, error)
	// Update(id int64, user *User) error
	// Delete(id int64) error
}

// UserRepositoryImpl is the concrete, database-backed implementation of
// UserRepository.
type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(conn *sql.DB) UserRepository {
	return &UserRepositoryImpl{
		db: conn,
	}
}

func (u *UserRepositoryImpl) GetUserByID(id int64) (*models.User, error) {
	query := `SELECT * FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)

	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return nil, err
	}

	return user, nil
}

// Create persists a new user record.
// TODO: not yet implemented — currently a no-op.
func (u *UserRepositoryImpl) Create(
	name, email, hashPassword string,
	role models.Role) (*models.User, error) {

	query := `INSERT INTO users (name, email, password, role) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, role, created_at, updated_at`

	fmt.Println("query string : ", query)
	var user models.User
	err := u.db.QueryRow(query, name, email, hashPassword, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
