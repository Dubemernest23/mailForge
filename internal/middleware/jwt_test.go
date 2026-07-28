package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mailForgeApi/internal/response"
	tokens "mailForgeApi/pkg/token"
)

func generateTestKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, &key.PublicKey
}

func stubNext(called *bool, gotUserID, gotRole *string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		if uid, ok := UserIDFromContext(r.Context()); ok {
			*gotUserID = uid
		}
		if role, ok := RoleFromContext(r.Context()); ok {
			*gotRole = role
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestJWTMiddleware(t *testing.T) {
	privateKey, publicKey := generateTestKeyPair(t)
	otherPrivateKey, _ := generateTestKeyPair(t)

	validToken, err := tokens.GenerateAccessToken(privateKey, "user-123", "admin", time.Hour)
	require.NoError(t, err)

	expiredToken, err := tokens.GenerateAccessToken(privateKey, "user-123", "admin", -time.Hour)
	require.NoError(t, err)

	wrongSignerToken, err := tokens.GenerateAccessToken(otherPrivateKey, "user-123", "admin", time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
		wantNextCalled bool
	}{
		{"valid_token", "Bearer " + validToken, http.StatusOK, true},
		{"missing_header", "", http.StatusUnauthorized, false},
		{"malformed_header_no_bearer_prefix", validToken, http.StatusUnauthorized, false},
		{"expired_token", "Bearer " + expiredToken, http.StatusUnauthorized, false},
		{"wrong_signer_token", "Bearer " + wrongSignerToken, http.StatusUnauthorized, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()

			var nextCalled bool
			var gotUserID, gotRole string
			handler := JWTMiddleware(publicKey)(stubNext(&nextCalled, &gotUserID, &gotRole))

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantStatusCode, rec.Code)
			assert.Equal(t, tc.wantNextCalled, nextCalled)

			if tc.name == "valid_token" {
				assert.Equal(t, "user-123", gotUserID)
				assert.Equal(t, "admin", gotRole)
				return
			}

			// every 401 case should carry the standard error JSON shape
			var body response.ErrorBody
			require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
			assert.False(t, body.Success)
			assert.Equal(t, response.CodeUnauthorized, body.Error.Code)
			assert.NotEmpty(t, body.Error.Message)
		})
	}
}
