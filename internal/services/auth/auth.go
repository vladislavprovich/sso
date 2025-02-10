package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/vladislavprovich/sso/internal/domain/models"
	"github.com/vladislavprovich/sso/internal/lib/jwtlib"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log         *slog.Logger
	usrSaver    UserSaver
	usrProvider UserProvider
	appProvider AppProvider
	tokenTTL    time.Duration
	tracer      trace.Tracer
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
		tracer:      otel.Tracer("auth-service"),
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
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	// Get user from DB.
	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		span.SetStatus(codes.Error, "user not found")
		span.RecordError(err)
		if errors.Is(err, ErrUserNotFound) {
			a.log.WarnContext(ctx, "error get user", op, ErrUserNotFound)
			return "", fmt.Errorf("%s : %w", op, ErrUserNotFound)
		}

		a.log.ErrorContext(ctx, "error get user", op, err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	// Check password.
	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		span.RecordError(err)

		if errors.Is(err, ErrInvalidCredentials) {
			a.log.WarnContext(ctx, "error password", op, ErrInvalidCredentials)
			return "", fmt.Errorf("%s : %w", op, ErrInvalidCredentials)
		}

		a.log.ErrorContext(ctx, "error password", op, err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	// Get the app models. We take the secret key from the app.
	app, err := a.appProvider.App(ctx, int64(appID))
	if err != nil {
		span.SetStatus(codes.Error, "error getting app")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "error get app", op, err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	// Created token.
	token, err := jwtlib.GenerateToken(user, app, a.tokenTTL)
	if err != nil {
		span.SetStatus(codes.Error, "error generating token")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "error generate token", op, err)
		return "", fmt.Errorf("%s : %w", op, err)
	}

	a.log.InfoContext(ctx, "op:", slog.String("operation", op),
		slog.Int64("userID", user.ID),
		slog.String("token", token),
	)

	return token, nil
}

// RegisterNewUser registers new user in the system and returns user ID.
// If user with given username already exists, returns error.
func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	// GenerateFromPassword returns the bcrypt hash of the password at the given cost.
	// If the cost given is less than MinCost, the cost will be set to DefaultCost, instead.
	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		span.SetStatus(codes.Error, "error generating password hash")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "error generate password", op, err)
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	// Save user id DB.
	userID, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		span.SetStatus(codes.Error, "error saving user")
		span.RecordError(err)

		if errors.Is(err, ErrUserExists) {
			a.log.WarnContext(ctx, "error save user", op, ErrUserExists)
			return 0, fmt.Errorf("%s : %w", op, ErrUserExists)
		}

		a.log.ErrorContext(ctx, "error save user", op, err)
		return 0, fmt.Errorf("%s : %w", op, err)
	}

	span.SetAttributes(attribute.Int64("registered_user_id", userID))

	a.log.InfoContext(ctx, "op:", slog.String("operation", op),
		slog.Int64("registered_user_id", userID),
	)

	return userID, nil
}

// IsAdmin checks if user is admin.
func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	// Check admin, true or false.
	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		span.SetStatus(codes.Error, "error checking admin status")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "error check admin", op, err)
		return false, fmt.Errorf("%s : %w", op, err)
	}

	a.log.InfoContext(ctx, "op:", slog.String("operation", op),
		slog.Int64("userID", userID),
		slog.Bool("isAdmin", isAdmin),
	)

	return isAdmin, nil
}
