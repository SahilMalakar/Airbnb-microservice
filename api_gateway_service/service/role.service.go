package service

import (
	"errors"
	"fmt"
	"strings"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type RoleService interface {
	CreateRoleService(data *dto.CreateRoleRequestDTO) (*models.Role, error)
	UpdateRoleService(id int64, data *dto.UpdateRoleRequestDTO) (*models.Role, error)
	GetRoleByIDService(id int64) (*models.Role, error)
	GetRoleByNameService(name string) (*models.Role, error)
	GetAllRolesService() ([]*models.Role, error)
	DeleteRoleService(id int64) error
}

type RoleServiceImpl struct {
	roleRepository db.RoleRepository
}

func NewRoleService(roleRepo db.RoleRepository) RoleService {
	return &RoleServiceImpl{
		roleRepository: roleRepo,
	}
}

func (s *RoleServiceImpl) CreateRoleService(data *dto.CreateRoleRequestDTO) (*models.Role, error) {
	role, err := s.roleRepository.CreateRole(strings.TrimSpace(data.Name), data.Description)
	if err != nil {
		if errors.Is(err, db.ErrRoleNameTaken) {
			return nil, fmt.Errorf("role with name %s already exists", data.Name)
		}
		return nil, err
	}
	return role, nil
}

func (s *RoleServiceImpl) UpdateRoleService(id int64, data *dto.UpdateRoleRequestDTO) (*models.Role, error) {
	existing, err := s.roleRepository.GetRoleById(id)
	if err != nil {
		return nil, err
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

	updated, err := s.roleRepository.UpdateRoleById(id, name, description)
	if err != nil {
		if errors.Is(err, db.ErrRoleNameTaken) {
			return nil, fmt.Errorf("role with name %s already exists", name)
		}
		return nil, err
	}
	return updated, nil
}

func (s *RoleServiceImpl) GetRoleByIDService(id int64) (*models.Role, error) {
	role, err := s.roleRepository.GetRoleById(id)
	if err != nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

func (s *RoleServiceImpl) GetRoleByNameService(name string) (*models.Role, error) {
	role, err := s.roleRepository.GetRoleByName(name)
	if err != nil {
		return nil, fmt.Errorf("role not found")
	}
	return role, nil
}

func (s *RoleServiceImpl) GetAllRolesService() ([]*models.Role, error) {
	roles, err := s.roleRepository.GetAllRoles()
	if err != nil {
		return nil, fmt.Errorf("failed to get all roles")
	}
	return roles, nil
}

func (s *RoleServiceImpl) DeleteRoleService(id int64) error {
	if err := s.roleRepository.DeleteRoleById(id); err != nil {
		return fmt.Errorf("role not found")
	}
	return nil
}