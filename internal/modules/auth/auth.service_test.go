package auth

import (
	"context"
	"mailForgeApi/internal/shared/apperrors"
	testutils "mailForgeApi/internal/testUtils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// **Tests:** Unit tests targeting **90% coverage of the service layer**, per PDR §14.4's Phase B target specifically.
// Cover every happy and error path: successful register, duplicate email, successful login, wrong password, wrong email, successful refresh,
// expired/reused refresh token, logout.

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) RegisterRequest
		expectedErr error
	}{
		{
			name: "successful_signup",
			setup: func(t *testing.T) RegisterRequest {
				return RegisterRequest{
					Email:    "firstuser1@gmail.com",
					Username: "firstuser",
					Password: "firstuser1@g",
				}
			},
			expectedErr: nil,
		},
		{
			name: "duplicate_email",
			setup: func(t *testing.T) RegisterRequest {
				existing, err := testutils.CreateTestUser(testDB)
				require.NoError(t, err)
				return RegisterRequest{
					Email:    existing.Email,
					Username: "firstuser",
					Password: "firstuser1@g",
				}
			},
			expectedErr: apperrors.ErrDuplicate,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := setupService(t)
			ctx := context.Background()

			req := tc.setup(t)
			resp, err := svc.Register(ctx, req)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			assert.Equal(t, 3600, resp.ExpiresIn)
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) LoginRequest
		expectedErr error
	}{
		{
			name: "successful_login",
			setup: func(t *testing.T) LoginRequest {
				user, err := testutils.CreateTestUserWithPassword(testDB, "firstuser1@g")
				require.NoError(t, err)
				return LoginRequest{Email: user.Email, Password: "firstuser1@g"}
			},
			expectedErr: nil,
		},
		{
			name: "wrong_password",
			setup: func(t *testing.T) LoginRequest {
				user, err := testutils.CreateTestUserWithPassword(testDB, "firstuser1@g")
				require.NoError(t, err)
				return LoginRequest{Email: user.Email, Password: "wrong-password"}
			},
			expectedErr: apperrors.ErrUnauthorized,
		},
		{
			name: "wrong_email",
			setup: func(t *testing.T) LoginRequest {
				return LoginRequest{Email: "does-not-exist@example.com", Password: "irrelevant"}
			},
			expectedErr: apperrors.ErrUnauthorized,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := setupService(t)
			ctx := context.Background()
			req := tc.setup(t)
			resp, err := svc.Login(ctx, req)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			assert.Equal(t, 3600, resp.ExpiresIn)
		})
	}
}

func TestRefreshReq(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T, svc *Service) RefreshRequest
		expectedErr error
	}{
		{
			name: "successful_refresh",
			setup: func(t *testing.T, svc *Service) RefreshRequest {
				req := RegisterRequest{
					Username: "refreshuser",
					Email:    "refreshuser@example.com",
					Password: "password123",
				}
				resp, err := svc.Register(context.Background(), req)
				require.NoError(t, err)
				return RefreshRequest{RefreshToken: resp.RefreshToken}
			},
			expectedErr: nil,
		},
		{
			name: "reused_refresh_token",
			setup: func(t *testing.T, svc *Service) RefreshRequest {
				req := RegisterRequest{
					Username: "reuseuser",
					Email:    "reuseuser@example.com",
					Password: "password123",
				}
				resp, err := svc.Register(context.Background(), req)
				require.NoError(t, err)

				// use it once — this rotates and deletes the original token
				_, err = svc.Refresh(context.Background(), RefreshRequest{RefreshToken: resp.RefreshToken})
				require.NoError(t, err)

				// return the SAME (now-consumed) token again
				return RefreshRequest{RefreshToken: resp.RefreshToken}
			},
			expectedErr: apperrors.ErrInvalidRefreshToken,
		},
		{
			name: "nonexistent_token",
			setup: func(t *testing.T, svc *Service) RefreshRequest {
				return RefreshRequest{RefreshToken: "totally-made-up-token"}
			},
			expectedErr: apperrors.ErrInvalidRefreshToken,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := setupService(t)
			req := tc.setup(t, svc)

			resp, err := svc.Refresh(context.Background(), req)

			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			assert.Equal(t, 3600, resp.ExpiresIn)
		})
	}
}

func TestLogOutReq(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, svc *Service) LogoutRequest
		wantErr bool
	}{
		{
			name: "successful_logout",
			setup: func(t *testing.T, svc *Service) LogoutRequest {
				req := RegisterRequest{
					Username: "logoutuser",
					Email:    "logoutuser@example.com",
					Password: "password123",
				}
				resp, err := svc.Register(context.Background(), req)
				require.NoError(t, err)
				return LogoutRequest{RefreshToken: resp.RefreshToken}
			},
			wantErr: false,
		},
		{
			name: "logout_nonexistent_token",
			setup: func(t *testing.T, svc *Service) LogoutRequest {
				return LogoutRequest{RefreshToken: "never-issued-token"}
			},
			wantErr: false, // RevokeRefreshToken is idempotent by design — see D4
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			svc := setupService(t)
			req := tc.setup(t, svc)

			err := svc.Logout(context.Background(), req)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}
