package service

import "errors"

// Client-safe sentinel errors returned by auth services.
// Handlers map these to generic API responses without exposing internals.
var (
	ErrEmailAlreadyExists        = errors.New("user with this email already exists")
	ErrInvalidCredentials        = errors.New("invalid email or password")
	ErrInvalidVerificationCode   = errors.New("invalid email or verification code")
	ErrVerificationCodeExpired   = errors.New("verification code has expired or is invalid")
	ErrTooManyOTPAttempts        = errors.New("too many failed verification attempts, please try again later")
	ErrEmailNotVerified          = errors.New("please verify your email before logging in")
	ErrInvalidRefreshToken       = errors.New("refresh token invalid, please login again")
	ErrInternal                  = errors.New("something went wrong, please try again")
	ErrDefaultRoleNotConfigured  = errors.New("default role not configured")
	ErrFailedToAssignDefaultRole = errors.New("failed to assign default role")
)
