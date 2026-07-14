package handler

import (
	"fmt"
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

	createdUser, accessToken, refreshToken, err := u.UserService.SignUpService(r.Context(), &req)
	if err != nil {
		utils.SendError(
			w,
			http.StatusConflict,
			"Error signup",
			err.Error(),
		)
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	utils.SendSuccess(
		w,
		http.StatusCreated,
		"signup successful",
		createdUser,
	)
}

// Login is now a ValidatedHandler[dto.LoginRequestDTO].
func (u *UserController) Login(w http.ResponseWriter, r *http.Request, req dto.LoginRequestDTO) {

	existingUser, accessToken, refreshToken, err := u.UserService.LoginService(r.Context(), &req)
	if err != nil {
		utils.SendError(
			w,
			http.StatusUnauthorized,
			"Error login",
			err.Error(),
		)
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	utils.SendSuccess(
		w,
		http.StatusOK,
		"login successful",
		existingUser,
	)
}

// RefreshToken and Logout don't take a request body, so they stay as
// plain http.HandlerFunc — no DecodeAndValidate needed.
func (u *UserController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.SendError(
			w,
			http.StatusUnauthorized,
			"Error refresh token",
			"refresh token missing",
		)
		return
	}

	newAccessToken, newRefreshToken, err := u.UserService.RefreshTokenService(r.Context(), cookie.Value)
	if err != nil {
		utils.ClearAuthCookies(w)
		utils.SendError(w, http.StatusUnauthorized, "Error refresh token", err.Error())
		return
	}

	utils.SetAuthCookies(w, newAccessToken, newRefreshToken)

	utils.SendSuccess(
		w,
		http.StatusOK,
		"token refreshed",
		nil,
	)
}

func (u *UserController) Logout(w http.ResponseWriter, r *http.Request) {

	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.SendError(
			w,
			http.StatusBadRequest,
			"Error user id",
			"invalid user id",
		)
		return
	}
	user, err := u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.SendError(
			w,
			http.StatusNotFound,
			"Error user",
			"user not found",
		)
		return
	}
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		if err := u.UserService.LogoutService(r.Context(), cookie.Value); err != nil {
			fmt.Println("Error revoking refresh token family on logout:", err)
			// don't fail the logout over this — cookies still get cleared below
		}
	}
	utils.ClearAuthCookies(w)
	utils.SendSuccess(
		w,
		http.StatusOK,
		"logged out",
		map[string]string{"user-email": user.Email},
	)
}

func (u *UserController) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.SendError(
			w,
			http.StatusBadRequest,
			"Error user id",
			"invalid user id",
		)
		return
	}
	_, err = u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.SendError(
			w,
			http.StatusNotFound,
			"Error user",
			"user not found",
		)
		return
	}

	users, err := u.UserService.GetAllUsersService()
	if err != nil {
		utils.SendError(
			w,
			http.StatusInternalServerError,
			"Error fetching users",
			err.Error(),
		)
		return
	}

	utils.SendSuccess(
		w,
		http.StatusOK,
		"users fetched successfully",
		users,
	)
}
