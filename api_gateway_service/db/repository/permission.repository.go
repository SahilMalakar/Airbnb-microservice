package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

var ErrPermissionNotFound = errors.New("permission not found")
var ErrPermissionNameTaken = errors.New("permission name already exists")

type PermissionRepository interface {
	GetPermissionById(id int64) (*models.Permission, error)
	GetPermissionByName(name string) (*models.Permission, error)
	GetAllPermissions() ([]*models.Permission, error)
	CreatePermission(name string, description string, resource string, action string) (*models.Permission, error)
	UpdatePermissionById(id int64, name string, description string, resource string, action string) (*models.Permission, error)
	DeletePermissionById(id int64) error
}

type PermissionRepositoryImpl struct {
	db *sql.DB
}

func NewPermissionRepository(conn *sql.DB) PermissionRepository {
	return &PermissionRepositoryImpl{
		db: conn,
	}
}

func (r *PermissionRepositoryImpl) GetPermissionById(id int64) (*models.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at, updated_at FROM permissions WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Resource, &permission.Action, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) GetPermissionByName(name string) (*models.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at, updated_at FROM permissions WHERE name = $1`

	row := r.db.QueryRow(query, name)

	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Resource, &permission.Action, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPermissionNotFound
		}
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) GetAllPermissions() ([]*models.Permission, error) {
	query := `SELECT id, name, description, resource, action, created_at, updated_at FROM permissions`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Resource, &permission.Action, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *PermissionRepositoryImpl) CreatePermission(name string, description string, resource string, action string) (*models.Permission, error) {
	query := `
		INSERT INTO permissions (name, description, resource, action)
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, description, resource, action, created_at, updated_at`

	row := r.db.QueryRow(query, name, description, resource, action)

	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Resource, &permission.Action, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			if strings.Contains(pqErr.Constraint, "name") {
				return nil, ErrPermissionNameTaken
			}
			return nil, errors.New("permission with this resource and action already exists")
		}
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) UpdatePermissionById(id int64, name string, description string, resource string, action string) (*models.Permission, error) {
	query := `
		UPDATE permissions
		SET name = $2, description = $3, resource = $4, action = $5, updated_at = now()
		WHERE id = $1
		RETURNING id, name, description, resource, action, created_at, updated_at`

	row := r.db.QueryRow(query, id, name, description, resource, action)

	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.Resource, &permission.Action, &permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPermissionNotFound
		}
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			if strings.Contains(pqErr.Constraint, "name") {
				return nil, ErrPermissionNameTaken
			}
			return nil, errors.New("permission with this resource and action already exists")
		}
		return nil, err
	}
	return &permission, nil
}

func (r *PermissionRepositoryImpl) DeletePermissionById(id int64) error {
	query := `DELETE FROM permissions WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return nil
}