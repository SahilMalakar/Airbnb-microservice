package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
)

type PermissionRouter struct {
	PermissionController *handler.PermissionController
}

func NewPermissionRouter(permissionController *handler.PermissionController) Router {
	return &PermissionRouter{
		PermissionController: permissionController,
	}
}

func (router *PermissionRouter) Register(r chi.Router) {
	r.Route("/permissions", func(r chi.Router) {
		r.Use(middleware.AuthCookie)
		r.Get("/", router.PermissionController.GetAllPermissions)
		r.Post("/", middleware.DecodeAndValidate(router.PermissionController.CreatePermission))
		r.Get("/{id}", router.PermissionController.GetPermissionByID)
		r.Patch("/{id}", middleware.DecodeAndValidate(router.PermissionController.UpdatePermission))
		r.Delete("/{id}", router.PermissionController.DeletePermission)
	})
}