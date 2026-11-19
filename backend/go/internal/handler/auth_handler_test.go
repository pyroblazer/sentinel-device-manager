package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/sentinel-device-manager/backend/go/internal/middleware"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
	"github.com/sentinel-device-manager/backend/go/internal/service"
)

const (
	testSecret = "0123456789abcdef0123456789abcdef0123"
	testIssuer = "sentinel-device-manager"
)

type memUsers map[string]*model.User

func (m memUsers) GetByUsername(_ context.Context, username string) (*model.User, error) {
	u, ok := m[username]
	if !ok {
		return nil, repository.ErrUserNotFound
	}
	return u, nil
}

func (m memUsers) Put(_ context.Context, user *model.User) error {
	m[user.Username] = user
	return nil
}

func newAuthRouter(t *testing.T) http.Handler {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse-battery"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := memUsers{"admin": {UserID: "u-1", Username: "admin", PasswordHash: string(hash), Role: "admin"}}
	h := NewAuthHandler(service.NewAuthService(users, testSecret, testIssuer))

	r := chi.NewRouter()
	h.RegisterRoutes(r)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(middleware.JWTConfig{SecretKey: testSecret, Issuer: testIssuer}))
		h.RegisterProtectedRoutes(r)
	})
	return r
}

func postLogin(router http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLoginReturnsTokenAndUserWithoutHash(t *testing.T) {
	rec := postLogin(newAuthRouter(t), `{"username":"admin","password":"correct-horse-battery"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("response must never include password material: %s", rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
		User  struct {
			UserID   string `json:"user_id"`
			Username string `json:"username"`
			Role     string `json:"role"`
		} `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Token == "" || out.User.Username != "admin" || out.User.Role != "admin" {
		t.Fatalf("unexpected response: %+v", out)
	}
}

func TestLoginRejectsBadInput(t *testing.T) {
	router := newAuthRouter(t)
	cases := map[string]struct {
		body   string
		status int
	}{
		"wrong password": {`{"username":"admin","password":"nope"}`, http.StatusUnauthorized},
		"unknown user":   {`{"username":"ghost","password":"correct-horse-battery"}`, http.StatusUnauthorized},
		"malformed json": {`{not json`, http.StatusBadRequest},
	}
	for name, tc := range cases {
		if rec := postLogin(router, tc.body); rec.Code != tc.status {
			t.Errorf("%s: status = %d, want %d", name, rec.Code, tc.status)
		}
	}
}

func TestMeEchoesIdentityOfIssuedToken(t *testing.T) {
	router := newAuthRouter(t)
	rec := postLogin(router, `{"username":"admin","password":"correct-horse-battery"}`)
	var login struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &login); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+login.Token)
	me := httptest.NewRecorder()
	router.ServeHTTP(me, req)
	if me.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", me.Code, me.Body.String())
	}
	var claims struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
		Issuer string `json:"issuer"`
	}
	if err := json.Unmarshal(me.Body.Bytes(), &claims); err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "u-1" || claims.Role != "admin" || claims.Issuer != testIssuer {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestMeRequiresAToken(t *testing.T) {
	router := newAuthRouter(t)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}

	// Handler guards on its own too, should it ever be mounted without the middleware.
	direct := httptest.NewRecorder()
	NewAuthHandler(nil).Me(direct, httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil))
	if direct.Code != http.StatusUnauthorized {
		t.Fatalf("direct status = %d, want 401", direct.Code)
	}
}
