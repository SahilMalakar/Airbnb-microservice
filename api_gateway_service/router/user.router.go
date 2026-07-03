package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
)

type UserRouter struct {
	UserController *handler.UserController
}

func NewUserRouter(userController *handler.UserController) Router {
	return &UserRouter{
		UserController: userController,
	}
}

func (router *UserRouter) Register(r chi.Router) {
	r.Post("/signup", router.UserController.SignUp)
}
