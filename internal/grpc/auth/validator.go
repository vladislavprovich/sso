package authgrpc

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"
)

type validator struct {
}

func NewValidator() *validator {
	return &validator{}
}

func (v *validator) validateLoginRequest(ctx context.Context, email string, password string, appID int32) error {

	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if email == "" {
		return errors.New("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 8 characters long")
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return errors.New("password can only contain English letters and digits")
		}
	}

	if appID <= 0 {
		return errors.New("invalid appID: must be greater than zero")
	}

	return nil
}

func (v *validator) validateRegisterRequest(ctx context.Context, email string, password string) error {

	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if email == "" {
		return errors.New("email cannot be empty")
	}
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 8 characters long")
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return errors.New("password can only contain English letters and digits")
		}
	}

	return nil
}

func (v *validator) validateIsAdminRequest(ctx context.Context, userID int64) error {
	if userID <= 0 {
		return errors.New("user id cannot be empty")
	}

	return nil
}

func (v *validator) validateLogoutRequest(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}

	if strings.Count(token, ".") != 2 {
		return errors.New("invalid token format")
	}

	if len(token) < 20 {
		return errors.New("token is too short")
	}

	return nil
}
