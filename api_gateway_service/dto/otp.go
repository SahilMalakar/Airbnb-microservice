package dto

type VerifySignupOTPRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp"   validate:"required,len=6,numeric"`
}

type ForgotPasswordRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequestDTO struct {
	Email       string `json:"email"       validate:"required,email"`
	OTP         string `json:"otp"         validate:"required,len=6,numeric"`
	NewPassword string `json:"newPassword" validate:"required,min=6,max=64,strongpassword"`
}

type ResendOTPRequestDTO struct {
	Email   string `json:"email"   validate:"required,email"`
	Purpose string `json:"purpose" validate:"required,oneof=SIGNUP FORGOT_PASSWORD"`
}
