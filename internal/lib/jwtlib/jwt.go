package jwtlib

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vladislavprovich/sso/internal/domain/models"
	"log/slog"
	"time"
)

// GenerateToken created new JWT token.
func GenerateToken(user models.User, app models.App, ttl time.Duration) (string, error) {
	const op = "lib.jwtlib.GenerateToken"
	var log slog.Logger

	expirationTime := time.Now().Add(ttl).Unix()

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"app_id":  app.ID,
		"exp":     expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		log.Error(op, "JWT Error: "+err.Error())
		return "", fmt.Errorf("%s : %s", op, err)
	}

	return tokenStr, nil
}
