package service

import (
	"errors"
	"fmt"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserService interface {
	SignUpService(data *dto.SignUpRequestDTO) (*models.User, string, string, error)
	LoginService(data *dto.LoginRequestDTO) (*models.User, string, string, error)
	RefreshTokenService(refreshToken string) (string, string, error)
	GetAllUsersService() ([]*models.User, error)
	GetUserByIDService(id int64) (*models.User, error)
}

// UserServiceImpl is the concrete, database-backed implementation of
// UserService.
type UserServiceImpl struct {
	userRepository     db.UserRepository
	userRoleRepository db.UserRoleRepository
	roleRepository     db.RoleRepository // add this
}

func NewUserService(userRepo db.UserRepository, userRoleRepo db.UserRoleRepository, roleRepo db.RoleRepository) UserService {
	return &UserServiceImpl{
		userRepository:     userRepo,
		userRoleRepository: userRoleRepo,
		roleRepository:     roleRepo,
	}
}

// SignUpService hashes the password of the user, persists the user record
// via the injected UserRepository, and issues an access + refresh token pair
// so the user is immediately logged in after signup.
func (u *UserServiceImpl) SignUpService(data *dto.SignUpRequestDTO) (*models.User, string, string, error) {
	_, err := u.userRepository.GetUserByEmail(data.Email)
	if err == nil {
		return nil, "", "", fmt.Errorf("user with email %s already exists", data.Email)
	}
	if !errors.Is(err, db.ErrEmailNotFound) {
		return nil, "", "", fmt.Errorf("checking existing user: %w", err)
	}

	hashPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		return nil, "", "", err
	}

	createdUser, err := u.userRepository.Create(data.Name, data.Email, hashPassword)
	if err != nil {
		if errors.Is(err, db.ErrEmailAlreadyExists) {
			return nil, "", "", fmt.Errorf("user with email %s already exists", data.Email)
		}
		return nil, "", "", err
	}

	// look up the default "user" role dynamically — never hardcode its ID
	defaultRole, err := u.roleRepository.GetRoleByName("user")
	if err != nil {
		return nil, "", "", fmt.Errorf("default role not configured")
	}

	// assign the default "user" role to every new signup — adjust roleId to your seeded default
	if err := u.userRoleRepository.AssignRoleToUser(createdUser.ID, defaultRole.ID); err != nil {
		return nil, "", "", fmt.Errorf("failed to assign default role")
	}

	// fetch fresh from DB — never from the request body
	roles, err := u.userRoleRepository.GetUserRoleNames(createdUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load user roles")
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(createdUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load user permissions")
	}

	accessToken, err := utils.CreateAccessToken(createdUser.ID, createdUser.Email, roles, permissions)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := utils.CreateRefreshToken(createdUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return createdUser, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) LoginService(data *dto.LoginRequestDTO) (*models.User, string, string, error) {
	existingUser, err := u.userRepository.GetUserByEmail(data.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	if !utils.CheckPasswordHash(data.Password, existingUser.Password) {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	roles, err := u.userRoleRepository.GetUserRoleNames(existingUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load user roles")
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(existingUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to load user permissions")
	}

	accessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, roles, permissions)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := utils.CreateRefreshToken(existingUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return existingUser, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) RefreshTokenService(refreshToken string) (string, string, error) {
	claims, err := utils.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}
	userID := int64(idFloat)

	existingUser, err := u.userRepository.GetUserByID(userID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	// re-fetch fresh roles/permissions — this is your staleness-correction point
	roles, err := u.userRoleRepository.GetUserRoleNames(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load user roles")
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to load user permissions")
	}

	newAccessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, roles, permissions)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := utils.CreateRefreshToken(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UserServiceImpl) GetAllUsersService() ([]*models.User, error) {
	users, err := u.userRepository.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all users")
	}
	return users, nil
}

func (u *UserServiceImpl) GetUserByIDService(id int64) (*models.User, error) {
	user, err := u.userRepository.GetUserByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user")
	}
	return user, nil
}
