package jwtlib

import (
	"github.com/golang-jwt/jwt/v5"
	"log/slog"
	"time"
)

// Claims struct for JWT-claims.
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	AppID  int64  `json:"app_id"`
	jwt.RegisteredClaims
}

// GenerateToken created new JWT token.
func GenerateToken(userID int64, email string, appID int64, ttl time.Duration) (string, error) {
	const op = "service.jwtlib.GenerateToken"
	var log slog.Logger

	expirationTime := time.Now().Add(ttl)

	claims := Claims{
		UserID: userID,
		Email:  email,
		AppID:  appID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString([]byte("secretKey!!!!!!!!!!!!!!!!!!!!!!!!!!!!"))
	if err != nil {
		log.Error(op, "JWT Error:"+err.Error())
		return op, err
	}

	return tokenStr, nil
}
