package dto

type CreateRoleRequestDTO struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description" validate:"required,max=255"`
}

type UpdateRoleRequestDTO struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

type CreatePermissionRequestDTO struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description" validate:"max=255"`
	Resource    string `json:"resource" validate:"required,max=100"`
	Action      string `json:"action" validate:"required,max=50"`
}

type UpdatePermissionRequestDTO struct {
	Name        *string `json:"name" validate:"omitempty,min=3,max=100"`
	Description *string `json:"description" validate:"omitempty,max=255"`
	Resource    *string `json:"resource" validate:"omitempty,max=100"`
	Action      *string `json:"action" validate:"omitempty,max=50"`
}

type AssignRoleRequestDTO struct {
	RoleID int64 `json:"role_id" validate:"required"`
}

type AddPermissionRequestDTO struct {
	PermissionID int64 `json:"permission_id" validate:"required"`
}