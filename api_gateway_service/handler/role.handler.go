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
		utils.SendError(w, http.StatusConflict, "Error role create", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusCreated, "role created", role)
}

// UpdateRole is a ValidatedHandler[dto.UpdateRoleRequestDTO].
func (c *RoleController) UpdateRole(w http.ResponseWriter, r *http.Request, req dto.UpdateRoleRequestDTO) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	role, err := c.RoleService.UpdateRoleService(id, &req)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role updated", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role updated", role)
}

func (c *RoleController) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	role, err := c.RoleService.GetRoleByIDService(id)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "Error role not found", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role found", role)
}

func (c *RoleController) GetAllRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := c.RoleService.GetAllRolesService()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error role fetching", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role fetched", roles)
}

func (c *RoleController) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	if err := c.RoleService.DeleteRoleService(id); err != nil {
		utils.SendError(w, http.StatusNotFound, "Error role not found", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role deleted", nil)
}
