package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type RoleController struct {
	RoleService service.RoleService
}

func NewRoleController(roleService service.RoleService) *RoleController {
	return &RoleController{
		RoleService: roleService,
	}
}

// CreateRole is a ValidatedHandler[dto.CreateRoleRequestDTO].
func (c *RoleController) CreateRole(w http.ResponseWriter, r *http.Request, req dto.CreateRoleRequestDTO) {
	role, err := c.RoleService.CreateRoleService(&req)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]any{"message": "role created", "data": role})
}

// UpdateRole is a ValidatedHandler[dto.UpdateRoleRequestDTO].
func (c *RoleController) UpdateRole(w http.ResponseWriter, r *http.Request, req dto.UpdateRoleRequestDTO) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	role, err := c.RoleService.UpdateRoleService(id, &req)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"message": "role updated", "data": role})
}

func (c *RoleController) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	role, err := c.RoleService.GetRoleByIDService(id)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": role})
}

func (c *RoleController) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.RoleService.GetAllRolesService()
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"message": "roles fetched successfully", "data": roles})
}

func (c *RoleController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	if err := c.RoleService.DeleteRoleService(id); err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "role deleted"})
}
