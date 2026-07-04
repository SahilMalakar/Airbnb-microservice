package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

// UserRepository defines the data-access operations available for users.
type UserRepository interface {
	Create(name, email, hashPassword string,
		role models.Role) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)

	// GetUsers() ([]User, error)
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

func (u *UserRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return user, nil
}

// Create persists a new user record.
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
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user already exists")
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserRepositoryImpl) GetUserByID(id int64) (*models.User, error) {
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return user, nil
}
