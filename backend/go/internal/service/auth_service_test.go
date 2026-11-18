package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/sentinel-device-manager/backend/go/internal/middleware"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
)

type fakeUsers struct {
	users map[string]*model.User
}

func (f *fakeUsers) GetByUsername(_ context.Context, username string) (*model.User, error) {
	u, ok := f.users[username]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUsers) Put(_ context.Context, user *model.User) error {
	if f.users == nil {
		f.users = map[string]*model.User{}
	}
	copied := *user
	f.users[user.Username] = &copied
	return nil
}

func newTestService(t *testing.T) *AuthService {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse-battery"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := &fakeUsers{users: map[string]*model.User{
		"admin": {UserID: "u-1", Username: "admin", PasswordHash: string(hash), Role: "admin"},
	}}
	return NewAuthService(users, "0123456789abcdef0123456789abcdef0123", "sentinel-device-manager")
}

func TestLoginIssuesTokenAcceptedByMiddleware(t *testing.T) {
	svc := newTestService(t)
	token, user, err := svc.Login(context.Background(), "admin", "correct-horse-battery")
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if user.Role != "admin" {
		t.Fatalf("role = %q", user.Role)
	}

	// The middleware is the real consumer of the token: run it through.
	called := false
	protected := middleware.JWTAuth(middleware.JWTConfig{
		SecretKey: "0123456789abcdef0123456789abcdef0123",
		Issuer:    "sentinel-device-manager",
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		claims := middleware.GetClaims(r)
		if claims == nil || claims.UserID != "u-1" || claims.Role != "admin" {
			t.Errorf("unexpected claims: %+v", claims)
		}
	}))

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	protected.ServeHTTP(httptest.NewRecorder(), req)
	if !called {
		t.Fatal("middleware rejected the issued token")
	}
}

func TestLoginRejectsBadCredentials(t *testing.T) {
	svc := newTestService(t)

	if _, _, err := svc.Login(context.Background(), "admin", "wrong"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong password: err = %v", err)
	}
	if _, _, err := svc.Login(context.Background(), "ghost", "correct-horse-battery"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown user: err = %v", err)
	}
	if _, _, err := svc.Login(context.Background(), "  ", ""); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("empty input: err = %v", err)
	}
}

func TestEnsureAdminSeedCreatesOnceOnly(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()

	if err := svc.EnsureAdminSeed(ctx, "boot", "at-least-12-chars"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// Seeding is skipped when the password is too short or the user exists.
	if err := svc.EnsureAdminSeed(ctx, "boot", "different-but-long-enough"); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	token, user, err := svc.Login(ctx, "boot", "at-least-12-chars")
	if err != nil || user.Role != "admin" || token == "" {
		t.Fatalf("seeded admin cannot log in: %v", err)
	}
}

func TestClaimsPayloadShape(t *testing.T) {
	svc := newTestService(t)
	token, _, err := svc.Login(context.Background(), "admin", "correct-horse-battery")
	if err != nil {
		t.Fatal(err)
	}

	parts := splitToken(token)
	if len(parts) != 3 {
		t.Fatalf("token has %d parts", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims struct {
		Sub  string `json:"sub"`
		Role string `json:"role"`
		Iss  string `json:"iss"`
		Nbf  int64  `json:"nbf"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims.Sub != "u-1" || claims.Role != "admin" || claims.Iss != "sentinel-device-manager" || claims.Nbf == 0 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func splitToken(token string) []string {
	out := []string{}
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			out = append(out, token[start:i])
			start = i + 1
		}
	}
	out = append(out, token[start:])
	return out
}
