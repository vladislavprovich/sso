package authgrpc

import (
	"context"
	"errors"
	"regexp"
	"unicode"
)

// This regex validates email addresses.
// ^               - Start of the string
// [a-zA-Z0-9._%+-]+ - Local part (letters, numbers, ._%+-)
// @               - Requires exactly one "@"
// [a-zA-Z0-9.-]+  - Domain name (letters, numbers, .-)
// \.              - Matches a dot before the TLD
// [a-zA-Z]{2,}    - Top-level domain (at least 2 letters)
// $               - End of the string
// Example: "user@example.com" ✅, "invalid@com" ❌
const emailValidationParams = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var (
	emailNoEmpty          = errors.New("email cannot be empty")
	emailInvalidFormat    = errors.New("invalid email format")
	passwordInvalidFormat = errors.New("invalid password format")
	passwordInvalidText   = errors.New("password can only contain English letters and digits")
	appIDInvalidFormat    = errors.New("cannot be less than 0")
	userIDInvalidFormat   = errors.New("cannot be less than 0")
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) validateLoginRequest(ctx context.Context, email string, password string, appID int32) error {
	var emailRegex = regexp.MustCompile(emailValidationParams)

	if email == "" {
		return emailNoEmpty
	}
	if !emailRegex.MatchString(email) {
		return emailInvalidFormat
	}

	if password == "" {
		return passwordInvalidFormat
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return passwordInvalidText
		}
	}

	if appID < 0 {
		return appIDInvalidFormat
	}

	return nil
}

func (v *validator) validateRegisterRequest(ctx context.Context, email string, password string) error {
	var emailRegex = regexp.MustCompile(emailValidationParams)

	if email == "" {
		return emailNoEmpty
	}
	if !emailRegex.MatchString(email) {
		return emailInvalidFormat
	}

	if password == "" {
		return passwordInvalidFormat
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return passwordInvalidText
		}
	}

	return nil
}

func (v *validator) validateIsAdminRequest(ctx context.Context, userID int64) error {
	if userID < 0 {
		return userIDInvalidFormat
	}

	return nil
}
