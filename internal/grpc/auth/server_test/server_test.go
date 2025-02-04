package server_test

import (
	"context"
	"errors"
	"testing"

	authgrpc "github.com/vladislavprovich/sso/internal/grpc/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ssov1 "github.com/vladislavprovich/protobufContract/gen/go/sso"
)

type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	args := m.Called(ctx, email, password, appID)
	return args.String(0), args.Error(1)
}

func (m *MockAuth) RegisterNewUser(ctx context.Context, email string, password string) (int64, error) {
	args := m.Called(ctx, email, password)

	id, ok := args.Get(0).(int64)
	if !ok {
		return 0, args.Error(1)
	}

	return id, args.Error(1)
}

func (m *MockAuth) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func setupTestServer() (*authgrpc.ServerAPI, *MockAuth) {
	mockAuth := new(MockAuth)
	server := &authgrpc.ServerAPI{
		Auth:      mockAuth,
		Validator: authgrpc.NewValidator(),
	}
	return server, mockAuth
}

func TestLogin(t *testing.T) {
	server, mockAuth := setupTestServer()

	testCases := []struct {
		name         string
		email        string
		password     string
		appID        int32
		mockAuthResp string
		mockAuthErr  error
		expectedCode codes.Code
	}{
		{
			"Auth Error",
			"user@example.com",
			"Password123",
			1,
			"",
			status.Error(codes.Internal, "internal error"),
			codes.Internal,
		},
		{
			"Valid Login",
			"user@example.com",
			"Password123",
			6,
			"valid_token",
			nil,
			codes.OK,
		},
		{
			"Empty Email",
			"",
			"Password123",
			1,
			"",
			authgrpc.ErrEmailNoEmpty,
			codes.InvalidArgument,
		},
		{
			"Empty Password",
			"user@example.com",
			"",
			1,
			"",
			authgrpc.ErrPasswordInvalidFormat,
			codes.InvalidArgument,
		},
		{
			"Invalid AppID",
			"user@example.com",
			"Password123",
			-1,
			"",
			authgrpc.ErrAppIDInvalidFormat,
			codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuth.On("Login", mock.Anything, tc.email, tc.password, int(tc.appID)).
				Return(tc.mockAuthResp, tc.mockAuthErr)

			req := &ssov1.LoginRequest{Email: tc.email, Password: tc.password, AppId: tc.appID}
			_, err := server.Login(context.Background(), req)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestRegister(t *testing.T) {
	server, mockAuth := setupTestServer()

	testCases := []struct {
		name         string
		email        string
		password     string
		mockAuthResp int64
		mockAuthErr  error
		expectedCode codes.Code
	}{
		{
			"Valid Registration",
			"new@example.com",
			"Password123",
			1,
			nil,
			codes.OK,
		},
		{
			"Empty Email",
			"",
			"Password123",
			0,
			authgrpc.ErrEmailNoEmpty,
			codes.InvalidArgument,
		},
		{
			"Empty Password",
			"user@example.com",
			"",
			0,
			authgrpc.ErrPasswordInvalidFormat,
			codes.InvalidArgument,
		},
		{
			"Auth Error",
			"user@example.com",
			"Password123",
			0,
			errors.New("internal error"),
			codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuth.On(
				"RegisterNewUser",
				mock.Anything,
				tc.email,
				tc.password,
			).Return(tc.mockAuthResp, tc.mockAuthErr)

			req := &ssov1.RegisterRequest{Email: tc.email, Password: tc.password}
			_, err := server.Register(context.Background(), req)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	server, mockAuth := setupTestServer()

	testCases := []struct {
		name         string
		userID       int64
		mockAuthResp bool
		mockAuthErr  error
		expectedCode codes.Code
	}{
		{
			"Is Admin",
			19999999999,
			true,
			nil,
			codes.OK,
		},
		{
			"Not Admin",
			2999999999,
			false,
			nil,
			codes.OK,
		},
		{
			"Invalid UserID",
			-1999999,
			false,
			authgrpc.ErrUserIDInvalidFormat,
			codes.InvalidArgument,
		},
		{
			"Auth Error",
			99999999,
			false,
			status.Error(codes.Internal, "internal error"),
			codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuth.On("IsAdmin", mock.Anything, tc.userID).
				Return(tc.mockAuthResp, tc.mockAuthErr)

			req := &ssov1.IsAdminRequest{UserId: tc.userID}
			_, err := server.IsAdmin(context.Background(), req)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
			} else {
				st, _ := status.FromError(err)
				assert.Equal(t, tc.expectedCode, st.Code())
			}
		})
	}
}
