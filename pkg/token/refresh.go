package tokens

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// refreshKeyPrefix namespaces refresh token keys in Redis so they're easy to
// spot/scan/expire distinctly from any other key space (e.g. asynq's own keys)
// sharing the same Redis instance.
const refreshKeyPrefix = "refresh:"

// ErrInvalidRefreshToken is returned when a refresh token doesn't exist in Redis —
// either it was never issued, it already expired, or (most importantly) it was
// already used once before. GETDEL makes reuse indistinguishable from "never existed",
// which is exactly the single-use guarantee we want.
var ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")

type RefreshTokenManager interface {
	IssueRefreshToken(ctx context.Context, userID string) (string, error)
	ValidateAndRotate(ctx context.Context, token string) (newToken string, userID string, err error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type refreshTokenManager struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRefreshTokenManager returns a RefreshTokenManager. ttl should be the parsed
// duration corresponding to JWT_REFRESH_EXPIRY (e.g. 7 * 24 * time.Hour) — resolved
// by the caller (D5's service layer), same reasoning as GenerateAccessToken's expiry param.
func NewRefreshTokenManager(client *redis.Client, ttl time.Duration) RefreshTokenManager { // this is a constructor
	return &refreshTokenManager{client: client, ttl: ttl}
}

// IssueRefreshToken method generates a fresh opaque UUID, stores it in Redis as
// refresh:<uuid> -> userID with the configured TTL, and returns the UUID
// (this is what actually gets handed to the client — never a JWT).
func (m *refreshTokenManager) IssueRefreshToken(ctx context.Context, userID string) (string, error) { // a method
	// TODO 1: generate a new UUID string — uuid.NewString() (from google/uuid).
	rId := uuid.NewString()

	// TODO 2: build the Redis key: refreshKeyPrefix + the UUID from TODO 1.
	key := refreshKeyPrefix + rId

	// TODO 3: client.Set(ctx, key, userID, m.ttl).Err() — check the error.
	err := m.client.Set(ctx, key, userID, m.ttl).Err()
	if err != nil {
		return "", err
	}
	// TODO 4: return the UUID string (not the key) and nil error on success.
	return rId, nil
}

// ValidateAndRotate implements the exact rotation flow from the PDR:
//  1. GETDEL refresh:<token> — atomically read and delete in one command
//  2. If nil (key didn't exist) -> reject: token doesn't exist or was already used
//  3. If found -> the returned value IS the userID -> issue a new token pair for it
func (m *refreshTokenManager) ValidateAndRotate(ctx context.Context, token string) (string, string, error) {
	key := refreshKeyPrefix + token

	// TODO 5: call m.client.GetDel(ctx, key).Result() — this returns (string, error).
	// The string result, if no error, IS the userID that was stored — that's the
	// whole point of storing userID as the value back in IssueRefreshToken.
	result, err := m.client.GetDel(ctx, key).Result()

	// TODO 6: check the error from TODO 5. Specifically check errors.Is(err, redis.Nil) —
	// if true, return "", "", ErrInvalidRefreshToken (this is the "already used or
	// never existed" case — GETDEL can't tell these apart, which is correct behavior).
	// If it's some OTHER non-nil error (e.g. Redis connection failure), return that
	// raw error instead — don't mask a real infrastructure failure as an invalid token.
	if errors.Is(err, redis.Nil) {
		return "", "", ErrInvalidRefreshToken
	}

	if err != nil {
		return "", "", err
	}

	// TODO 7: on success, call m.IssueRefreshToken(ctx, userID) to get a new token,
	// using the userID you extracted in TODO 5. Return (newToken, userID, err).
	newToken, err := m.IssueRefreshToken(ctx, result)
	if err != nil {
		return "", "", err
	}

	return newToken, result, nil
}

// RevokeRefreshToken deletes a refresh token outright — used on logout.
// Unlike ValidateAndRotate, we don't care what (if anything) was stored there,
// just that it's gone afterward.
func (m *refreshTokenManager) RevokeRefreshToken(ctx context.Context, token string) error {
	key := refreshKeyPrefix + token

	// TODO 8: m.client.Del(ctx, key).Err() — check and return the error.
	// Think about whether "the key didn't exist" should even count as an error here —
	// does Del return an error in that case, or just report 0 keys deleted?
	err := m.client.Del(ctx, key).Err()

	if err != nil {
		return err
	}

	return nil
}
