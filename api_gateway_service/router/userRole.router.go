package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
)

type UserRoleRouter struct {
	UserRoleController *handler.UserRoleController
}

func NewUserRoleRouter(userRoleController *handler.UserRoleController) Router {
	return &UserRoleRouter{
		UserRoleController: userRoleController,
	}
}

func (router *UserRoleRouter) Register(r chi.Router) {
	r.Route("/users/{userId}/roles", func(r chi.Router) {
		r.Use(middleware.AuthCookie)
		r.Get("/", router.UserRoleController.GetUserRoles)
		r.Post("/", middleware.DecodeAndValidate(router.UserRoleController.AssignRole))
		r.Delete("/{roleId}", router.UserRoleController.RemoveRole)
	})
	r.With(middleware.AuthCookie).Get("/users/{userId}/permissions", router.UserRoleController.GetUserPermissions)
}