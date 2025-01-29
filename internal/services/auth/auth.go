package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vladislavprovich/sso/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"os"
	"time"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	secretKey             = []byte(os.Getenv("SECRET_KEY_JWT"))
)

type UserSaver interface {
	SaveUser(
		ctx context.Context,
		email string,
		passHash []byte,
	) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

// Login checks if user with given credentials exists in the system and returns access token.
//
// If user exists, but password is incorrect, returns error.
// If user doesn't exist, returns error.
func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
	appID int,
) (string, error) {
	const op = "auth.Login"

	// Get user from DB.
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		a.log.Warn(op, "error", ErrUserNotFound)
		return "", ErrUserNotFound
	}

	// Check password.
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Warn(op, "error", ErrInvalidCredentials)
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"userID": user.ID,
		"email":  user.Email,
		"appID":  appID,
		"exp":    time.Now().Add(a.tokenTTL).Unix(),
	}
	// Created token.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// SignedString creates and returns a complete, signed JWT.
	tokenStr, err := token.SignedString(secretKey)
	if err != nil {
		a.log.Error(op, "error", err)
		return "", err
	}

	a.log.Info(op, "userID", user.ID, "token", tokenStr)

	return tokenStr, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	// Check user.
	_, err := a.usrProvider.User(ctx, email)
	if err == nil {
		a.log.Warn(op, "error", ErrUserExists)
		return 0, ErrUserExists
	}

	// GenerateFromPassword returns the bcrypt hash of the password at the given cost.
	// If the cost given is less than MinCost, the cost will be set to DefaultCost, instead.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error(op, "error", err)
		return 0, err
	}

	// Save user id DB.
	userID, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		a.log.Error(op, "error", err)
		return 0, err
	}

	a.log.Info(op, "registered user id", userID)
	return userID, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin:"

	// Check admin, true or false.
	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		a.log.Error(op, "error", err)
		return false, err
	}

	a.log.Info(op, "userID", userID, "isAdmin", isAdmin)
	return isAdmin, nil
}
