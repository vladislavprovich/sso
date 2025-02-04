package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/vladislavprovich/sso/internal/storage/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vladislavprovich/sso/internal/domain/models"
	_ "github.com/vladislavprovich/sso/internal/storage"
)

func TestSaveUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &postgres.Storage{DB: db}

	testCases := []struct {
		name        string
		email       string
		passHash    []byte
		mockResult  int64
		mockError   error
		expectedErr error
	}{
		{
			"Success",
			"test@example.com",
			[]byte("hash"),
			1,
			nil,
			nil,
		},
		{
			"DB Error",
			"error@example.com",
			[]byte("hash"),
			0,
			errors.New("db error"),
			errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectQuery(`INSERT INTO users`).
				WithArgs(tc.email, tc.passHash).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tc.mockResult)).
				WillReturnError(tc.mockError)

			ctx := context.Background()

			var id int64
			id, err = storage.SaveUser(ctx, tc.email, tc.passHash)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.mockResult, id)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &postgres.Storage{DB: db}

	testCases := []struct {
		name        string
		email       string
		mockUser    models.User
		mockError   error
		expectedErr error
	}{
		{
			"User Found",
			"test@example.com",
			models.User{
				ID:       1,
				Email:    "test@example.com",
				PassHash: []byte("hash"),
			},
			nil,
			nil,
		},
		{
			"User Not Found",
			"notfound@example.com",
			models.User{},
			sql.ErrNoRows,
			errors.New("user not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"id", "email", "pass_hash"})
			if tc.mockError == nil {
				rows.AddRow(tc.mockUser.ID, tc.mockUser.Email, tc.mockUser.PassHash)
			}
			mock.ExpectQuery(`SELECT id, email, pass_hash FROM users WHERE email = \$1`).
				WithArgs(tc.email).
				WillReturnRows(rows).
				WillReturnError(tc.mockError)

			ctx := context.Background()

			var user models.User
			user, err = storage.User(ctx, tc.email)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.mockUser, user)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestApp(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &postgres.Storage{DB: db}

	testCases := []struct {
		name        string
		appID       int64
		mockApp     models.App
		mockError   error
		expectedErr error
	}{
		{
			"App Found",
			1,
			models.App{
				ID:     1,
				Name:   "TestApp",
				Secret: "secret",
			},
			nil,
			nil,
		},
		{
			"App Not Found",
			99,
			models.App{},
			sql.ErrNoRows,
			errors.New("app not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"id", "name", "secret"})
			if tc.mockError == nil {
				rows.AddRow(tc.mockApp.ID, tc.mockApp.Name, tc.mockApp.Secret)
			}
			mock.ExpectQuery(`SELECT id, name, secret FROM apps WHERE id = \$1`).
				WithArgs(tc.appID).
				WillReturnRows(rows).
				WillReturnError(tc.mockError)

			ctx := context.Background()
			var app models.App
			app, err = storage.App(ctx, tc.appID)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.mockApp, app)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestIsAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	storage := &postgres.Storage{DB: db}

	testCases := []struct {
		name        string
		userID      int64
		mockResult  bool
		mockError   error
		expectedErr error
	}{
		{
			"Admin User",
			1,
			true,
			nil,
			nil,
		},
		{
			"Non-Admin User",
			2,
			false,
			nil,
			nil,
		},
		{
			"User Not Found",
			99,
			false,
			sql.ErrNoRows,
			errors.New("user not found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"is_admin"})
			if tc.mockError == nil {
				rows.AddRow(tc.mockResult)
			}
			mock.ExpectQuery(`SELECT is_admin FROM users WHERE id = \$1`).
				WithArgs(tc.userID).
				WillReturnRows(rows).
				WillReturnError(tc.mockError)

			ctx := context.Background()
			var isAdmin bool
			isAdmin, err = storage.IsAdmin(ctx, tc.userID)

			if tc.expectedErr == nil {
				require.NoError(t, err)
				assert.Equal(t, tc.mockResult, isAdmin)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr.Error())
			}
		})
	}
}
