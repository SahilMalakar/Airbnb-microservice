package service

import (
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserService interface {
	SignUpService(user *models.User) (*models.User, error)
	LoginService(user *models.User) (*models.User, error)
}

// UserServiceImpl is the concrete, database-backed implementation of
// UserService.
type UserServiceImpl struct {
	userRepository db.UserRepository
}

// NewUserServiceImpl is a constructor that builds and returns a new
// UserService instance with the provided UserRepository wired in.
func NewUserService(userRepo db.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: userRepo,
	}
}

// CreateUserService hashes the password of the user and persists the user
// record via the injected UserRepository.
func (u *UserServiceImpl) SignUpService(user *models.User) (*models.User, error) {
	
	//1.check if user with id is present or not

	hashPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}

	return u.userRepository.Create(user.Name, user.Email, hashPassword, user.Role)
}

func (u *UserServiceImpl) LoginService(user *models.User) (*models.User, error) {

	// 1. make a repo call to get the user by email , if user with email present or not

	// 2.if user is present then compare the hash password 

	// 3.if password is match then generate jwt token

	// 4.return the user with token

	return nil, nil
}