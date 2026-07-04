package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
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
	r.Post("/signup", middleware.DecodeAndValidate(router.UserController.SignUp))
	r.Post("/login", middleware.DecodeAndValidate(router.UserController.Login))
	// no request body, no DecodeAndValidate wrapper needed
	r.Post("/refresh", router.UserController.RefreshToken)
	r.Post("/logout", router.UserController.Logout)
}
