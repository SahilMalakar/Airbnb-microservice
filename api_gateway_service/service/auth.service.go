package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/cache"
	db "github.com/sahilmalakar/airbnb-microservice/api-gateway/db/repository"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/dto"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/models"
	"github.com/sahilmalakar/airbnb-microservice/api-gateway/utils"
)

type UserService interface {
	SignUpService(ctx context.Context, data *dto.SignUpRequestDTO) (*models.User, error)
	VerifySignupOTPService(ctx context.Context, email, otp string) (*models.User, string, string, error)
	ForgotPasswordService(ctx context.Context, email string) error
	ResetPasswordService(ctx context.Context, email, otp, newPassword string) error
	ResendOTPService(ctx context.Context, email, purpose string) error
	LoginService(ctx context.Context, data *dto.LoginRequestDTO) (*models.User, string, string, error)
	RefreshTokenService(ctx context.Context, refreshToken string) (string, string, error)
	LogoutService(ctx context.Context, familyID string) error
	GetAllUsersService() ([]*models.User, error)
	GetUserByIDService(id int64) (*models.User, error)
}

type UserServiceImpl struct {
	userRepository     db.UserRepository
	userRoleRepository db.UserRoleRepository
	roleRepository     db.RoleRepository
	refreshTokenStore  cache.RefreshTokenStore
	otpStore           cache.OTPStore
	notificationClient *NotificationClient
}

func NewUserService(
	userRepo db.UserRepository,
	userRoleRepo db.UserRoleRepository,
	roleRepo db.RoleRepository,
	refreshTokenStore cache.RefreshTokenStore,
	otpStore cache.OTPStore,
	notificationClient *NotificationClient,
) UserService {
	return &UserServiceImpl{
		userRepository:     userRepo,
		userRoleRepository: userRoleRepo,
		roleRepository:     roleRepo,
		refreshTokenStore:  refreshTokenStore,
		otpStore:           otpStore,
		notificationClient: notificationClient,
	}
}

func (u *UserServiceImpl) issueRefreshToken(ctx context.Context, userID int64, familyID string) (string, error) {
	jti := utils.NewTokenID()

	if err := u.refreshTokenStore.IssueFamily(ctx, familyID, jti, utils.RefreshTokenTTL); err != nil {
		return "", fmt.Errorf("failed to persist refresh token family: %w", err)
	}
	if err := u.refreshTokenStore.TrackUserFamily(ctx, userID, familyID, utils.RefreshTokenTTL); err != nil {
		return "", fmt.Errorf("failed to track refresh token family: %w", err)
	}

	token, err := utils.SignRefreshToken(userID, familyID, jti)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *UserServiceImpl) rotateRefreshToken(ctx context.Context, userID int64, familyID, oldJTI string) (string, error) {
	newJTI := utils.NewTokenID()

	if err := u.refreshTokenStore.Rotate(ctx, familyID, oldJTI, newJTI, utils.RefreshTokenTTL); err != nil {
		return "", err
	}

	token, err := utils.SignRefreshToken(userID, familyID, newJTI)
	if err != nil {
		return "", fmt.Errorf("failed to generate refresh token")
	}
	return token, nil
}

func (u *UserServiceImpl) verifyOTP(
	ctx context.Context,
	userID int64,
	purpose cache.OTPPurpose,
	otp string,
) error {
	locked, err := u.otpStore.IsLockedOut(ctx, userID, purpose)
	if err != nil {
		utils.Logger.Error("checking OTP lockout", "error", err, "userID", userID, "purpose", purpose)
		return ErrInternal
	}
	if locked {
		return ErrTooManyOTPAttempts
	}

	err = u.otpStore.Consume(ctx, userID, purpose, otp)
	if err == nil {
		_ = u.otpStore.ClearAttempts(ctx, userID, purpose)
		return nil
	}

	if errors.Is(err, cache.ErrOTPMismatch) {
		if attemptErr := u.otpStore.RecordFailedAttempt(ctx, userID, purpose); attemptErr != nil {
			if errors.Is(attemptErr, cache.ErrOTPLockedOut) {
				return ErrTooManyOTPAttempts
			}
			utils.Logger.Error("recording OTP failed attempt", "error", attemptErr, "userID", userID, "purpose", purpose)
		}
		return ErrInvalidVerificationCode
	}

	if errors.Is(err, cache.ErrOTPNotFound) {
		return ErrVerificationCodeExpired
	}

	utils.Logger.Error("consuming OTP", "error", err, "userID", userID, "purpose", purpose)
	return ErrInternal
}

