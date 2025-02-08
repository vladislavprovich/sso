package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vladislavprovich/sso/internal/domain/models"
	"github.com/vladislavprovich/sso/internal/lib/jwtlib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

var (
	loginCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_attempts_total",
		Help: "Total number of login attempts",
	})
	loginFailuresCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "auth_login_failures_total",
		Help: "Total number of failed login attempts",
	})
)

func init() {
	prometheus.MustRegister(loginCounter, loginFailuresCounter)
}

type UserSaver interface {
	SaveUser(ctx context.Context, email string, passHash []byte) (uid int64, err error)
}

type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

type AppProvider interface {
	App(ctx context.Context, appID int64) (models.App, error)
}

// todo Params.
func New(log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration) *Auth {
	return &Auth{
		log:         log,
		usrSaver:    userSaver,
		usrProvider: userProvider,
		appProvider: appProvider,
		tracer:      otel.Tracer("auth-service"),
		tokenTTL:    tokenTTL,
	}
}

func (a *Auth) Login(ctx context.Context, email, password string, appID int) (string, error) {
	const op = "auth.Login"
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	loginCounter.Inc()

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		span.SetStatus(codes.Error, "user not found")
		span.RecordError(err)

		a.log.WarnContext(ctx, "User not found",
			slog.String("operation", op),
			slog.String("email", email))
		loginFailuresCounter.Inc()
		return "", fmt.Errorf("%s: %w", op, ErrUserNotFound)
	}

	if err = bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		span.SetStatus(codes.Error, "invalid credentials")
		span.RecordError(err)

		a.log.WarnContext(ctx, "Invalid password",
			slog.String("operation", op),
			slog.String("email", email))
		loginFailuresCounter.Inc()
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, int64(appID))
	if err != nil {
		span.SetStatus(codes.Error, "error getting app")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "Error getting app",
			slog.String("operation", op),
			slog.Int("appID", appID),
			slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := jwtlib.GenerateToken(user, app, a.tokenTTL)
	if err != nil {
		span.SetStatus(codes.Error, "error generating token")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "Error generating token",
			slog.String("operation", op),
			slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	span.SetAttributes(attribute.Int64("userID", user.ID), attribute.String("email", email))
	a.log.InfoContext(ctx, "User logged in",
		slog.Int64("userID", user.ID),
		slog.String("email", email))

	return token, nil
}

func (a *Auth) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		span.SetStatus(codes.Error, "error generating password hash")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "Error generating password hash",
			slog.String("operation", op),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		span.SetStatus(codes.Error, "error saving user")
		span.RecordError(err)

		if errors.Is(err, ErrUserExists) {
			a.log.WarnContext(ctx, "User already exists",
				slog.String("operation", op),
				slog.String("email", email))
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		a.log.ErrorContext(ctx, "Error saving user",
			slog.String("operation", op),
			slog.String("error", err.Error()))
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	span.SetAttributes(attribute.Int64("registered_user_id", userID))
	a.log.InfoContext(ctx, "User registered",
		slog.Int64("userID", userID),
		slog.String("email", email))

	return userID, nil
}

func (a *Auth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "auth.IsAdmin"
	ctx, span := a.tracer.Start(ctx, op)
	defer span.End()

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		span.SetStatus(codes.Error, "error checking admin status")
		span.RecordError(err)

		a.log.ErrorContext(ctx, "Error checking admin status",
			slog.String("operation", op),
			slog.Int64("userID", userID),
			slog.String("error", err.Error()))
		return false, fmt.Errorf("%s: %w", op, err)
	}

	span.SetAttributes(attribute.Int64("userID", userID), attribute.Bool("isAdmin", isAdmin))
	a.log.InfoContext(
		ctx,
		"Checked admin status",
		slog.Int64("userID", userID),
		slog.Bool("isAdmin", isAdmin))

	return isAdmin, nil
}
