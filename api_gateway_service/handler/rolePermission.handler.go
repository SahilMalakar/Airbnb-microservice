package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type RolePermissionController struct {
	RolePermissionService service.RolePermissionService
}

func NewRolePermissionController(rolePermissionService service.RolePermissionService) *RolePermissionController {
	return &RolePermissionController{
		RolePermissionService: rolePermissionService,
	}
}

func (c *RolePermissionController) GetRolePermissionsByRoleID(w http.ResponseWriter, r *http.Request) {
	roleId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	rps, err := c.RolePermissionService.GetRolePermissionsByRoleIDService(roleId)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": rps})
}

func (c *RolePermissionController) AddPermission(w http.ResponseWriter, r *http.Request, req dto.AddPermissionRequestDTO) {
	roleId, err := strconv.ParseInt(chi.URLParam(r, "roleId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	rp, err := c.RolePermissionService.AddPermissionToRoleService(roleId, &req)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]any{"message": "permission added to role", "data": rp})
}

func (c *RolePermissionController) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	permissionId, err := strconv.ParseInt(chi.URLParam(r, "permissionId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid permission id"})
		return
	}

	if err := c.RolePermissionService.RemovePermissionFromRoleService(roleId, permissionId); err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "permission removed from role"})
}

func (c *RolePermissionController) GetAllRolePermissions(w http.ResponseWriter, r *http.Request) {
	rps, err := c.RolePermissionService.GetAllRolePermissionsService()
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": rps})
}