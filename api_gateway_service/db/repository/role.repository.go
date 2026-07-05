package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

var ErrRoleNotFound = errors.New("role not found")
var ErrRoleNameTaken = errors.New("role name already exists")

type RoleRepository interface {
	GetRoleById(id int64) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
	GetAllRoles() ([]*models.Role, error)
	CreateRole(name string, description string) (*models.Role, error)
	UpdateRoleById(id int64, name string, description string) (*models.Role, error)
	DeleteRoleById(id int64) error
}

type RoleRepositoryImpl struct {
	db *sql.DB
}

func NewRoleRepository(conn *sql.DB) RoleRepository {
	return &RoleRepositoryImpl{
		db: conn,
	}
}

func (r *RoleRepositoryImpl) GetRoleById(id int64) (*models.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var role models.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) GetRoleByName(name string) (*models.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles WHERE name = $1`

	row := r.db.QueryRow(query, name)

	var role models.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) GetAllRoles() ([]*models.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM roles`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepositoryImpl) CreateRole(name string, description string) (*models.Role, error) {
	query := `INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id, name, description, created_at, updated_at`

	row := r.db.QueryRow(query, name, description)

	var role models.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && strings.Contains(pqErr.Constraint, "name") {
			return nil, ErrRoleNameTaken
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) UpdateRoleById(id int64, name string, description string) (*models.Role, error) {
	query := `
		UPDATE roles
		SET name = $2, description = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`

	row := r.db.QueryRow(query, id, name, description)

	var role models.Role
	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRoleNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && strings.Contains(pqErr.Constraint, "name") {
			return nil, ErrRoleNameTaken
		}
		return nil, err
	}
	return &role, nil
}

func (r *RoleRepositoryImpl) DeleteRoleById(id int64) error {
	query := `DELETE FROM roles WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRoleNotFound
	}
	return nil
}
