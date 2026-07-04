package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
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

// ErrEmailNotFound is returned when no user matches the given email.
var ErrEmailNotFound = errors.New("email not found")

func (u *UserRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	fmt.Printf("[DEBUG] UserRepository: GetUserByEmail database query started for email %q\n", email)
	query := `SELECT id, name, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt); err != nil {
		fmt.Printf("[DEBUG] UserRepository: GetUserByEmail Scan error: %v\n", err)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	fmt.Printf("[DEBUG] UserRepository: GetUserByEmail Scan success: found user ID %d\n", user.ID)
	return user, nil
}

// ErrEmailAlreadyExists is returned when an insert violates the unique
// constraint on users.email.
var ErrEmailAlreadyExists = errors.New("email already exists")

// Create persists a new user record.
func (u *UserRepositoryImpl) Create(
	name, email, hashPassword string,
	role models.Role) (*models.User, error) {

	query := `INSERT INTO users (name, email, password, role) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, role, created_at, updated_at`

	var user models.User
	err := u.db.QueryRow(query, name, email, hashPassword, role).Scan(
		&user.ID, &user.Name, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" { // unique_violation
			return nil, ErrEmailAlreadyExists
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("creating user: %w", err)
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