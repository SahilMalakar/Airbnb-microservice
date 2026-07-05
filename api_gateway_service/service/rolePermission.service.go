package service

import (
	"errors"
	"fmt"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type RolePermissionService interface {
	GetRolePermissionByIDService(id int64) (*models.RolePermission, error)
	GetRolePermissionsByRoleIDService(roleId int64) ([]*models.RolePermission, error)
	AddPermissionToRoleService(roleId int64, data *dto.AddPermissionRequestDTO) (*models.RolePermission, error)
	RemovePermissionFromRoleService(roleId int64, permissionId int64) error
	GetAllRolePermissionsService() ([]*models.RolePermission, error)
}

type RolePermissionServiceImpl struct {
	rolePermissionRepository db.RolePermissionRepository
}

func NewRolePermissionService(rolePermissionRepo db.RolePermissionRepository) RolePermissionService {
	return &RolePermissionServiceImpl{
		rolePermissionRepository: rolePermissionRepo,
	}
}

func (s *RolePermissionServiceImpl) GetRolePermissionByIDService(id int64) (*models.RolePermission, error) {
	rp, err := s.rolePermissionRepository.GetRolePermissionById(id)
	if err != nil {
		return nil, fmt.Errorf("role permission not found")
	}
	return rp, nil
}

func (s *RolePermissionServiceImpl) GetRolePermissionsByRoleIDService(roleId int64) ([]*models.RolePermission, error) {
	rps, err := s.rolePermissionRepository.GetRolePermissionByRoleId(roleId)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions")
	}
	return rps, nil
}

func (s *RolePermissionServiceImpl) AddPermissionToRoleService(roleId int64, data *dto.AddPermissionRequestDTO) (*models.RolePermission, error) {
	rp, err := s.rolePermissionRepository.AddPermissionToRole(roleId, data.PermissionID)
	if err != nil {
		if errors.Is(err, db.ErrPermissionAlreadyAssigned) {
			return nil, fmt.Errorf("permission already assigned to this role")
		}
		return nil, err
	}
	return rp, nil
}

func (s *RolePermissionServiceImpl) RemovePermissionFromRoleService(roleId int64, permissionId int64) error {
	if err := s.rolePermissionRepository.RemovePermissionFromRole(roleId, permissionId); err != nil {
		if errors.Is(err, db.ErrRolePermissionNotFound) {
			return fmt.Errorf("role does not have this permission")
		}
		return err
	}
	return nil
}

func (s *RolePermissionServiceImpl) GetAllRolePermissionsService() ([]*models.RolePermission, error) {
	rps, err := s.rolePermissionRepository.GetAllRolePermissions()
	if err != nil {
		return nil, fmt.Errorf("failed to get all role permissions")
	}
	return rps, nil
}