func (u *UserServiceImpl) enqueueOTPEmail(user *models.User, otp, templateID, subject, idempotencyPrefix string) {
	correlationID := uuid.New().String()
	idempotencyKey := fmt.Sprintf("%s-%d-%s", idempotencyPrefix, user.ID, correlationID)
	emailReq := dto.EnqueueEmailRequest{
		NotificationType: "EMAIL",
		To:               user.Email,
		Subject:          subject,
		TemplateID:       templateID,
		Params: dto.OtpEmailPayload{
			Name:             user.Name,
			OTP:              otp,
			ExpiresInMinutes: utils.OTPExpiresInMinutes,
		},
		CorrelationID:  correlationID,
		IdempotencyKey: idempotencyKey,
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.notificationClient.EnqueueEmail(bgCtx, emailReq); err != nil {
			utils.Logger.Error("failed to enqueue OTP email", "error", err, "userID", user.ID, "template", templateID)
		}
	}()
}

func (u *UserServiceImpl) SignUpService(ctx context.Context, data *dto.SignUpRequestDTO) (*models.User, error) {
	_, err := u.userRepository.GetUserByEmailBasic(data.Email)
	if err == nil {
		return nil, ErrEmailAlreadyExists
	}
	if !errors.Is(err, db.ErrEmailNotFound) {
		utils.Logger.Error("checking existing user on signup", "error", err, "email", data.Email)
		return nil, ErrInternal
	}

	hashPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		utils.Logger.Error("hashing password on signup", "error", err)
		return nil, ErrInternal
	}

	createdUser, err := u.userRepository.Create(data.Name, data.Email, hashPassword)
	if err != nil {
		if errors.Is(err, db.ErrEmailAlreadyExists) {
			return nil, ErrEmailAlreadyExists
		}
		utils.Logger.Error("creating user on signup", "error", err, "email", data.Email)
		return nil, ErrInternal
	}

	defaultRole, err := u.roleRepository.GetRoleByName("user")
	if err != nil {
		utils.Logger.Error("loading default role on signup", "error", err)
		return nil, ErrDefaultRoleNotConfigured
	}

	if err := u.userRoleRepository.AssignRoleToUser(createdUser.ID, defaultRole.ID); err != nil {
		utils.Logger.Error("assigning default role on signup", "error", err, "userID", createdUser.ID)
		return nil, ErrFailedToAssignDefaultRole
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		utils.Logger.Error("generating signup OTP", "error", err, "userID", createdUser.ID)
		return createdUser, nil
	}

	if err := u.otpStore.Store(ctx, createdUser.ID, cache.OTPPurposeSignup, otp, utils.OTPTTL); err != nil {
		utils.Logger.Error("storing signup OTP", "error", err, "userID", createdUser.ID)
		return createdUser, nil
	}

	u.enqueueOTPEmail(createdUser, otp, "otp-signup", "Verify your Airbnb Account", "otp-signup")
	return createdUser, nil
}

