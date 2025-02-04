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
// Example: "user@example.com" ✅, "invalid@com" ❌.
const emailValidationParams = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

var (
	ErrEmailNoEmpty          = errors.New("email cannot be empty")
	errEmailInvalidFormat    = errors.New("invalid email format")
	ErrPasswordInvalidFormat = errors.New("invalid password format")
	errPasswordInvalidText   = errors.New("password can only contain English letters and digits")
	ErrAppIDInvalidFormat    = errors.New("cannot be less than 0")
	ErrUserIDInvalidFormat   = errors.New("cannot be less than 0")
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) validateLoginRequest(ctx context.Context, email string, password string, appID int32) error {
	err := validateEmailAndPassword(ctx, email, password)
	if err != nil {
		return err
	}

	if appID < 0 {
		return ErrAppIDInvalidFormat
	}

	return nil
}

func (v *Validator) validateRegisterRequest(ctx context.Context, email string, password string) error {
	err := validateEmailAndPassword(ctx, email, password)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateIsAdminRequest(_ context.Context, userID int64) error {
	if userID < 0 {
		return ErrUserIDInvalidFormat
	}

	return nil
}

func validateEmailAndPassword(_ context.Context, email string, password string) error {
	var emailRegex = regexp.MustCompile(emailValidationParams)

	if email == "" {
		return ErrEmailNoEmpty
	}
	if !emailRegex.MatchString(email) {
		return errEmailInvalidFormat
	}

	if password == "" {
		return ErrPasswordInvalidFormat
	}

	for _, char := range password {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) {
			return errPasswordInvalidText
		}
	}
	return nil
}
