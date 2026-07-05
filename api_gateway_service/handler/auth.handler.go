package handler

import (
	"net/http"
	"strconv"

	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
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

// SignUp is now a ValidatedHandler[dto.SignUpRequestDTO] — decode + validate
// already happened in middleware.DecodeAndValidate before this runs.
func (u *UserController) SignUp(w http.ResponseWriter, r *http.Request, req dto.SignUpRequestDTO) {

	createdUser, accessToken, refreshToken, err := u.UserService.SignUpService(&req)
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

// Login is now a ValidatedHandler[dto.LoginRequestDTO].
func (u *UserController) Login(w http.ResponseWriter, r *http.Request, req dto.LoginRequestDTO) {

	existingUser, accessToken, refreshToken, err := u.UserService.LoginService(&req)
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

// RefreshToken and Logout don't take a request body, so they stay as
// plain http.HandlerFunc — no DecodeAndValidate needed.
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

	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	user, err := u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	utils.ClearAuthCookies(w)
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "logged out", "user-email": user.Email})
}

func (u *UserController) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusBadRequest, map[string]string{"error": "invalid user id"})
		return
	}
	_, err = u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusNotFound, map[string]string{"error": "user not found"})
		return
	}

	users, err := u.UserService.GetAllUsersService()
	if err != nil {
		utils.WriteJSONResponse(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]any{"message": "users fetched successfully", "data": users})
}