func (u *UserServiceImpl) VerifySignupOTPService(ctx context.Context, email, otp string) (*models.User, string, string, error) {
	user, err := u.userRepository.GetUserByEmailBasic(email)
	if err != nil {
		if errors.Is(err, db.ErrEmailNotFound) {
			return nil, "", "", ErrInvalidVerificationCode
		}
		utils.Logger.Error("fetching user for signup verification", "error", err, "email", email)
		return nil, "", "", ErrInternal
	}

	if user.IsVerified {
		return nil, "", "", ErrInvalidVerificationCode
	}

	if err := u.verifyOTP(ctx, user.ID, cache.OTPPurposeSignup, otp); err != nil {
		return nil, "", "", err
	}

	if err := u.userRepository.MarkUserVerified(user.ID); err != nil {
		utils.Logger.Error("marking user verified", "error", err, "userID", user.ID)
		return nil, "", "", ErrInternal
	}

	correlationID := uuid.New().String()
	idempotencyKey := fmt.Sprintf("welcome-%d-%s", user.ID, correlationID)
	emailReq := dto.EnqueueEmailRequest{
		NotificationType: "EMAIL",
		To:               user.Email,
		Subject:          "Welcome to Airbnb!",
		TemplateID:       "welcome",
		Params: dto.WelcomeEmailPayload{
			Name: user.Name,
		},
		CorrelationID:  correlationID,
		IdempotencyKey: idempotencyKey,
	}

	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.notificationClient.EnqueueEmail(bgCtx, emailReq); err != nil {
			utils.Logger.Error("failed to enqueue welcome email", "error", err, "userID", user.ID)
		}
	}()

	roles, err := u.userRoleRepository.GetUserRoleNames(user.ID)
	if err != nil {
		utils.Logger.Error("loading user roles after verification", "error", err, "userID", user.ID)
		return nil, "", "", ErrInternal
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(user.ID)
	if err != nil {
		utils.Logger.Error("loading user permissions after verification", "error", err, "userID", user.ID)
		return nil, "", "", ErrInternal
	}

	familyID := utils.NewRefreshFamilyID()
	accessToken, err := utils.CreateAccessToken(user.ID, user.Email, user.Name, familyID, roles, permissions)
	if err != nil {
		utils.Logger.Error("generating access token after verification", "error", err, "userID", user.ID)
		return nil, "", "", ErrInternal
	}

	refreshToken, err := u.issueRefreshToken(ctx, user.ID, familyID)
	if err != nil {
		utils.Logger.Error("generating refresh token after verification", "error", err, "userID", user.ID)
		return nil, "", "", ErrInternal
	}

	user.IsVerified = true
	return user, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) ForgotPasswordService(ctx context.Context, email string) error {
	user, err := u.userRepository.GetUserByEmailBasic(email)
	if err != nil {
		if errors.Is(err, db.ErrEmailNotFound) {
			return nil
		}
		utils.Logger.Error("fetching user for forgot password", "error", err, "email", email)
		return ErrInternal
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		utils.Logger.Error("generating forgot-password OTP", "error", err, "userID", user.ID)
		return ErrInternal
	}

	if err := u.otpStore.Store(ctx, user.ID, cache.OTPPurposeForgotPassword, otp, utils.OTPTTL); err != nil {
		utils.Logger.Error("storing forgot-password OTP", "error", err, "userID", user.ID)
		return ErrInternal
	}

	u.enqueueOTPEmail(user, otp, "otp-forgot-password", "Reset your Airbnb Password", "otp-forgot")
	return nil
}

func (u *UserServiceImpl) ResetPasswordService(ctx context.Context, email, otp, newPassword string) error {
	user, err := u.userRepository.GetUserByEmailBasic(email)
	if err != nil {
		if errors.Is(err, db.ErrEmailNotFound) {
			return ErrInvalidVerificationCode
		}
		utils.Logger.Error("fetching user for password reset", "error", err, "email", email)
		return ErrInternal
	}

	if err := u.verifyOTP(ctx, user.ID, cache.OTPPurposeForgotPassword, otp); err != nil {
		return err
	}

	newHashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.Logger.Error("hashing new password", "error", err, "userID", user.ID)
		return ErrInternal
	}

	if err := u.userRepository.UpdatePassword(user.ID, newHashedPassword); err != nil {
		utils.Logger.Error("updating password", "error", err, "userID", user.ID)
		return ErrInternal
	}

	if err := u.refreshTokenStore.RevokeAllUserFamilies(ctx, user.ID, utils.AccessTokenTTL); err != nil {
		utils.Logger.Error("revoking refresh families after password reset", "error", err, "userID", user.ID)
	}

	return nil
}

func (u *UserServiceImpl) ResendOTPService(ctx context.Context, email, purpose string) error {
	user, err := u.userRepository.GetUserByEmailBasic(email)
	if err != nil {
		if errors.Is(err, db.ErrEmailNotFound) {
			return nil
		}
		utils.Logger.Error("fetching user for resend OTP", "error", err, "email", email)
		return ErrInternal
	}

	var otpPurpose cache.OTPPurpose
	var templateID string
	var subject string

	switch purpose {
	case "SIGNUP":
		if user.IsVerified {
			return nil
		}
		otpPurpose = cache.OTPPurposeSignup
		templateID = "otp-signup"
		subject = "Verify your Airbnb Account"
	case "FORGOT_PASSWORD":
		otpPurpose = cache.OTPPurposeForgotPassword
		templateID = "otp-forgot-password"
		subject = "Reset your Airbnb Password"
	default:
		return nil
	}

	otp, err := utils.GenerateOTP()
	if err != nil {
		utils.Logger.Error("generating resent OTP", "error", err, "userID", user.ID)
		return ErrInternal
	}

	if err := u.otpStore.Store(ctx, user.ID, otpPurpose, otp, utils.OTPTTL); err != nil {
		utils.Logger.Error("storing resent OTP", "error", err, "userID", user.ID)
		return ErrInternal
	}

	u.enqueueOTPEmail(user, otp, templateID, subject, "otp-resend")
	return nil
}

