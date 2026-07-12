package utils

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("strongpassword", validateStrongPassword)
}

func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	uniqueChars := make(map[rune]bool)

	for _, ch := range password {
		uniqueChars[ch] = true
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	classCount := 0
	for _, ok := range []bool{hasUpper, hasLower, hasDigit, hasSpecial} {
		if ok {
			classCount++
		}
	}

	// require at least 3 of the 4 character classes, AND enough
	// character variety that it's not just one repeated pattern
	// (e.g. "aaaaaaaa1A!" would pass a naive class check but has
	// almost no real entropy)
	minUniqueChars := len(password) / 2
	if minUniqueChars < 4 {
		minUniqueChars = 4
	}

	return classCount >= 3 && len(uniqueChars) >= minUniqueChars
}

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
