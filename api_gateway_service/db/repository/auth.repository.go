package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	GetUsersSnapshot(ctx context.Context, cursor int64, limit int) ([]*models.User, int64, error)
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
var ErrEmailAlreadyExists = errors.New("email already exists")

func (u *UserRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password, is_verified, version, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetUserByEmailBasic(email string) (*models.User, error) {
	query := `SELECT id, name, email, is_verified, version, created_at, updated_at FROM users WHERE email = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, email)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmailNotFound
		}
		return nil, fmt.Errorf("querying user by email basic: %w", err)
	}

	return user, nil
}

// Create persists a new user record and inserts a UserCreated outbox entry in the same transaction.
func (u *UserRepositoryImpl) Create(
	name, email, hashPassword string,
) (*models.User, error) {
	tx, err := u.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("beginning create user transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO users (name, email, password, version) VALUES ($1, $2, $3, 1) RETURNING id, name, email, is_verified, version, created_at, updated_at`
	var user models.User

	err = tx.QueryRow(query, name, email, hashPassword).Scan(
		&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("inserting user row: %w", err)
	}

	// Prepare payload
	payloadData := map[string]interface{}{
		"userId":     user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"isVerified": user.IsVerified,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return nil, fmt.Errorf("marshaling UserCreated payload: %w", err)
	}

	// Insert into outbox
	outboxQuery := `INSERT INTO outbox (event_id, event_type, aggregate_type, aggregate_id, aggregate_version, schema_version, payload, occurred_at) 
	                VALUES ($1, $2, $3, $4, $5, 1, $6, NOW())`

	eventId := uuid.New().String()
	aggregateIdStr := fmt.Sprintf("%d", user.ID)

	_, err = tx.Exec(outboxQuery, eventId, "UserCreated", "User", aggregateIdStr, user.Version, string(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("inserting UserCreated outbox entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("committing create user transaction: %w", err)
	}

	return &user, nil
}

func (u *UserRepositoryImpl) GetUserByID(id int64) (*models.User, error) {
	query := `SELECT id, name, email, password, is_verified, version, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetUserByIDBasic(id int64) (*models.User, error) {
	query := `SELECT id, name, email, is_verified, version, created_at, updated_at FROM users WHERE id = $1`
	user := &models.User{}

	row := u.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, err
	}

	return user, nil
}

func (u *UserRepositoryImpl) GetAllUsers() ([]*models.User, error) {
	query := `SELECT id, name, email, is_verified, version, created_at, updated_at FROM users`
	rows, err := u.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.Version, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

// MarkUserVerified flips is_verified to true and enqueues a UserVerified outbox event in the same transaction.
func (u *UserRepositoryImpl) MarkUserVerified(id int64) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning mark user verified transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET is_verified = true, version = version + 1, updated_at = now() WHERE id = $1 RETURNING name, email, version`

	var name, email string
	var version int64
	err = tx.QueryRow(query, id).Scan(&name, &email, &version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user with id %d not found", id)
		}
		return fmt.Errorf("updating user to verified: %w", err)
	}

	// Prepare payload
	payloadData := map[string]interface{}{
		"userId":     id,
		"name":       name,
		"email":      email,
		"isVerified": true,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("marshaling UserVerified payload: %w", err)
	}

	// Insert into outbox
	outboxQuery := `INSERT INTO outbox (event_id, event_type, aggregate_type, aggregate_id, aggregate_version, schema_version, payload, occurred_at) 
	                VALUES ($1, $2, $3, $4, $5, 1, $6, NOW())`

	eventId := uuid.New().String()
	aggregateIdStr := fmt.Sprintf("%d", id)

	_, err = tx.Exec(outboxQuery, eventId, "UserVerified", "User", aggregateIdStr, version, string(payloadBytes))
	if err != nil {
		return fmt.Errorf("inserting UserVerified outbox entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing mark user verified transaction: %w", err)
	}

	return nil
}

// UpdatePassword replaces the user's hashed password and enqueues a UserUpdated outbox event in the same transaction.
func (u *UserRepositoryImpl) UpdatePassword(id int64, hashedPassword string) error {
	tx, err := u.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning update password transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE users SET password = $1, version = version + 1, updated_at = now() WHERE id = $2 RETURNING name, email, version`

	var name, email string
	var version int64
	err = tx.QueryRow(query, hashedPassword, id).Scan(&name, &email, &version)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("user with id %d not found", id)
		}
		return fmt.Errorf("updating user password: %w", err)
	}

	// Prepare payload
	payloadData := map[string]interface{}{
		"userId": id,
		"name":   name,
		"email":  email,
	}
	payloadBytes, err := json.Marshal(payloadData)
	if err != nil {
		return fmt.Errorf("marshaling UserUpdated payload: %w", err)
	}

	// Insert into outbox
	outboxQuery := `INSERT INTO outbox (event_id, event_type, aggregate_type, aggregate_id, aggregate_version, schema_version, payload, occurred_at) 
	                VALUES ($1, $2, $3, $4, $5, 1, $6, NOW())`

	eventId := uuid.New().String()
	aggregateIdStr := fmt.Sprintf("%d", id)

	_, err = tx.Exec(outboxQuery, eventId, "UserUpdated", "User", aggregateIdStr, version, string(payloadBytes))
	if err != nil {
		return fmt.Errorf("inserting UserUpdated outbox entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing update password transaction: %w", err)
	}

	return nil
}

func (u *UserRepositoryImpl) GetUsersSnapshot(ctx context.Context, cursor int64, limit int) ([]*models.User, int64, error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var outboxCursor int64
	err = tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(id), 0) FROM outbox").Scan(&outboxCursor)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, email, is_verified, version FROM users WHERE id > $1 ORDER BY id ASC LIMIT $2`
	rows, err := tx.QueryContext(ctx, query, cursor, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.IsVerified, &user.Version); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}

	if err := tx.Commit(); err != nil {
		return nil, 0, err
	}

	return users, outboxCursor, nil
}