func (u *UserServiceImpl) LoginService(ctx context.Context, data *dto.LoginRequestDTO) (*models.User, string, string, error) {
	existingUser, err := u.userRepository.GetUserByEmail(data.Email)
	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if !utils.CheckPasswordHash(data.Password, existingUser.Password) {
		return nil, "", "", ErrInvalidCredentials
	}

	if !existingUser.IsVerified {
		return nil, "", "", ErrEmailNotVerified
	}

	roles, err := u.userRoleRepository.GetUserRoleNames(existingUser.ID)
	if err != nil {
		utils.Logger.Error("loading user roles on login", "error", err, "userID", existingUser.ID)
		return nil, "", "", ErrInternal
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(existingUser.ID)
	if err != nil {
		utils.Logger.Error("loading user permissions on login", "error", err, "userID", existingUser.ID)
		return nil, "", "", ErrInternal
	}

	familyID := utils.NewRefreshFamilyID()
	accessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, existingUser.Name, familyID, roles, permissions)
	if err != nil {
		utils.Logger.Error("generating access token on login", "error", err, "userID", existingUser.ID)
		return nil, "", "", ErrInternal
	}

	refreshToken, err := u.issueRefreshToken(ctx, existingUser.ID, familyID)
	if err != nil {
		utils.Logger.Error("generating refresh token on login", "error", err, "userID", existingUser.ID)
		return nil, "", "", ErrInternal
	}

	return existingUser, accessToken, refreshToken, nil
}

func (u *UserServiceImpl) RefreshTokenService(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.VerifyRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}

	idFloat, ok := claims["id"].(float64)
	if !ok {
		return "", "", ErrInvalidRefreshToken
	}
	familyID, ok := claims["familyId"].(string)
	if !ok {
		return "", "", ErrInvalidRefreshToken
	}
	jti, ok := claims["jti"].(string)
	if !ok {
		return "", "", ErrInvalidRefreshToken
	}
	userID := int64(idFloat)

	existingUser, err := u.userRepository.GetUserByIDBasic(userID)
	if err != nil {
		utils.Logger.Error("fetching user on refresh", "error", err, "userID", userID)
		return "", "", ErrInvalidRefreshToken
	}

	newRefreshToken, err := u.rotateRefreshToken(ctx, existingUser.ID, familyID, jti)
	if err != nil {
		if errors.Is(err, cache.ErrReuseDetected) {
			utils.Logger.Warn("refresh token reuse detected, family revoked", "userID", userID, "familyID", familyID)
			if denylistErr := u.refreshTokenStore.DenylistFamily(ctx, familyID, utils.AccessTokenTTL); denylistErr != nil {
				utils.Logger.Error("failed to denylist access tokens after reuse detection", "error", denylistErr, "familyID", familyID)
			}
			return "", "", ErrInvalidRefreshToken
		}
		if errors.Is(err, cache.ErrFamilyNotFound) {
			return "", "", ErrInvalidRefreshToken
		}
		utils.Logger.Error("rotating refresh token", "error", err, "userID", userID)
		return "", "", ErrInvalidRefreshToken
	}

	roles, err := u.userRoleRepository.GetUserRoleNames(existingUser.ID)
	if err != nil {
		utils.Logger.Error("loading user roles on refresh", "error", err, "userID", existingUser.ID)
		return "", "", ErrInternal
	}
	permissions, err := u.userRoleRepository.GetUserPermissionNames(existingUser.ID)
	if err != nil {
		utils.Logger.Error("loading user permissions on refresh", "error", err, "userID", existingUser.ID)
		return "", "", ErrInternal
	}

	newAccessToken, err := utils.CreateAccessToken(existingUser.ID, existingUser.Email, existingUser.Name, familyID, roles, permissions)
	if err != nil {
		utils.Logger.Error("generating access token on refresh", "error", err, "userID", existingUser.ID)
		return "", "", ErrInternal
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UserServiceImpl) LogoutService(ctx context.Context, familyID string) error {
	if familyID == "" {
		return nil
	}
	if err := u.refreshTokenStore.Revoke(ctx, familyID); err != nil {
		return err
	}
	return u.refreshTokenStore.DenylistFamily(ctx, familyID, utils.AccessTokenTTL)
}

func (u *UserServiceImpl) GetAllUsersService() ([]*models.User, error) {
	users, err := u.userRepository.GetAllUsers()
	if err != nil {
		utils.Logger.Error("fetching all users", "error", err)
		return nil, ErrInternal
	}
	return users, nil
}

func (u *UserServiceImpl) GetUserByIDService(id int64) (*models.User, error) {
	user, err := u.userRepository.GetUserByIDBasic(id)
	if err != nil {
		utils.Logger.Error("fetching user by id", "error", err, "userID", id)
		return nil, ErrInternal
	}
	return user, nil
}
