package service

import (
	"fmt"

	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserService interface {
	SignUpService(user *models.User) (*models.User, string, string, error)
	LoginService(user *models.User) (*models.User, string, string, error)
	RefreshTokenService(refreshToken string) (string, string, error)
}

// UserServiceImpl is the concrete, database-backed implementation of
// UserService.
type UserServiceImpl struct {
	userRepository db.UserRepository
}

// NewUserService is a constructor that builds and returns a new
// UserService instance with the provided UserRepository wired in.
func NewUserService(userRepo db.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: userRepo,
	}
}

// SignUpService hashes the password of the user, persists the user record
// via the injected UserRepository, and issues an access + refresh token pair
// so the user is immediately logged in after signup.
func (u *UserServiceImpl) SignUpService(user *models.User) (*models.User, string, string, error) {

	// 1. check if user with email is already present
	existingUser, err := u.userRepository.GetUserByEmail(user.Email)
	if err == nil && existingUser != nil {
		return nil, "", "", fmt.Errorf("user with email %s already exists", user.Email)
	}

	// 2. if user is not present then hash the password
	hashPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, "", "", err
	}

	// 3. call the repository function to create the user
	createdUser, err := u.userRepository.Create(user.Name, user.Email, hashPassword, user.Role)
	if err != nil {
		return nil, "", "", err
	}

	// 4. generate access token
	accessToken, err := utils.CreateAccessToken(createdUser.ID, createdUser.Email, string(createdUser.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	// 5. generate refresh token
	refreshToken, err := utils.CreateRefreshToken(createdUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return createdUser, accessToken, refreshToken, nil
}

// LoginService verifies credentials and issues an access + refresh token pair.
func (u *UserServiceImpl) LoginService(user *models.User) (*models.User, string, string, error) {

	// 1. get user by email
	existingUser, err := u.userRepository.GetUserByEmail(user.Email)
	if err != nil {
		fmt.Println("DEBUG GetUserByEmail error:", err)
		return nil, "", "", fmt.Errorf("user with email %s not found", user.Email)
	}

	fmt.Println("DEBUG existingUser:", existingUser)
	fmt.Println("DEBUG user.Password:", user.Password)
	fmt.Println("DEBUG existingUser.Password:", existingUser.Password)

	// 2. compare password hash
	if !utils.CheckPasswordHash(user.Password, existingUser.Password) {
		return nil, "", "", fmt.Errorf("invalid password")
	}

	// 3. generate access token
	accessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, string(existingUser.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	// 4. generate refresh token
	refreshToken, err := utils.CreateRefreshToken(existingUser.ID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	return existingUser, accessToken, refreshToken, nil
}

// RefreshTokenService validates a refresh token and issues a new access +
// refresh token pair.
func (u *UserServiceImpl) RefreshTokenService(refreshToken string) (string, string, error) {

	// 1. validate the refresh token
	claims, err := utils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return "", "", fmt.Errorf("invalid token claims")
	}
	userID := int64(idFloat)

	// 2. fetch the user to embed current email/role in the new access token
	existingUser, err := u.userRepository.GetUserByID(userID)
	if err != nil {
		return "", "", fmt.Errorf("user not found")
	}

	// 3. issue new token pair
	newAccessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, string(existingUser.Role))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token")
	}

	newRefreshToken, err := utils.CreateRefreshToken(existingUser.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token")
	}

	return newAccessToken, newRefreshToken, nil
}