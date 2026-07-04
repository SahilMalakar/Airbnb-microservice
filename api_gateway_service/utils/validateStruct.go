package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct runs struct tag validation and returns a single
// human-readable error combining all failed fields, or nil if valid.
func ValidateStruct(s any) error {
	if err := validate.Struct(s); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return err
		}
		var msgs []string
		for _, fe := range validationErrors {
			msgs = append(msgs, formatFieldError(fe))
		}
		return fmt.Errorf("%s", strings.Join(msgs, "; "))
	}
	return nil
}

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
