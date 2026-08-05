package testutils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"mailForgeApi/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
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

func BuildTestList(name string) (*models.List, error) {
	publicID, err := randomPublicID()
	if err != nil {
		return nil, err
	}

	return &models.List{
		PublicID:    publicID,
		Name:        name,
		Description: "for wedding invites to well wishers",
		Status:      "active",
		CreatedAt:   time.Now(),
	}, nil
}

func BuildTestUserWithPassword(plaintextPassword string) (*models.User, error) {
	user, err := BuildTestUser()
	if err != nil {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = string(hashed)

	return user, nil
}
