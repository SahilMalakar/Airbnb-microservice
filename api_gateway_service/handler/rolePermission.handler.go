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
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	rps, err := c.RolePermissionService.GetRolePermissionsByRoleIDService(roleId)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error on role permission fetching", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role permission fetched", rps)
}

func (c *RolePermissionController) AddPermission(w http.ResponseWriter, r *http.Request, req dto.AddPermissionRequestDTO) {
	roleId, err := strconv.ParseInt(chi.URLParam(r, "roleId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	rp, err := c.RolePermissionService.AddPermissionToRoleService(roleId, &req)
	if err != nil {
		utils.SendError(w, http.StatusConflict, "Error on role permission adding", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusCreated, "permission added to role", rp)
}

func (c *RolePermissionController) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleId, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	permissionId, err := strconv.ParseInt(chi.URLParam(r, "permissionId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error permission id", "invalid permission id")
		return
	}

	if err := c.RolePermissionService.RemovePermissionFromRoleService(roleId, permissionId); err != nil {
		utils.SendError(w, http.StatusNotFound, "Error on role permission removing", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permission removed from role", nil)
}

func (c *RolePermissionController) GetAllRolePermissions(w http.ResponseWriter, r *http.Request) {
	rps, err := c.RolePermissionService.GetAllRolePermissionsService()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error on role permission fetching", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "all role permissions fetched", rps)
}