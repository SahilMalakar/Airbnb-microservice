package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
)

type RoleRouter struct {
	RoleController *handler.RoleController
}

func NewRoleRouter(roleController *handler.RoleController) Router {
	return &RoleRouter{
		RoleController: roleController,
	}
}

func (router *RoleRouter) Register(r chi.Router) {
	r.Route("/role", func(r chi.Router) {
		r.Use(middleware.AuthCookie) // admin-only management; add role check middleware once RequirePermission exists
		r.Get("/", router.RoleController.GetAllRoles)
		r.Post("/", middleware.DecodeAndValidate(router.RoleController.CreateRole))
		r.Get("/{id}", router.RoleController.GetRoleByID)
		r.Patch("/{id}", middleware.DecodeAndValidate(router.RoleController.UpdateRole))
		r.Delete("/{id}", router.RoleController.DeleteRole)
	})
}