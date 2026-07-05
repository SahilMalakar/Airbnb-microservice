package service

import (
	"errors"
	"fmt"
	"strings"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type PermissionService interface {
	CreatePermissionService(data *dto.CreatePermissionRequestDTO) (*models.Permission, error)
	UpdatePermissionService(id int64, data *dto.UpdatePermissionRequestDTO) (*models.Permission, error)
	GetPermissionByIDService(id int64) (*models.Permission, error)
	GetPermissionByNameService(name string) (*models.Permission, error)
	GetAllPermissionsService() ([]*models.Permission, error)
	DeletePermissionService(id int64) error
}

type PermissionServiceImpl struct {
	permissionRepository db.PermissionRepository
}

func NewPermissionService(permissionRepo db.PermissionRepository) PermissionService {
	return &PermissionServiceImpl{
		permissionRepository: permissionRepo,
	}
}

func (s *PermissionServiceImpl) CreatePermissionService(data *dto.CreatePermissionRequestDTO) (*models.Permission, error) {
	permission, err := s.permissionRepository.CreatePermission(
		strings.TrimSpace(data.Name),
		data.Description,
		strings.TrimSpace(data.Resource),
		strings.TrimSpace(data.Action),
	)
	if err != nil {
		if errors.Is(err, db.ErrPermissionNameTaken) {
			return nil, fmt.Errorf("permission with name %s already exists", data.Name)
		}
		return nil, err
	}
	return permission, nil
}

func (s *PermissionServiceImpl) UpdatePermissionService(id int64, data *dto.UpdatePermissionRequestDTO) (*models.Permission, error) {
	existing, err := s.permissionRepository.GetPermissionById(id)
	if err != nil {
		return nil, fmt.Errorf("permission not found")
	}

	name := existing.Name
	if data.Name != nil {
		trimmed := strings.TrimSpace(*data.Name)
		if trimmed == "" {
			return nil, fmt.Errorf("name cannot be empty")
		}
		name = trimmed
	}

	description := existing.Description
	if data.Description != nil {
		description = *data.Description
	}

	resource := existing.Resource
	if data.Resource != nil {
		resource = strings.TrimSpace(*data.Resource)
	}

	action := existing.Action
	if data.Action != nil {
		action = strings.TrimSpace(*data.Action)
	}

	updated, err := s.permissionRepository.UpdatePermissionById(id, name, description, resource, action)
	if err != nil {
		if errors.Is(err, db.ErrPermissionNameTaken) {
			return nil, fmt.Errorf("permission with name %s already exists", name)
		}
		return nil, err
	}
	return updated, nil
}

func (s *PermissionServiceImpl) GetPermissionByIDService(id int64) (*models.Permission, error) {
	permission, err := s.permissionRepository.GetPermissionById(id)
	if err != nil {
		return nil, fmt.Errorf("permission not found")
	}
	return permission, nil
}

func (s *PermissionServiceImpl) GetPermissionByNameService(name string) (*models.Permission, error) {
	permission, err := s.permissionRepository.GetPermissionByName(name)
	if err != nil {
		return nil, fmt.Errorf("permission not found")
	}
	return permission, nil
}

func (s *PermissionServiceImpl) GetAllPermissionsService() ([]*models.Permission, error) {
	permissions, err := s.permissionRepository.GetAllPermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to get all permissions")
	}
	return permissions, nil
}

func (s *PermissionServiceImpl) DeletePermissionService(id int64) error {
	if err := s.permissionRepository.DeletePermissionById(id); err != nil {
		return fmt.Errorf("permission not found")
	}
	return nil
}