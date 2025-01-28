package authgrpc

import (
	"regexp"
)

type validator struct {
}

func (v *validator) validateEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return emailRegex.MatchString(email)
}

func (v *validator) validatePassword(password string) bool {
	// Min len password.
	// todo changes this or del this
	minLength := 8
	if len(password) < minLength {
		return false
	}

	return true
}

func (v *validator) validateToken(token string) bool {
	if token == "" {
		return false
	}

	return true
}

func (v *validator) validateUserID(userID int64) bool {
	if userID == 0 {
		return false
	}

	return true
}
