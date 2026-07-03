package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

// UserRepository defines the data-access operations available for users.
type UserRepository interface {
	Create(*models.User) (*models.User, error)
	// GetUsers() ([]User, error)
	GetUserByID(id int64) (*models.User, error)
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
func (u *UserRepositoryImpl) Create(user *models.User) (*models.User, error) {
	//TODO: lets make one dummy db query
	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3)`

	result, err := u.db.Exec(query, "sahil malakar", "sahil@gmail.com", "sahil@123456")
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()

	if rowsAffected == 0 {
		fmt.Println("failed to insert", user)
		return nil, errors.New("failed to insert user")
	}

	fmt.Println("successfully inserted in database", rowsAffected)
	return user, nil
}
