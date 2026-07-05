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
		utils.WriteJSONResponse(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]any{"message": "permission created", "data": permission})
}

func (c *PermissionController) UpdatePermission(w http.ResponseWriter, r *http.Request, req dto.UpdatePermissionRequestDTO) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid permission id"})
		return
	}

	permission, err := c.PermissionService.UpdatePermissionService(id, &req)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"message": "permission updated", "data": permission})
}

func (c *PermissionController) GetPermissionByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid permission id"})
		return
	}

	permission, err := c.PermissionService.GetPermissionByIDService(id)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": permission})
}

func (c *PermissionController) GetAllPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := c.PermissionService.GetAllPermissionsService()
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"message": "permissions fetched successfully", "data": permissions})
}

func (c *PermissionController) DeletePermission(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid permission id"})
		return
	}

	if err := c.PermissionService.DeletePermissionService(id); err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "permission deleted"})
}