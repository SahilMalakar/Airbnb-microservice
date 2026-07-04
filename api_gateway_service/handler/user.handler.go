package handler

import (
	"net/http"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/service"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserController struct {
	UserService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (u *UserController) SignUp(w http.ResponseWriter, r *http.Request) {
	var req dto.SignUpRequestDTO
	if err := utils.ReadJSONRequest(w, r, &req); err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	role := models.RoleUser
	if req.Role == string(models.RoleHost) {
		role = models.RoleHost
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     role,
	}

	createdUser, accessToken, refreshToken, err := u.UserService.SignUpService(user)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	utils.WriteJSONResponse(w, http.StatusCreated, dto.UserResponseDTO{
		Message: "signup successful",
		User:    createdUser,
	})
}

func (u *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO
	if err := utils.ReadJSONRequest(w, r, &req); err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if err := utils.ValidateStruct(req); err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
	}

	existingUser, accessToken, refreshToken, err := u.UserService.LoginService(user)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	utils.WriteJSONResponse(w, http.StatusOK, dto.UserResponseDTO{
		Message: "login successful",
		User:    existingUser,
	})
}

func (u *UserController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": "refresh token missing"})
		return
	}

	newAccessToken, newRefreshToken, err := u.UserService.RefreshTokenService(cookie.Value)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	utils.SetAuthCookies(w, newAccessToken, newRefreshToken)

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "token refreshed"})
}

func (u *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	utils.ClearAuthCookies(w)
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "logged out"})
}
