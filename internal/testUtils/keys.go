// testUtils/keys.go
package testutils

import (
	"crypto/rand"
	"crypto/rsa"
)

// TestKeyPair returns a fresh in-memory RSA key pair for tests.
// 2048 bits is plenty for tests — no need to match production's key size
// if it's larger, since these keys never leave the test process.
func TestKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}
