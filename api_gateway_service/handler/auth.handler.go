package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/config"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/middleware"
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

func mapAuthError(err error) (status int, message, errorCode string) {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists):
		return http.StatusConflict, "Error signup", "EMAIL_ALREADY_EXISTS"
	case errors.Is(err, service.ErrDefaultRoleNotConfigured),
		errors.Is(err, service.ErrFailedToAssignDefaultRole):
		return http.StatusInternalServerError, "Signup could not be completed", "SIGNUP_FAILED"
	case errors.Is(err, service.ErrInvalidVerificationCode):
		return http.StatusBadRequest, "Verification failed", "INVALID_VERIFICATION_CODE"
	case errors.Is(err, service.ErrVerificationCodeExpired):
		return http.StatusBadRequest, "Verification failed", "VERIFICATION_CODE_EXPIRED"
	case errors.Is(err, service.ErrTooManyOTPAttempts):
		return http.StatusTooManyRequests, "Verification failed", "TOO_MANY_OTP_ATTEMPTS"
	case errors.Is(err, service.ErrInvalidCredentials):
		return http.StatusUnauthorized, "Error login", "INVALID_CREDENTIALS"
	case errors.Is(err, service.ErrEmailNotVerified):
		return http.StatusUnauthorized, "Error login", "EMAIL_NOT_VERIFIED"
	case errors.Is(err, service.ErrInvalidRefreshToken):
		return http.StatusUnauthorized, "Error refresh token", "INVALID_REFRESH_TOKEN"
	case errors.Is(err, service.ErrInternal):
		return http.StatusInternalServerError, "Something went wrong", "INTERNAL_ERROR"
	default:
		utils.Logger.Error("unmapped auth service error", "error", err)
		return http.StatusInternalServerError, "Something went wrong", "INTERNAL_ERROR"
	}
}

// SignUp validates signup data and creates an unverified user record.
func (u *UserController) SignUp(w http.ResponseWriter, r *http.Request, req dto.SignUpRequestDTO) {
	createdUser, err := u.UserService.SignUpService(r.Context(), &req)
	if err != nil {
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
		return
	}

	utils.SendSuccess(
		w,
		http.StatusCreated,
		"signup successful, please verify your email",
		map[string]interface{}{
			"email": createdUser.Email,
		},
	)
}

// VerifySignupOTP validates verification code, marks user verified, and issues tokens.
func (u *UserController) VerifySignupOTP(w http.ResponseWriter, r *http.Request, req dto.VerifySignupOTPRequestDTO) {
	user, accessToken, refreshToken, err := u.UserService.VerifySignupOTPService(r.Context(), req.Email, req.OTP)
	if err != nil {
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
		return
	}

	utils.SetAuthCookies(w, accessToken, refreshToken)

	utils.SendSuccess(
		w,
		http.StatusOK,
		"email verified successfully",
		user,
	)
}

// ForgotPassword generates OTP and sends it if the user is registered.
func (u *UserController) ForgotPassword(w http.ResponseWriter, r *http.Request, req dto.ForgotPasswordRequestDTO) {
	_ = u.UserService.ForgotPasswordService(r.Context(), req.Email)

	utils.SendSuccess(
		w,
		http.StatusOK,
		"If the email is registered, a password reset code has been sent",
		nil,
	)
}

// ResetPassword resets password if OTP is valid.
func (u *UserController) ResetPassword(w http.ResponseWriter, r *http.Request, req dto.ResetPasswordRequestDTO) {
	err := u.UserService.ResetPasswordService(r.Context(), req.Email, req.OTP, req.NewPassword)
	if err != nil {
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
		return
	}

	utils.SendSuccess(
		w,
		http.StatusOK,
		"password reset successfully",
		nil,
	)
}

// ResendOTP triggers a new OTP send with anti-enumeration behavior.
func (u *UserController) ResendOTP(w http.ResponseWriter, r *http.Request, req dto.ResendOTPRequestDTO) {
	_ = u.UserService.ResendOTPService(r.Context(), req.Email, req.Purpose)

	utils.SendSuccess(
		w,
		http.StatusOK,
		"If the email is registered, a verification code has been sent",
		nil,
	)
}

// Login validates credentials. Gated on email verification.
func (u *UserController) Login(w http.ResponseWriter, r *http.Request, req dto.LoginRequestDTO) {
	existingUser, accessToken, refreshToken, err := u.UserService.LoginService(r.Context(), &req)
	if err != nil {
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
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

// RefreshToken rotates refresh token and updates access token.
func (u *UserController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		utils.SendError(
			w,
			http.StatusUnauthorized,
			"Error refresh token",
			"REFRESH_TOKEN_MISSING",
		)
		return
	}

	newAccessToken, newRefreshToken, err := u.UserService.RefreshTokenService(r.Context(), cookie.Value)
	if err != nil {
		utils.ClearAuthCookies(w)
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
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

// Logout clears user session.
func (u *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.SendError(
			w,
			http.StatusBadRequest,
			"Error user id",
			"INVALID_USER_ID",
		)
		return
	}
	user, err := u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.SendError(
			w,
			http.StatusNotFound,
			"Error user",
			"USER_NOT_FOUND",
		)
		return
	}
	familyID, _ := r.Context().Value(middleware.CtxUserFamilyID).(string)
	if err := u.UserService.LogoutService(r.Context(), familyID); err != nil {
		utils.Logger.Error("failed to revoke session on logout", "error", err, "familyID", familyID)
	}
	utils.ClearAuthCookies(w)
	utils.SendSuccess(
		w,
		http.StatusOK,
		"logged out",
		map[string]string{"user-email": user.Email},
	)
}

// GetAllUsersHandler lists all users.
func (u *UserController) GetAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		utils.SendError(
			w,
			http.StatusBadRequest,
			"Error user id",
			"INVALID_USER_ID",
		)
		return
	}
	_, err = u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.SendError(
			w,
			http.StatusNotFound,
			"Error user",
			"USER_NOT_FOUND",
		)
		return
	}

	users, err := u.UserService.GetAllUsersService()
	if err != nil {
		status, message, errorCode := mapAuthError(err)
		utils.SendError(w, status, message, errorCode)
		return
	}

	utils.SendSuccess(
		w,
		http.StatusOK,
		"users fetched successfully",
		users,
	)
}

// GetInternalUserByID is called by internal microservices via internal service key auth.
func (u *UserController) GetInternalUserByID(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-Internal-Service-Key")
	if key == "" || key != config.RequireEnvString("INTERNAL_SERVICE_KEY") {
		utils.SendError(w, http.StatusUnauthorized, "Unauthorized", "INVALID_SERVICE_KEY")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID", "INVALID_USER_ID")
		return
	}

	user, err := u.UserService.GetUserByIDService(id)
	if err != nil {
		utils.SendError(w, http.StatusNotFound, "User not found", "USER_NOT_FOUND")
		return
	}

	utils.SendSuccess(w, http.StatusOK, "user details retrieved", map[string]string{
		"name":  user.Name,
		"email": user.Email,
	})
}
