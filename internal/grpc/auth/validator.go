package authgrpc

import (
	"context"
	"errors"
	"regexp"
	"unicode"
)

const (
	emailValidationParams = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	emailNoEmpty          = "email cannot be empty"
	emailInvalidFormat    = "invalid email format"
	passwordInvalidFormat = "invalid password format"
	passwordInvalidText   = "password can only contain English letters and digits"
	appIDInvalidFormat    = "cannot be less than 0"
	userIDInvalidFormat   = "cannot be less than 0"
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) validateLoginRequest(ctx context.Context, email string, password string, appID int32) error {

	var emailRegex = regexp.MustCompile(emailValidationParams)

	if email == "" {
		return errors.New(emailNoEmpty)
	}
	if !emailRegex.MatchString(email) {
		return errors.New(emailInvalidFormat)
	}

	if password == "" {
		return errors.New(passwordInvalidFormat)
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return errors.New(passwordInvalidText)
		}
	}

	if appID < 0 {
		return errors.New(appIDInvalidFormat)
	}

	return nil
}

func (v *validator) validateRegisterRequest(ctx context.Context, email string, password string) error {

	var emailRegex = regexp.MustCompile(emailValidationParams)

	if email == "" {
		return errors.New(emailNoEmpty)
	}
	if !emailRegex.MatchString(email) {
		return errors.New(emailInvalidFormat)
	}

	if password == "" {
		return errors.New(passwordInvalidFormat)
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return errors.New(passwordInvalidText)
		}
	}

	return nil
}

func (v *validator) validateIsAdminRequest(ctx context.Context, userID int64) error {
	if userID < 0 {
		return errors.New(userIDInvalidFormat)
	}

	return nil
}
