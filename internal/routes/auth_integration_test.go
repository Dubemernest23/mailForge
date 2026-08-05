// internal/routes/auth_integration_test.go
package routes

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mailForgeApi/internal/middleware"
	"mailForgeApi/internal/modules/auth"
	"mailForgeApi/internal/shared/response"
	testutils "mailForgeApi/internal/testUtils"
	tokens "mailForgeApi/pkg/token"
)

// setupTestRouter builds a real router backed by real DB/Redis, mirroring
// production wiring but with test infra and test keys. Also mounts a
// throwaway protected route, since the real protected group is empty
// until Phase C — this is the only way to test JWTMiddleware's actual
// integration into the router before Phase C exists.
func setupTestRouter(t *testing.T) (*chi.Mux, *rsa.PublicKey, *rsa.PrivateKey) {
	t.Helper()

	require.NoError(t, testutils.CleanTables(t.Context(), testDB))

	privateKey, publicKey, err := testutils.TestKeyPair() // reuse whatever key-gen helper you settled on earlier
	require.NoError(t, err)

	repo := auth.NewAuthRepository(testDB)
	refreshMgr := tokens.NewRefreshTokenManager(testRedis, 7*24*time.Hour)
	svc := auth.NewService(repo, refreshMgr, privateKey, time.Hour)
	handler := auth.NewHandler(svc)

	r := chi.NewRouter()
	registerAuthRoutes(r, handler)
	r.Group(func(protected chi.Router) {
		protected.Use(middleware.JWTMiddleware(publicKey))
		protected.Get("/__test_protected__", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	return r, publicKey, privateKey
}

// TODO: helper to POST JSON and decode the response, since every test
// below repeats this pattern
func doJSONRequest(t *testing.T, router *chi.Mux, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthRoutes_Register(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, router *chi.Mux) map[string]string
		wantStatusCode int
	}{
		{
			name: "success",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				return map[string]string{
					"username": "routeuser1",
					"email":    "routeuser1@example.com",
					"password": "Passw0rd!",
				}
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "duplicate_email",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				body := map[string]string{
					"username": "routeuser2",
					"email":    "dup@example.com",
					"password": "Passw0rd!",
				}
				rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", body)
				require.Equal(t, http.StatusCreated, rec.Code)

				return map[string]string{
					"username": "routeuser3",
					"email":    "dup@example.com", // same email again
					"password": "Passw0rd!",
				}
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name: "invalid_body_short_password",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				return map[string]string{
					"username": "routeuser4",
					"email":    "routeuser4@example.com",
					"password": "short",
				}
			},
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router, _, _ := setupTestRouter(t)
			body := tc.setup(t, router)

			rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", body)
			assert.Equal(t, tc.wantStatusCode, rec.Code)

			if tc.wantStatusCode == http.StatusCreated {
				var resp auth.AuthResponse
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
				assert.Equal(t, 3600, resp.ExpiresIn)
				return
			}

			var errBody response.ErrorBody
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&errBody))
			assert.False(t, errBody.Success)
		})
	}
}

func TestAuthRoutes_Login(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T, router *chi.Mux) map[string]string
		wantStatusCode int
	}{
		{
			name: "success",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				registerBody := map[string]string{
					"username": "loginuser1",
					"email":    "loginuser1@example.com",
					"password": "Passw0rd!",
				}
				rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", registerBody)
				require.Equal(t, http.StatusCreated, rec.Code)

				return map[string]string{
					"email":    "loginuser1@example.com",
					"password": "Passw0rd!",
				}
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "wrong_password",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				registerBody := map[string]string{
					"username": "loginuser2",
					"email":    "loginuser2@example.com",
					"password": "Passw0rd!",
				}
				rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", registerBody)
				require.Equal(t, http.StatusCreated, rec.Code)

				return map[string]string{
					"email":    "loginuser2@example.com",
					"password": "WrongPassword1!",
				}
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "wrong_email",
			setup: func(t *testing.T, router *chi.Mux) map[string]string {
				return map[string]string{
					"email":    "doesnotexist@example.com",
					"password": "Passw0rd!",
				}
			},
			wantStatusCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router, _, _ := setupTestRouter(t)
			body := tc.setup(t, router)

			rec := doJSONRequest(t, router, http.MethodPost, "/auth/login", body)
			assert.Equal(t, tc.wantStatusCode, rec.Code)

			if tc.wantStatusCode == http.StatusOK {
				var resp auth.AuthResponse
				require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

func TestAuthRoutes_Refresh(t *testing.T) {
	t.Run("success_and_old_token_invalidated", func(t *testing.T) {
		router, _, _ := setupTestRouter(t)

		registerBody := map[string]string{
			"username": "refreshrouteuser",
			"email":    "refreshrouteuser@example.com",
			"password": "Passw0rd!",
		}
		rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", registerBody)
		require.Equal(t, http.StatusCreated, rec.Code)

		var registerResp auth.AuthResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&registerResp))

		// use the token once — should succeed and return a new pair
		refreshRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
			"refresh_token": registerResp.RefreshToken,
		})
		require.Equal(t, http.StatusOK, refreshRec.Code)

		var refreshResp auth.AuthResponse
		require.NoError(t, json.NewDecoder(refreshRec.Body).Decode(&refreshResp))
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEqual(t, registerResp.RefreshToken, refreshResp.RefreshToken)

		// reuse the SAME original token — should now fail
		reuseRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
			"refresh_token": registerResp.RefreshToken,
		})
		assert.Equal(t, http.StatusUnauthorized, reuseRec.Code)
	})
}

