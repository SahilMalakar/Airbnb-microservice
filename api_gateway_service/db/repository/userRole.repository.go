package db

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

var ErrRoleAlreadyAssigned = errors.New("role already assigned to user")
var ErrUserRoleNotFound = errors.New("user does not have this role")

type UserRoleRepository interface {
	GetUserRoles(userId int64) ([]*models.Role, error)
	AssignRoleToUser(userId int64, roleId int64) error
	RemoveRoleFromUser(userId int64, roleId int64) error
	GetUserPermissions(userId int64) ([]*models.Permission, error)
	HasPermission(userId int64, permissionName string) (bool, error)
	HasRole(userId int64, roleName string) (bool, error)
}

type UserRoleRepositoryImpl struct {
	db *sql.DB
}

func NewUserRoleRepository(_db *sql.DB) UserRoleRepository {
	return &UserRoleRepositoryImpl{
		db: _db,
	}
}

func (r *UserRoleRepositoryImpl) GetUserRoles(userId int64) ([]*models.Role, error) {
	query := `
		SELECT ro.id, ro.name, ro.description, ro.created_at, ro.updated_at
		FROM roles ro
		INNER JOIN user_roles ur ON ur.role_id = ro.id
		WHERE ur.user_id = $1`

	rows, err := r.db.Query(query, userId)
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

func (r *UserRoleRepositoryImpl) AssignRoleToUser(userId int64, roleId int64) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`

	_, err := r.db.Exec(query, userId, roleId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrRoleAlreadyAssigned
		}
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return errors.New("invalid user_id or role_id")
		}
		return err
	}
	return nil
}

func (r *UserRoleRepositoryImpl) RemoveRoleFromUser(userId int64, roleId int64) error {
	query := `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	result, err := r.db.Exec(query, userId, roleId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserRoleNotFound
	}
	return nil
}

func (r *UserRoleRepositoryImpl) GetUserPermissions(userId int64) ([]*models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.description, p.resource, p.action, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		INNER JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1`

	rows, err := r.db.Query(query, userId)
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

func (r *UserRoleRepositoryImpl) HasPermission(userId int64, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			INNER JOIN role_permissions rp ON rp.role_id = ur.role_id
			INNER JOIN permissions p ON p.id = rp.permission_id
			WHERE ur.user_id = $1 AND p.name = $2
		)`

	var exists bool
	err := r.db.QueryRow(query, userId, permissionName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *UserRoleRepositoryImpl) HasRole(userId int64, roleName string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM user_roles ur
			INNER JOIN roles ro ON ro.id = ur.role_id
			WHERE ur.user_id = $1 AND ro.name = $2
		)`

	var exists bool
	err := r.db.QueryRow(query, userId, roleName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}