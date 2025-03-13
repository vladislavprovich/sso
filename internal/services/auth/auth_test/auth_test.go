package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"log/slog"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/vladislavprovich/sso/internal/domain/models"
	"github.com/vladislavprovich/sso/internal/rabbitmq/publisher"
	authpkg "github.com/vladislavprovich/sso/internal/services/auth"
)

type MockUserSaver struct {
	mock.Mock
}

func (m *MockUserSaver) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	args := m.Called(ctx, email, passHash)
	id, ok := args.Get(0).(int64)
	if !ok {
		return 0, args.Error(1)
	}
	return id, args.Error(1)
}

type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) User(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	user, ok := args.Get(0).(models.User)
	if !ok {
		return models.User{}, fmt.Errorf("unexpected type for user: %T", args.Get(0))
	}
	return user, args.Error(1)
}

func (m *MockUserProvider) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type MockAppProvider struct {
	mock.Mock
}

func (m *MockAppProvider) App(ctx context.Context, appID int64) (models.App, error) {
	args := m.Called(ctx, appID)
	app, ok := args.Get(0).(models.App)
	if !ok {
		return models.App{}, fmt.Errorf("unexpected type for app: %T", args.Get(0))
	}
	return app, args.Error(1)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) PublishUser(msg *publisher.RegisteredUser) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockPublisher) PublisherClose() error {
	args := m.Called()
	return args.Error(0)
}

func setupTestAuth() (*authpkg.Auth, *MockUserSaver, *MockUserProvider, *MockAppProvider, *MockPublisher) {
	mockUserSaver := new(MockUserSaver)
	mockUserProvider := new(MockUserProvider)
	mockAppProvider := new(MockAppProvider)
	mockPublisher := new(MockPublisher)

	logger := slog.Default()
	noopTracerProvider := noop.NewTracerProvider()

	auth := authpkg.New(
		logger,
		mockUserSaver,
		mockUserProvider,
		mockAppProvider,
		time.Hour,
		noopTracerProvider,
		mockPublisher,
	)
	return auth, mockUserSaver, mockUserProvider, mockAppProvider, mockPublisher
}

func TestLogin(t *testing.T) {
	auth, _, mockUserProvider, mockAppProvider, _ := setupTestAuth()

	hashedPass, _ := bcrypt.GenerateFromPassword([]byte("Password123"), bcrypt.DefaultCost)

	testCases := []struct {
		name          string
		email         string
		password      string
		mockUserResp  models.User
		mockUserErr   error
		mockAppResp   models.App
		mockAppErr    error
		expectedErr   error
		expectedToken bool
	}{
		{
			"Valid Login",
			"user@example.com",
			"Password123",
			models.User{
				ID:       1,
				Email:    "user@example.com",
				PassHash: hashedPass,
			},
			nil,
			models.App{
				ID:     1,
				Secret: "secret",
			},
			nil,
			nil,
			true,
		},
		{
			"User Not Found",
			"notfound@example.com",
			"Password123",
			models.User{},
			authpkg.ErrUserNotFound,
			models.App{},
			nil,
			authpkg.ErrUserNotFound,
			false,
		},
		{
			"Database Error",
			"user@example.com",
			"Password123",
			models.User{},
			errors.New("db error"),
			models.App{},
			nil,
			errors.New("db error"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserProvider.ExpectedCalls = nil
			mockAppProvider.ExpectedCalls = nil

			mockUserProvider.On("User", mock.Anything, tc.email).Return(tc.mockUserResp, tc.mockUserErr)
			mockAppProvider.On("App", mock.Anything, tc.mockUserResp.ID).Return(tc.mockAppResp, tc.mockAppErr)

			token, err := auth.Login(context.Background(), tc.email, tc.password, 1)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestRegisterNewUser(t *testing.T) {
	auth, mockUserSaver, _, _, mockPublisher := setupTestAuth()

	testCases := []struct {
		name          string
		email         string
		password      string
		mockSaveResp  int64
		mockSaveErr   error
		mockPubErr    error
		expectedErr   error
		expectPublish bool
	}{
		{
			"Successful Registration",
			"newuser@example.com",
			"Password123",
			10,
			nil,
			nil,
			nil,
			true,
		},
		{
			"Publisher Error",
			"user@example.com",
			"Password123",
			11,
			nil,
			errors.New("publish error"),
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserSaver.ExpectedCalls = nil
			mockPublisher.ExpectedCalls = nil

			mockUserSaver.On("SaveUser", mock.Anything, tc.email, mock.Anything).Return(tc.mockSaveResp, tc.mockSaveErr)

			if tc.expectPublish {
				mockPublisher.On("PublishUser", mock.Anything).Return(tc.mockPubErr)
			}

			userID, err := auth.RegisterNewUser(context.Background(), tc.email, tc.password)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.mockSaveResp, userID)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}

			if tc.expectPublish {
				mockPublisher.AssertCalled(t, "PublishUser", mock.Anything)
			} else {
				mockPublisher.AssertNotCalled(t, "PublishUser", mock.Anything)
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	auth, _, mockUserProvider, _, _ := setupTestAuth()

	testCases := []struct {
		name         string
		userID       int64
		mockResp     bool
		mockErr      error
		expectedErr  error
		expectedBool bool
	}{
		{
			"Is Admin",
			1,
			true,
			nil,
			nil,
			true,
		},
		{
			"Not Admin",
			2,
			false,
			nil,
			nil,
			false,
		},
		{
			"Error Checking Admin",
			1,
			false,
			errors.New("db error"),
			errors.New("db error"),
			false,
		},
		{
			"Unexpected Response",
			3,
			false,
			errors.New("unexpected error"),
			errors.New("unexpected error"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserProvider.ExpectedCalls = nil
			mockUserProvider.On("IsAdmin", mock.Anything, tc.userID).Return(tc.mockResp, tc.mockErr)

			isAdmin, err := auth.IsAdmin(context.Background(), tc.userID)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedBool, isAdmin)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestPublisherClose(t *testing.T) {
	mockPublisher := new(MockPublisher)

	mockPublisher.On("PublisherClose").Return(nil)

	err := mockPublisher.PublisherClose()
	require.NoError(t, err)

	mockPublisher.AssertExpectations(t)
}
