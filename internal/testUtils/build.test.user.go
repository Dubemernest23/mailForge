package testutils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mailForgeApi/internal/models"
	"time"
)

func randomPublicID() (string, error) {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func BuildTestUser() (*models.User, error) {

	publicID, err := randomPublicID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixNano()

	return &models.User{
		PublicId: publicID,
		Username: fmt.Sprintf("user_%d", now),
		Email:    fmt.Sprintf("user_%d@example.com", now),

		PasswordHash: "hashed-password",

		Role: "user",

		Status: "active",

		FailedLoginAttempts: 0,
	}, nil
}
