package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// writeTestKeyPair generates a real (throwaway) RSA key pair and writes it as
// PEM files inside dir. Returns the paths to the private and public key files.
// This is complete — you don't need to touch it, just call it from your tests.
func writeTestKeyPair(t *testing.T, dir string) (privPath, pubPath string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generating test RSA key: %v", err)
	}

	privBytes := x509.MarshalPKCS1PrivateKey(key)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshaling test public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	privPath = filepath.Join(dir, "private.pem")
	pubPath = filepath.Join(dir, "public.pem")

	if err := os.WriteFile(privPath, privPEM, 0600); err != nil {
		t.Fatalf("writing test private key: %v", err)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0644); err != nil {
		t.Fatalf("writing test public key: %v", err)
	}

	return privPath, pubPath
}

func TestLoadPrivateKey(t *testing.T) {
	// Step 1: get a real, valid key pair written to a temp dir.
	// t.TempDir() is unique per test and auto-deleted when the test finishes.
	validPrivPath, _ := writeTestKeyPair(t, t.TempDir())

	// Step 2: write one broken file by hand, for the "malformed pem" case.
	// This is just garbage bytes — not a PEM block at all.
	malformedPath := filepath.Join(t.TempDir(), "malformed.pem")
	if err := os.WriteFile(malformedPath, []byte("this is not a key"), 0600); err != nil {
		t.Fatalf("writing malformed key fixture: %v", err)
	}

	// Step 3: the table. Each row is one scenario: a path in, and whether
	// we expect LoadPrivateKey to return an error for it.
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid key",
			path:    validPrivPath,
			wantErr: false,
		},
		{
			name:    "missing file",
			path:    filepath.Join(t.TempDir(), "does-not-exist.pem"),
			wantErr: true,
		},
		{
			name:    "malformed pem",
			path:    malformedPath,
			wantErr: true,
		},
	}

	// Step 4: run each row as its own subtest. t.Run gives each case its own
	// name in test output (e.g. TestLoadPrivateKey/valid_key), so when one
	// fails you know exactly which scenario broke without reading the code.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := LoadPrivateKey(tt.path)

			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Fatalf("LoadPrivateKey(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}

			// Only makes sense to check the key itself when we expected success.
			if !tt.wantErr && key == nil {
				t.Fatalf("LoadPrivateKey(%q) returned nil key with no error", tt.path)
			}
		})
	}
}

func TestLoadPublicKey(t *testing.T) {
	// Same shape as TestLoadPrivateKey — this is the pattern you'll reuse
	// constantly: generate valid fixture, generate broken fixture, table, loop.
	_, validPubPath := writeTestKeyPair(t, t.TempDir())

	malformedPath := filepath.Join(t.TempDir(), "malformed.pem")
	if err := os.WriteFile(malformedPath, []byte("this is not a key"), 0600); err != nil {
		t.Fatalf("writing malformed key fixture: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid key",
			path:    validPubPath,
			wantErr: false,
		},
		{
			name:    "missing file",
			path:    filepath.Join(t.TempDir(), "does-not-exist.pem"),
			wantErr: true,
		},
		{
			name:    "malformed pem",
			path:    malformedPath,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := LoadPublicKey(tt.path)

			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Fatalf("LoadPublicKey(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}

			if !tt.wantErr && key == nil {
				t.Fatalf("LoadPublicKey(%q) returned nil key with no error", tt.path)
			}
		})
	}
}
