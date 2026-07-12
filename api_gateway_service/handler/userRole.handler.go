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
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	roles, err := c.UserRoleService.GetUserRolesService(userId)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error on role fetching", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "user roles fetched", roles)
}

func (c *UserRoleController) AssignRole(w http.ResponseWriter, r *http.Request, req dto.AssignRoleRequestDTO) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	if err := c.UserRoleService.AssignRoleService(userId, &req); err != nil {
		utils.SendError(w, http.StatusConflict, "Error on role adding", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusCreated, "role assigned", nil)
}

func (c *UserRoleController) RemoveRole(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error user id", "invalid user id")
		return
	}

	roleId, err := strconv.ParseInt(chi.URLParam(r, "roleId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error role id", "invalid role id")
		return
	}

	if err := c.UserRoleService.RemoveRoleService(userId, roleId); err != nil {
		utils.SendError(w, http.StatusNotFound, "Error role not found", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "role removed", nil)
}

func (c *UserRoleController) GetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Error user id", "invalid user id")
		return
	}

	permissions, err := c.UserRoleService.GetUserPermissionsService(userId)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Error on permission fetching", err.Error())
		return
	}
	utils.SendSuccess(w, http.StatusOK, "permissions fetched", permissions)
}