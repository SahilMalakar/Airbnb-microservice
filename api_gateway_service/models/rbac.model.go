package models

import "time"

type Role struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type Permission struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type RolePermission struct {
	ID           int64     `json:"id" db:"id"`
	RoleID       int64     `json:"role_id" db:"role_id"`
	PermissionID int64     `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type UserRoles struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	RoleID    int64     `json:"role_id" db:"role_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// models/role_permission_detail.go (or wherever you keep view/aggregate models)
type RolePermissionDetail struct {
	ID             int64     `json:"id"`
	RoleID         int64     `json:"role_id"`
	RoleName       string    `json:"role_name"`
	PermissionID   int64     `json:"permission_id"`
	PermissionName string    `json:"permission_name"`
	Resource       string    `json:"resource"`
	Action         string    `json:"action"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}