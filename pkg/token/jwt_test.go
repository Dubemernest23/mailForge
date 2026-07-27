package tokens

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test key: %v", err)
	}
	return key, &key.PublicKey
}

func TestGenerateAndVerifyAccessToken(t *testing.T) {
	priv, pub := testKeyPair(t)
	require.NotNil(t, priv)

	token, err := GenerateAccessToken(priv, "u-234", "user", time.Hour)
	require.NoError(t, err)

	claims, err := VerifyAccessToken(pub, token)
	require.NoError(t, err)

	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "u-234", claims.Subject)

}

func TestVerifyAccessToken_Expired(t *testing.T) {
	priv, pub := testKeyPair(t)
	require.NotNil(t, priv)

	token, err := GenerateAccessToken(priv, "u-234", "user", -time.Hour)
	require.NoError(t, err)

	_, err = VerifyAccessToken(pub, token)

	assert.Error(t, err)
}

func TestVerifyAccessToken_Tampered(t *testing.T) {
	priv, pub := testKeyPair(t)
	require.NotNil(t, priv)

	token, err := GenerateAccessToken(priv, "u-234", "user", time.Hour)
	require.NoError(t, err)

	tamperedToken := token[:len(token)-1] + "v"

	_, err = VerifyAccessToken(pub, tamperedToken)
	assert.Error(t, err)
}
