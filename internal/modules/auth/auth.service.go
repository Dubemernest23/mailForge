package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"log"

	// "fmt"
	"time"

	"mailForgeApi/internal/apperrors"
	"mailForgeApi/internal/models"
	tokens "mailForgeApi/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost int = 12

type Service struct {
	repo         AuthRepo
	refreshMgr   tokens.RefreshTokenManager
	privateKey   *rsa.PrivateKey
	accessExpiry time.Duration
}

// NewService is the Fx-injectable constructor.
// privateKey and accessExpiry come from D1's parsed config —
// verification (publicKey) belongs to D6's middleware, not here.
func NewService(repo AuthRepo, refreshMgr tokens.RefreshTokenManager, privateKey *rsa.PrivateKey, accessExpiry time.Duration) *Service {
	return &Service{
		repo:         repo,
		refreshMgr:   refreshMgr,
		privateKey:   privateKey,
		accessExpiry: accessExpiry,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// hashpassword -> build model and create user -> issue access and refresh token -> return auth response
	// TODO 1: bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
	if err != nil {
		return nil, err
	}
	// TODO 2: build a *models.User from req (username, email, hashed password)
	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}
	// TODO 3: s.repo.CreateUser(ctx, user) — if errors.Is(err, apperrors.ErrDuplicate), return nil, apperrors.ErrDuplicate
	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, apperrors.ErrDuplicate) {
			return nil, apperrors.ErrDuplicate
		}
		return nil, err
	}

	// TODO 4: issue access token via tokens.GenerateAccessToken(s.privateKey, user.PublicID, <role>, s.accessExpiry)
	accessToken, err := tokens.GenerateAccessToken(s.privateKey, user.PublicId, user.Role, s.accessExpiry)
	if err != nil {
		return nil, err
	}
	// TODO 5: issue refresh token via s.refreshMgr.IssueRefreshToken(ctx, user.PublicID)
	refreshToken, err := s.refreshMgr.IssueRefreshToken(ctx, user.PublicId)
	if err != nil {
		return nil, err
	}
	// TODO 6: return &AuthResponse{...}, nil
	// NOTE: do NOT call UpdateLastLogin here.
	return &AuthResponse{
		AccessToken:  string(accessToken),
		RefreshToken: string(refreshToken),
		ExpiresIn:    int(s.accessExpiry.Seconds()),
	}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// find user -> compare password -> update last login -> issue refresh and access token -> return auth response
	// TODO 1: s.repo.FindByEmail(ctx, req.Email)
	//   if errors.Is(err, apperrors.ErrNotFound) -> return nil, apperrors.ErrUnauthorized (generic, don't leak which)
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrUnauthorized
		}
		return nil, err
	}
	// TODO 2: bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(req.Password))
	//   if mismatch -> s.repo.IncrementFailedAttempts(ctx, user.PublicID) -> return nil, apperrors.ErrUnauthorized
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		if incErr := s.repo.IncrementFailedAttempts(ctx, user.PublicId); incErr != nil {
			// TODO: log incErr once logging is wired in — don't let it change the response
			// return nil, fmt.Errorf("incremental error: %v", incErr)
			log.Printf("failed to increment failed login attempts for user %s: %v", user.PublicId, incErr)
		}
		return nil, apperrors.ErrUnauthorized
	}
	// TODO 3: on success -> s.repo.UpdateLastLogin(ctx, user.PublicID)
	err = s.repo.UpdateLastLogin(ctx, user.PublicId)
	if err != nil {
		return nil, err
	}

	// TODO 4: issue access + refresh tokens, same as Register
	// TODO 5: return &AuthResponse{...}, nil

	accessToken, err := tokens.GenerateAccessToken(s.privateKey, user.PublicId, user.Role, s.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.refreshMgr.IssueRefreshToken(ctx, user.PublicId)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(s.accessExpiry.Seconds()),
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	// TODO 1: newToken, userID, err := s.refreshMgr.ValidateAndRotate(ctx, req.RefreshToken)
	//   if errors.Is(err, tokens.ErrInvalidRefreshToken) -> return nil, apperrors.ErrInvalidRefreshToken
	//   if err != nil (other) -> return nil, err  (real infra failure, don't mask it)
	// TODO 2: issue a NEW access token via tokens.GenerateAccessToken(s.privateKey, userID, <role>, s.accessExpiry)
	//   -- open question: role isn't stored in Redis, only userID. Where does role come from here?
	//   -- (worth resolving before you write this — see note below)
	// TODO 3: return &AuthResponse{AccessToken: ..., RefreshToken: newToken, ExpiresIn: ...}, nil

	newToken, userID, err := s.refreshMgr.ValidateAndRotate(ctx, req.RefreshToken)
	if errors.Is(err, tokens.ErrInvalidRefreshToken) {
		return nil, apperrors.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}

	// find user, so that we will have to get the role there
	user, err := s.repo.FindByPublicID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := tokens.GenerateAccessToken(s.privateKey, user.PublicId, user.Role, s.accessExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newToken,
		ExpiresIn:    int(s.accessExpiry.Seconds()),
	}, nil
}

func (s *Service) Logout(ctx context.Context, req LogoutRequest) error {
	// TODO 1: s.refreshMgr.RevokeRefreshToken(ctx, req.RefreshToken) — return its error directly.
	err := s.refreshMgr.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return err
	}

	return nil
}
