package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

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
	fmt.Println("=== SIGNUP HANDLER HIT ===")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println("inside user handler", user)

	createdUser, accessToken, refreshToken, err := u.UserService.SignUpService(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	fmt.Println("created user: ", createdUser)

	utils.SetAuthCookies(w, accessToken, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := struct {
		User *models.User `json:"user"`
	}{
		User: createdUser,
	}
	json.NewEncoder(w).Encode(response)
}

func (u *UserController) Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== LOGIN HANDLER HIT ===")

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	existingUser, accessToken, refreshToken, err := u.UserService.LoginService(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := struct {
		User *models.User `json:"user"`
	}{
		User: existingUser,
	}
	json.NewEncoder(w).Encode(response)
}

func (u *UserController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== REFRESH TOKEN HANDLER HIT ===")

	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "refresh token missing", http.StatusUnauthorized)
		return
	}

	newAccessToken, newRefreshToken, err := u.UserService.RefreshTokenService(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	utils.SetAuthCookies(w, newAccessToken, newRefreshToken)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "token refreshed"})
}

func (u *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("=== LOGOUT HANDLER HIT ===")
	utils.ClearAuthCookies(w)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}
