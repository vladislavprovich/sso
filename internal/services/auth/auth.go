package auth

import (
	"context"
	"errors"
	"github.com/vladislavprovich/sso/internal/domain/models"
	"github.com/vladislavprovich/sso/internal/services/jwtlib"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
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
		if errors.Is(err, ErrUserNotFound) {
			a.log.Warn(op, "error get user", ErrUserNotFound)
			return op, ErrUserNotFound
		}
		a.log.Error(op, "error get user", err)
		return op, err
	}

	// Check password.
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			a.log.Warn(op, "error password", ErrInvalidCredentials)
			return op, ErrInvalidCredentials
		}

		a.log.Error(op, "error password", err)
		return op, err
	}

	// Get the secret key.
	secret, err := a.appProvider.App(ctx, user.ID)
	if err != nil {
		a.log.Error(op, "error get secretKey", err)
		return op, err
	}

	// Created token.
	token, err := jwtlib.GenerateToken(user.ID, user.Email, int64(appID), a.tokenTTL, secret.Secret)
	a.log.Info(op, "userID", user.ID, "token", token)

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (string, int64, error) {
	const op = "auth.RegisterNewUser"

	// GenerateFromPassword returns the bcrypt hash of the password at the given cost.
	// If the cost given is less than MinCost, the cost will be set to DefaultCost, instead.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		a.log.Error(op, "error generate password", err)
		return op, 0, err
	}

	// Save user id DB.
	userID, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			a.log.Warn(op, "error save user", ErrUserExists)
			return op, 0, ErrUserExists
		}
		a.log.Error(op, "error save user", err)
		return op, 0, err
	}

	a.log.Info(op, "registered user id", userID)
	return op, userID, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (string, bool, error) {
	const op = "auth.IsAdmin"

	// Check admin, true or false.
	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		a.log.Error(op, "error check admin", err)
		return op, false, err
	}

	a.log.Info(op, "userID", userID, "isAdmin", isAdmin)
	return op, isAdmin, nil
}
