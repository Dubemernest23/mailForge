package tokens

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims carries exactly what the PDR specifies: sub, role, iat, exp.
// RegisteredClaims already provides Subject/IssuedAt/ExpiresAt with correct json tags.
type AccessTokenClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken signs a new RS256 access token for the given user.
// expiry is passed in rather than read from config, since pkg/tokens must
// stay decoupled from internal/config — the caller (D5's service layer)
// resolves the config string into a time.Duration before calling this.
func GenerateAccessToken(privateKey *rsa.PrivateKey, userID string, role string, expiry time.Duration) (string, error) {
	// TODO 1: build an AccessTokenClaims value.
	//   - Role: role
	//   - RegisteredClaims.Subject: userID
	//   - RegisteredClaims.IssuedAt: jwt.NewNumericDate(time.Now())
	//   - RegisteredClaims.ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry))

	// HEADER => PAYLOAD => SIGNATURE.
	claims := AccessTokenClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	// TODO 2: jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// TODO 3: call .SignedString(privateKey) on the token from TODO 2,
	// and return its (string, error) directly — no need to wrap it further.
	token, err := signedToken.SignedString(privateKey)

	if err != nil {
		return "", err
	}

	return token, nil
}

// VerifyAccessToken parses and validates an access token, returning its claims
// if — and only if — the signature is valid, the algorithm matches what we
// actually sign with, and the token hasn't expired.
func VerifyAccessToken(publicKey *rsa.PublicKey, tokenString string) (*AccessTokenClaims, error) {
	claims := &AccessTokenClaims{}

	// keyFunc is called by ParseWithClaims to obtain the key to verify with.
	// CRITICAL: don't just return publicKey — first confirm the token's signing
	// method is actually RSA. Skipping this check is a known JWT vulnerability
	// class (algorithm confusion / "alg: none" attacks).
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// TODO 4: type-assert token.Method to *jwt.SigningMethodRSA.
		// If the assertion fails, return an error (don't return publicKey).
		// If it succeeds, return publicKey, nil.

		if token.Method != jwt.SigningMethodRS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return publicKey, nil
	}

	// TODO 5: call jwt.ParseWithClaims(tokenString, claims, keyFunc).
	// It returns (*jwt.Token, error) — capture both.
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// TODO 6: if the error from TODO 5 is non-nil, return nil, err immediately.
	return claims, nil
}
