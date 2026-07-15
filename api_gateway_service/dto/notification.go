package dto

type OtpEmailPayload struct {
	Name             string `json:"name"`
	OTP              string `json:"otp"`
	ExpiresInMinutes int    `json:"expiresInMinutes"`
}

type WelcomeEmailPayload struct {
	Name string `json:"name"`
}

type EnqueueEmailRequest struct {
	NotificationType string `json:"notificationType"` // always "EMAIL"
	To               string `json:"to"`
	Subject          string `json:"subject"`
	TemplateID       string `json:"templateId"` // "welcome" | "otp-signup" | "otp-forgot-password"
	Params           any    `json:"params"`     // OtpEmailPayload or WelcomeEmailPayload
	CorrelationID    string `json:"correlationId"`
	IdempotencyKey   string `json:"idempotencyKey"`
}
