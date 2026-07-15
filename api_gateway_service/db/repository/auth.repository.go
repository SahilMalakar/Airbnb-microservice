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
	Create(name, email, hashPassword string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByEmailBasic(email string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	GetUserByIDBasic(id int64) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	MarkUserVerified(id int64) error
	UpdatePassword(id int64, hashedPassword string) error
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
	query := `SELECT id, name, email, password, is_verified, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetUserByEmailBasic(email string) (*models.User, error) {
	query := `SELECT id, name, email, is_verified, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return user, nil
}

// ErrEmailAlreadyExists is returned when an insert violates the unique
// constraint on users.email.
var ErrEmailAlreadyExists = errors.New("email already exists")

// Create persists a new user record.
func (u *UserRepositoryImpl) Create(
	name, email, hashPassword string,
) (*models.User, error) {
	query := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id, name, email, is_verified, created_at, updated_at`

	var user models.User

	err := u.db.QueryRow(query, name, email, hashPassword).Scan(
		&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
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
	query := `SELECT id, name, email, password, is_verified, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetUserByIDBasic(id int64) (*models.User, error) {
	query := `SELECT id, name, email, is_verified, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetAllUsers() ([]*models.User, error) {
	query := `SELECT id, name, email, is_verified, created_at, updated_at FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// MarkUserVerified flips is_verified to true for the given user.
func (u *UserRepositoryImpl) MarkUserVerified(id int64) error {
	query := `UPDATE users SET is_verified = true, updated_at = now() WHERE id = $1`
	result, err := u.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("marking user verified: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}

// UpdatePassword replaces the user's hashed password.
func (u *UserRepositoryImpl) UpdatePassword(id int64, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = now() WHERE id = $2`
	result, err := u.db.Exec(query, hashedPassword, id)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}
