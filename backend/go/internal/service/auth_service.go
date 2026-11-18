// Package service contains the business logic layer of the Sentinel API.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
)

// ErrInvalidCredentials is returned for an unknown username or wrong password.
// The message is deliberately identical for both cases so logins cannot be
// used to enumerate accounts.
var ErrInvalidCredentials = errors.New("invalid username or password")

// TokenTTL is how long an issued access token remains valid.
const TokenTTL = 8 * time.Hour

// AuthService issues and verifies operator sessions.
type AuthService struct {
	users  repository.UserRepository
	secret []byte
	issuer string
}

func NewAuthService(users repository.UserRepository, secret string, issuer string) *AuthService {
	return &AuthService{users: users, secret: []byte(secret), issuer: issuer}
}

// Login verifies the credentials and returns a signed HS256 access token.
func (s *AuthService) Login(ctx context.Context, username, password string) (string, *model.User, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		return "", nil, ErrInvalidCredentials
	}

	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			// Constant-time-ish: still run a bcrypt comparison so a missing
			// account takes about as long as a wrong password.
			_ = bcrypt.CompareHashAndPassword(
				[]byte("$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5B0S4M0Zzq6JXbCkQ0nQZ8f9QZ1Je"), []byte(password))
			return "", nil, ErrInvalidCredentials
		}
		return "", nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return "", nil, ErrInvalidCredentials
	}

	now := time.Now().Unix()
	claims := map[string]any{
		"sub":   user.UserID,
		"role":  user.Role,
		"iss":   s.issuer,
		"iat":   now,
		"nbf":   now,
		"exp":   now + int64(TokenTTL/time.Second),
	}
	token, err := s.sign(claims)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

// EnsureAdminSeed creates the initial admin account from the environment on
// first boot so an operator can sign in without manual table edits. It never
// overwrites an existing user.
func (s *AuthService) EnsureAdminSeed(ctx context.Context, username, password string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || len(password) < 12 {
		return nil // seeding disabled; nothing to do
	}
	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return nil // already exists
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Put(ctx, &model.User{
		Username:     username,
		PasswordHash: string(hash),
		Role:         "admin",
	})
}

// sign produces a compact HS256 JWS, matching the format parseJWT validates.
func (s *AuthService) sign(claims map[string]any) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(signingInput))
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
