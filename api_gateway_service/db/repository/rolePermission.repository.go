package db

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

var ErrRolePermissionNotFound = errors.New("role permission not found")
var ErrPermissionAlreadyAssigned = errors.New("permission already assigned to role")

type RolePermissionRepository interface {
	GetRolePermissionById(id int64) (*models.RolePermission, error)
	GetRolePermissionByRoleId(roleId int64) ([]*models.RolePermission, error)
	AddPermissionToRole(roleId int64, permissionId int64) (*models.RolePermission, error)
	RemovePermissionFromRole(roleId int64, permissionId int64) error
	GetAllRolePermissions() ([]*models.RolePermission, error)
}

type RolePermissionRepositoryImpl struct {
	db *sql.DB
}

func NewRolePermissionRepository(_db *sql.DB) RolePermissionRepository {
	return &RolePermissionRepositoryImpl{
		db: _db,
	}
}

func (r *RolePermissionRepositoryImpl) GetRolePermissionById(id int64) (*models.RolePermission, error) {
	query := `SELECT id, role_id, permission_id, created_at, updated_at FROM role_permissions WHERE id = $1`

	row := r.db.QueryRow(query, id)

	var rp models.RolePermission
	err := row.Scan(&rp.ID, &rp.RoleID, &rp.PermissionID, &rp.CreatedAt, &rp.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRolePermissionNotFound
		}
		return nil, err
	}
	return &rp, nil
}

func (r *RolePermissionRepositoryImpl) GetRolePermissionByRoleId(roleId int64) ([]*models.RolePermission, error) {
	query := `SELECT id, role_id, permission_id, created_at, updated_at FROM role_permissions WHERE role_id = $1`

	rows, err := r.db.Query(query, roleId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rolePermissions []*models.RolePermission
	for rows.Next() {
		rp := &models.RolePermission{}
		if err := rows.Scan(&rp.ID, &rp.RoleID, &rp.PermissionID, &rp.CreatedAt, &rp.UpdatedAt); err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rolePermissions, nil
}

func (r *RolePermissionRepositoryImpl) AddPermissionToRole(roleId int64, permissionId int64) (*models.RolePermission, error) {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
		RETURNING id, role_id, permission_id, created_at, updated_at`

	row := r.db.QueryRow(query, roleId, permissionId)

	var rp models.RolePermission
	err := row.Scan(&rp.ID, &rp.RoleID, &rp.PermissionID, &rp.CreatedAt, &rp.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return nil, ErrPermissionAlreadyAssigned
			}
			if pqErr.Code == "23503" {
				return nil, errors.New("invalid role_id or permission_id")
			}
		}
		return nil, err
	}
	return &rp, nil
}

func (r *RolePermissionRepositoryImpl) RemovePermissionFromRole(roleId int64, permissionId int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`

	result, err := r.db.Exec(query, roleId, permissionId)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRolePermissionNotFound
	}
	return nil
}

func (r *RolePermissionRepositoryImpl) GetAllRolePermissions() ([]*models.RolePermission, error) {
	query := `SELECT id, role_id, permission_id, created_at, updated_at FROM role_permissions`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rolePermissions []*models.RolePermission
	for rows.Next() {
		rp := &models.RolePermission{}
		if err := rows.Scan(&rp.ID, &rp.RoleID, &rp.PermissionID, &rp.CreatedAt, &rp.UpdatedAt); err != nil {
			return nil, err
		}
		rolePermissions = append(rolePermissions, rp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return rolePermissions, nil
}