package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadPrivateKey reads a PEM-encoded RSA private key from disk and parses it.
// Supports both PKCS1 ("RSA PRIVATE KEY") and PKCS8 ("PRIVATE KEY") encodings
// since gen-keys' underlying openssl invocation could change between environments.
func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading private key file %q: %w", path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in private key file %q", path)
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// openssl genrsa (used by `make gen-keys`) writes PKCS1 ("RSA PRIVATE KEY").
	// Try that first; fall back to PKCS8 in case the key was generated differently.
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing private key %q as PKCS1 or PKCS8: %w", path, err)
	}

	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key %q is not an RSA key", path)
	}

	return rsaKey, nil
}

// LoadPublicKey reads a PEM-encoded RSA public key from disk and parses it.
// Expects PKIX encoding ("PUBLIC KEY"), which is what `openssl rsa -pubout` produces.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading public key file %q: %w", path, err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in public key file %q", path)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key %q: %w", path, err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key %q is not an RSA key", path)
	}

	return rsaKey, nil
}
