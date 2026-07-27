package tokens

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testRedisClient *redis.Client

func TestMain(m *testing.M) {
	testRedisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // or read from an env var, your call
	})

	if err := testRedisClient.Ping(context.Background()).Err(); err != nil {
		panic("cannot connect to test redis: " + err.Error())
	}

	code := m.Run()
	_ = testRedisClient.Close()
	os.Exit(code)
}

func TestValidateAndRotate_SingleUse(t *testing.T) {
	manager := NewRefreshTokenManager(testRedisClient, time.Hour)
	ctx := context.Background()

	originalToken, err := manager.IssueRefreshToken(ctx, "234-345-rtgf")
	require.NoError(t, err)

	newToken, UserId, err := manager.ValidateAndRotate(ctx, originalToken)
	require.NoError(t, err)
	assert.Equal(t, "234-345-rtgf", UserId)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, originalToken, newToken)

	_, _, err = manager.ValidateAndRotate(ctx, originalToken)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}
