package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vladislavprovich/sso/internal/domain/models"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

// Mock для UserSaver
type MockUserSaver struct {
	mock.Mock
}

func (m *MockUserSaver) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {
	args := m.Called(ctx, email, passHash)
	return args.Get(0).(int64), args.Error(1)
}

// Mock для UserProvider
type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) User(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserProvider) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

// Mock для AppProvider
type MockAppProvider struct {
	mock.Mock
}

func (m *MockAppProvider) App(ctx context.Context, appID int64) (models.App, error) {
	args := m.Called(ctx, appID)
	return args.Get(0).(models.App), args.Error(1)
}

// Налаштування тестового сервера
func setupTestAuth() (*Auth, *MockUserSaver, *MockUserProvider, *MockAppProvider) {
	mockUserSaver := new(MockUserSaver)
	mockUserProvider := new(MockUserProvider)
	mockAppProvider := new(MockAppProvider)
	logger := slog.Default()

	auth := New(logger, mockUserSaver, mockUserProvider, mockAppProvider, 1*time.Hour)
	return auth, mockUserSaver, mockUserProvider, mockAppProvider
}

// 🔹 Тест для `Login`
func TestLogin(t *testing.T) {
	auth, _, mockUserProvider, mockAppProvider := setupTestAuth()

	// Хешуємо пароль для тесту
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
			ErrUserNotFound,
			models.App{},
			nil,
			ErrUserNotFound,
			false,
		},
		{
			"Invalid Password",
			"user@example.com",
			"WrongPass",
			models.User{
				ID:       1,
				Email:    "user@example.com",
				PassHash: hashedPass,
			},
			nil,
			models.App{},
			nil,
			errors.New("not the hash"),
			false,
		},
		{
			"App Not Found",
			"user@example.com",
			"Password123",
			models.User{
				ID:       1,
				Email:    "user@example.com",
				PassHash: hashedPass,
			},
			nil,
			models.App{},
			errors.New("app not found"),
			errors.New("app not found"),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserProvider.ExpectedCalls = nil
			mockAppProvider.ExpectedCalls = nil

			mockUserProvider.On("User", mock.Anything, tc.email).Return(tc.mockUserResp, tc.mockUserErr)
			mockAppProvider.On("App", mock.Anything, int64(tc.mockUserResp.ID)).Return(tc.mockAppResp, tc.mockAppErr)

			token, err := auth.Login(context.Background(), tc.email, tc.password, 1)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

// 🔹 Тест для `RegisterNewUser`
func TestRegisterNewUser(t *testing.T) {
	auth, mockUserSaver, _, _ := setupTestAuth()

	testCases := []struct {
		name         string
		email        string
		password     string
		mockSaveResp int64
		mockSaveErr  error
		expectedErr  error
	}{
		{
			"Valid Registration",
			"new@example.com",
			"Password123",
			1,
			nil,
			nil,
		},
		{
			"User Exists",
			"existing@example.com",
			"Password123",
			0,
			ErrUserExists,
			ErrUserExists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserSaver.ExpectedCalls = nil
			mockUserSaver.On("SaveUser", mock.Anything, tc.email, mock.Anything).Return(tc.mockSaveResp, tc.mockSaveErr)

			userID, err := auth.RegisterNewUser(context.Background(), tc.email, tc.password)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockSaveResp, userID)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

// 🔹 Тест для `IsAdmin`
func TestIsAdmin(t *testing.T) {
	auth, _, mockUserProvider, _ := setupTestAuth()

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockUserProvider.ExpectedCalls = nil
			mockUserProvider.On("IsAdmin", mock.Anything, tc.userID).Return(tc.mockResp, tc.mockErr)

			isAdmin, err := auth.IsAdmin(context.Background(), tc.userID)

			if tc.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBool, isAdmin)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
