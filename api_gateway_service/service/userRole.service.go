package service

import (
	"errors"
	"fmt"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type UserRoleService interface {
	GetUserRolesService(userId int64) ([]*models.Role, error)
	AssignRoleService(userId int64, data *dto.AssignRoleRequestDTO) error
	RemoveRoleService(userId int64, roleId int64) error
	GetUserPermissionsService(userId int64) ([]*models.Permission, error)
	HasPermissionService(userId int64, permissionName string) (bool, error)
	HasRoleService(userId int64, roleName string) (bool, error)
}

type UserRoleServiceImpl struct {
	userRoleRepository db.UserRoleRepository
}

func NewUserRoleService(userRoleRepo db.UserRoleRepository) UserRoleService {
	return &UserRoleServiceImpl{
		userRoleRepository: userRoleRepo,
	}
}

func (s *UserRoleServiceImpl) GetUserRolesService(userId int64) ([]*models.Role, error) {
	roles, err := s.userRoleRepository.GetUserRoles(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles")
	}
	return roles, nil
}

func (s *UserRoleServiceImpl) AssignRoleService(userId int64, data *dto.AssignRoleRequestDTO) error {
	err := s.userRoleRepository.AssignRoleToUser(userId, data.RoleID)
	if err != nil {
		if errors.Is(err, db.ErrRoleAlreadyAssigned) {
			return fmt.Errorf("role already assigned to this user")
		}
		return err
	}
	return nil
}

func (s *UserRoleServiceImpl) RemoveRoleService(userId int64, roleId int64) error {
	if err := s.userRoleRepository.RemoveRoleFromUser(userId, roleId); err != nil {
		if errors.Is(err, db.ErrUserRoleNotFound) {
			return fmt.Errorf("user does not have this role")
		}
		return err
	}
	return nil
}

func (s *UserRoleServiceImpl) GetUserPermissionsService(userId int64) ([]*models.Permission, error) {
	permissions, err := s.userRoleRepository.GetUserPermissions(userId)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions")
	}
	return permissions, nil
}

func (s *UserRoleServiceImpl) HasPermissionService(userId int64, permissionName string) (bool, error) {
	return s.userRoleRepository.HasPermission(userId, permissionName)
}

func (s *UserRoleServiceImpl) HasRoleService(userId int64, roleName string) (bool, error) {
	return s.userRoleRepository.HasRole(userId, roleName)
}