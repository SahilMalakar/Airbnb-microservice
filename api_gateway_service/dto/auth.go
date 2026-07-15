package dto

type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=64"`
}

type SignUpRequestDTO struct {
	Name     string `json:"name" validate:"required,min=2,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=64,strongpassword"`
}
