package router

import (
	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/handler"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
)

type RolePermissionRouter struct {
	RolePermissionController *handler.RolePermissionController
}

func NewRolePermissionRouter(rolePermissionController *handler.RolePermissionController) Router {
	return &RolePermissionRouter{
		RolePermissionController: rolePermissionController,
	}
}

func (router *RolePermissionRouter) Register(r chi.Router) {
	r.Route("/roles/{roleId}/permissions", func(r chi.Router) {
		r.Use(middleware.AuthCookie)
		r.Get("/", router.RolePermissionController.GetRolePermissionsByRoleID)
		r.Post("/", middleware.DecodeAndValidate(router.RolePermissionController.AddPermission))
		r.Delete("/{permissionId}", router.RolePermissionController.RemovePermission)
	})
	r.With(middleware.AuthCookie).Get("/role-permissions", router.RolePermissionController.GetAllRolePermissions)
}