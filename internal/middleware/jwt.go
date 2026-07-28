package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"mailForgeApi/internal/response"
	tokens "mailForgeApi/pkg/token"
)

type contextKey int

const (
	userIDKey contextKey = iota
	roleKey
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	// TODO 1: ctx.Value(userIDKey), type-assert to string, return (value, ok)
	ctxValue := ctx.Value(userIDKey)
	if ctxValue == nil {
		return "", false
	}
	if userID, ok := ctxValue.(string); ok {
		return userID, true
	}
	return "", false
}

func RoleFromContext(ctx context.Context) (string, bool) {
	// TODO 2: same pattern as TODO 1, but for roleKey
	ctxValue := ctx.Value(roleKey)
	if ctxValue == nil {
		return "", false
	}
	if role, ok := ctxValue.(string); ok {
		return role, true
	}
	return "", false
}

func JWTMiddleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO 3: header := r.Header.Get("Authorization")
			header := r.Header.Get("Authorization")
			// TODO 4: if header == "" -> response.Unauthorized(w, r, "missing authorization header"); return
			if header == "" {
				response.Unauthorized(w, r, "missing authorization header")
				return
			}
			// TODO 5: if !strings.HasPrefix(header, "Bearer ") ->
			//   response.Unauthorized(w, r, "malformed authorization header"); return
			if !strings.HasPrefix(header, "Bearer ") {
				response.Unauthorized(w, r, "malformed authorization header")
				return
			}
			// TODO 6: tokenString := strings.TrimPrefix(header, "Bearer ")
			tokenString := strings.TrimPrefix(header, "Bearer ")
			// TODO 7: claims, err := tokens.VerifyAccessToken(publicKey, tokenString)
			//   if err != nil -> response.Unauthorized(w, r, "invalid or expired token"); return
			claims, err := tokens.VerifyAccessToken(publicKey, tokenString)
			if err != nil {
				response.Unauthorized(w, r, "invalid or expired token")
				return
			}
			// TODO 8: chain context.WithValue twice — userIDKey then roleKey —
			//   using claims.Subject and claims.Role. Don't discard the first WithValue's result.
			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			// TODO 9: r = r.WithContext(ctx)
			r = r.WithContext(ctx)
			// TODO 10: next.ServeHTTP(w, r)
			next.ServeHTTP(w, r)
		})
	}
}
