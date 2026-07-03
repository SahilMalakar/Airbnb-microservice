package service

import (
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
)

type UserService interface {
	CreateUser(user *models.User) (*models.User, error)
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

// CreateUser wraps the Create method of the injected UserRepository,
// providing a clean separation between the service layer and the
// data-access layer.
func (u *UserServiceImpl) CreateUser(user *models.User) (*models.User, error) {
	return u.userRepository.Create(user)
}
