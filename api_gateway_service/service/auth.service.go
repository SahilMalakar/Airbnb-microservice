package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/cache"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserService interface {
	SignUpService(ctx context.Context, data *dto.SignUpRequestDTO) (*models.User, string, string, error)
	LoginService(ctx context.Context, data *dto.LoginRequestDTO) (*models.User, string, string, error)
	RefreshTokenService(ctx context.Context, refreshToken string) (string, string, error)
	LogoutService(ctx context.Context, familyID string) error
	GetAllUsersService() ([]*models.User, error)
	GetUserByIDService(id int64) (*models.User, error)
}

// UserServiceImpl is the concrete, database-backed implementation of
// UserService.
type UserServiceImpl struct {
	userRepository     db.UserRepository
	userRoleRepository db.UserRoleRepository
	roleRepository     db.RoleRepository // add this
	refreshTokenStore  cache.RefreshTokenStore
}

func NewUserService(userRepo db.UserRepository, userRoleRepo db.UserRoleRepository, roleRepo db.RoleRepository, refreshTokenStore cache.RefreshTokenStore) UserService {
	return &UserServiceImpl{
		userRepository:     userRepo,
		userRoleRepository: userRoleRepo,
		roleRepository:     roleRepo,
		refreshTokenStore:  refreshTokenStore,
	}
}

func (u *UserServiceImpl) issueRefreshToken(ctx context.Context, userID int64, familyID string) (string, error) {
	jti := utils.NewTokenID()

	if err := u.refreshTokenStore.IssueFamily(ctx, familyID, jti, utils.RefreshTokenTTL); err != nil {
		return "", fmt.Errorf("failed to persist refresh token family: %w", err)
	}

	token, err := utils.SignRefreshToken(userID, familyID, jti)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *UserServiceImpl) rotateRefreshToken(ctx context.Context, userID int64, familyID, oldJTI string) (string, error) {
	newJTI := utils.NewTokenID()

	if err := u.refreshTokenStore.Rotate(ctx, familyID, oldJTI, newJTI, utils.RefreshTokenTTL); err != nil {
		return "", err
	}

	token, err := utils.SignRefreshToken(userID, familyID, newJTI)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token")
	}
	return token, nil
}

func (u *UserServiceImpl) SignUpService(ctx context.Context, data *dto.SignUpRequestDTO) (*models.User, string, string, error) {
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

	familyID := utils.NewRefreshFamilyID()
	accessToken, err := utils.CreateAccessToken(createdUser.ID, createdUser.Email, familyID, roles, permissions)

	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := u.issueRefreshToken(ctx, createdUser.ID, familyID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return createdUser, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) LoginService(ctx context.Context, data *dto.LoginRequestDTO) (*models.User, string, string, error) {
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

	familyID := utils.NewRefreshFamilyID()
	accessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, familyID, roles, permissions)

	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	refreshToken, err := u.issueRefreshToken(ctx, existingUser.ID, familyID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return existingUser, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) RefreshTokenService(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}
	familyID, ok := claims["familyId"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}
	jti, ok := claims["jti"].(string)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}
	userID := int64(idFloat)

	existingUser, err := u.userRepository.GetUserByID(userID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	newRefreshToken, err := u.rotateRefreshToken(ctx, existingUser.ID, familyID, jti)

	if err != nil {
		if errors.Is(err, cache.ErrReuseDetected) {
			fmt.Println("SECURITY: refresh token reuse detected, family revoked for user", userID)
			if denylistErr := u.refreshTokenStore.DenylistFamily(ctx, familyID, utils.AccessTokenTTL); denylistErr != nil {
				fmt.Println("failed to denylist access tokens after reuse detection:", denylistErr)
			}
			return "", "", fmt.Errorf("refresh token invalid, please login again")
		}
		if errors.Is(err, cache.ErrFamilyNotFound) {
			return "", "", fmt.Errorf("refresh token invalid, please login again")
		}
		return "", "", fmt.Errorf("failed to rotate refresh token")

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

	newAccessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, familyID, roles, permissions)

	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token")
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UserServiceImpl) LogoutService(ctx context.Context, familyID string) error {
	if familyID == "" {
		return nil
	}
	if err := u.refreshTokenStore.Revoke(ctx, familyID); err != nil {
		return err
	}
	return u.refreshTokenStore.DenylistFamily(ctx, familyID, utils.AccessTokenTTL)
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
