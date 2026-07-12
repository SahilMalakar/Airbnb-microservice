package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type PermissionController struct {
	PermissionService service.PermissionService
}

func NewPermissionController(permissionService service.PermissionService) *PermissionController {
	return &PermissionController{
		PermissionService: permissionService,
	}
}

func (c *PermissionController) CreatePermission(w http.ResponseWriter, r *http.Request, req dto.CreatePermissionRequestDTO) {
	permission, err := c.PermissionService.CreatePermissionService(&req)
	if err != nil {
		utils.SendError(w, http.StatusConflict, "Error permission", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusCreated, "permission created", permission)
}

func (c *PermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request, req dto.UpdatePermissionRequestDTO) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error permission id", "invalid permission id")
		return
	}

	permission, err := c.PermissionService.UpdatePermissionService(id, &req)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error permission updated", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permission updated", permission)
}

func (c *PermissionController) GetPermissionByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error permission id", "invalid permission id")
		return
	}

	permission, err := c.PermissionService.GetPermissionByIDService(id)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Error permission id", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permission fetched", permission)
}

func (c *PermissionController) GetAllPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := c.PermissionService.GetAllPermissionsService()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error permission", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permissions fetched", permissions)
}

func (c *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error permission id", "invalid permission id")
		return
	}

	if err := c.PermissionService.DeletePermissionService(id); err != nil {
		utils.SendError(w, http.StatusNotFound, "Error permission delete", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permission deleted", nil)
}