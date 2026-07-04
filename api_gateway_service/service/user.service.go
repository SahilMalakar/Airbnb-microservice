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
func (u *UserServiceImpl) SignUpService(data *dto.SignUpRequestDTO) (*models.User, string, string, error) {

	// 1. check if user with email is already present
	_, err := u.userRepository.GetUserByEmail(data.Email)
	if err == nil {
		// no error means a row was found -> email is taken
		return nil, "", "", fmt.Errorf("user with email %s already exists", data.Email)
	}
	if !errors.Is(err, db.ErrEmailNotFound) {
		// a real DB error occurred (not just "no rows") -> do not proceed
		return nil, "", "", fmt.Errorf("checking existing user: %w", err)
	}

	// 2. hash the password
	hashPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		return nil, "", "", err
	}

	// 3. call the repository function to create the user
	createdUser, err := u.userRepository.Create(data.Name, data.Email, hashPassword)
	if err != nil {
		if errors.Is(err, db.ErrEmailAlreadyExists) {
			return nil, "", "", fmt.Errorf("user with email %s already exists", data.Email)
		}
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
func (u *UserServiceImpl) LoginService(data *dto.LoginRequestDTO) (*models.User, string, string, error) {

	// 1. get user by email
	existingUser, err := u.userRepository.GetUserByEmail(data.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	// 2. compare password hash
	passwordsMatch := utils.CheckPasswordHash(data.Password, existingUser.Password)
	if !passwordsMatch {
		return nil, "", "", fmt.Errorf("invalid email or password")
	}

	// 3. generate access token
	accessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, string(existingUser.Role))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate access token")
	}

	// 4. generate refresh token
	refreshToken, err := utils.CreateRefreshToken(existingUser.ID)
	if err != nil {
		fmt.Printf("[DEBUG] LoginService: CreateRefreshToken error: %v\n", err)
		return nil, "", "", fmt.Errorf("failed to generate refresh token")
	}

	fmt.Println("[DEBUG] LoginService: Tokens generated successfully")
	return existingUser, accessToken, refreshToken, nil
}

// RefreshTokenService validates a refresh token and issues a new access +
// refresh token pair.
func (u *UserServiceImpl) RefreshTokenService(refreshToken string) (string, string, error) {

	// 1. validate the refresh token
	claims, err := utils.VerifyRefreshToken(refreshToken)
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