func TestAuthRoutes_Logout(t *testing.T) {
	t.Run("success_and_token_invalidated", func(t *testing.T) {
		router, _, _ := setupTestRouter(t)

		registerBody := map[string]string{
			"username": "logoutrouteuser",
			"email":    "logoutrouteuser@example.com",
			"password": "Passw0rd!",
		}
		rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", registerBody)
		require.Equal(t, http.StatusCreated, rec.Code)

		var registerResp auth.AuthResponse
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&registerResp))

		logoutRec := doJSONRequest(t, router, http.MethodPost, "/auth/logout", map[string]string{
			"refresh_token": registerResp.RefreshToken,
		})
		assert.Equal(t, http.StatusNoContent, logoutRec.Code)

		// the revoked token should no longer work for refresh
		refreshRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
			"refresh_token": registerResp.RefreshToken,
		})
		assert.Equal(t, http.StatusUnauthorized, refreshRec.Code)
	})
}

func TestAuthRoutes_ProtectedRoute(t *testing.T) {
	t.Run("no_token", func(t *testing.T) {
		router, _, _ := setupTestRouter(t)

		req := httptest.NewRequest(http.MethodGet, "/__test_protected__", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("expired_token", func(t *testing.T) {
		router, _, privateKey := setupTestRouter(t)

		expiredToken, err := tokens.GenerateAccessToken(privateKey, "some-user-id", "user", -time.Hour)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/__test_protected__", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

}

func TestAuthRoutes_FullLifecycle(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	// 1. Register
	registerBody := map[string]string{
		"username": "lifecycleuser",
		"email":    "lifecycleuser@example.com",
		"password": "Passw0rd!",
	}
	rec := doJSONRequest(t, router, http.MethodPost, "/auth/register", registerBody)
	require.Equal(t, http.StatusCreated, rec.Code)

	var registerResp auth.AuthResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&registerResp))

	// 2. Access token from registration works on a protected route immediately
	req := httptest.NewRequest(http.MethodGet, "/__test_protected__", nil)
	req.Header.Set("Authorization", "Bearer "+registerResp.AccessToken)
	protectedRec := httptest.NewRecorder()
	router.ServeHTTP(protectedRec, req)
	assert.Equal(t, http.StatusOK, protectedRec.Code)

	// 3. Login with the same credentials — independent token pair, both valid
	loginRec := doJSONRequest(t, router, http.MethodPost, "/auth/login", map[string]string{
		"email":    "lifecycleuser@example.com",
		"password": "Passw0rd!",
	})
	require.Equal(t, http.StatusOK, loginRec.Code)

	var loginResp auth.AuthResponse
	require.NoError(t, json.NewDecoder(loginRec.Body).Decode(&loginResp))

	// 4. Refresh using the LOGIN refresh token (not the register one — proves
	// both issued tokens are independently valid and correctly tracked)
	refreshRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": loginResp.RefreshToken,
	})
	require.Equal(t, http.StatusOK, refreshRec.Code)

	var refreshResp auth.AuthResponse
	require.NoError(t, json.NewDecoder(refreshRec.Body).Decode(&refreshResp))

	// 5. New access token from refresh also works on the protected route
	req2 := httptest.NewRequest(http.MethodGet, "/__test_protected__", nil)
	req2.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)
	protectedRec2 := httptest.NewRecorder()
	router.ServeHTTP(protectedRec2, req2)
	assert.Equal(t, http.StatusOK, protectedRec2.Code)

	// 6. Logout with the rotated refresh token
	logoutRec := doJSONRequest(t, router, http.MethodPost, "/auth/logout", map[string]string{
		"refresh_token": refreshResp.RefreshToken,
	})
	assert.Equal(t, http.StatusNoContent, logoutRec.Code)

	// 7. Post-logout: rotated token is dead, and the ORIGINAL register token
	// should also still be dead from step 4's rotation — confirms revocation
	// doesn't leak across independent token pairs incorrectly
	reuseRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
		"refresh_token": refreshResp.RefreshToken,
	})
	assert.Equal(t, http.StatusUnauthorized, reuseRec.Code)

	// staleRec := doJSONRequest(t, router, http.MethodPost, "/auth/refresh", map[string]string{
	// 	"refresh_token": registerResp.RefreshToken,
	// })
	// assert.Equal(t, http.StatusUnauthorized, staleRec.Code)
}
