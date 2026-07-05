package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserRoleController struct {
	UserRoleService service.UserRoleService
}

func NewUserRoleController(userRoleService service.UserRoleService) *UserRoleController {
	return &UserRoleController{
		UserRoleService: userRoleService,
	}
}

func (c *UserRoleController) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	roles, err := c.UserRoleService.GetUserRolesService(userId)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": roles})
}

func (c *UserRoleController) AssignRole(w http.ResponseWriter, r *http.Request, req dto.AssignRoleRequestDTO) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	if err := c.UserRoleService.AssignRoleService(userId, &req); err != nil {
		utils.WriteJSONResponse(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusCreated, map[string]string{"message": "role assigned"})
}

func (c *UserRoleController) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	roleId, err := strconv.ParseInt(chi.URLParam(r, "roleId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid role id"})
		return
	}

	if err := c.UserRoleService.RemoveRoleService(userId, roleId); err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "role removed"})
}

func (c *UserRoleController) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}

	permissions, err := c.UserRoleService.GetUserPermissionsService(userId)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"data": permissions})
}