package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// generateTestToken creates a valid HMAC-SHA256 signed JWT for testing.
func generateTestToken(claims Claims, secret string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, _ := json.Marshal(claims)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)

	signingInput := header + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return header + "." + payloadB64 + "." + sig
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	config := JWTConfig{SecretKey: "test-secret"}
	handler := JWTAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	config := JWTConfig{SecretKey: "test-secret"}
	handler := JWTAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_WrongSecret(t *testing.T) {
	claims := Claims{
		UserID: "user-1",
		Role:   "admin",
		Exp:    time.Now().Add(1 * time.Hour).Unix(),
	}
	token := generateTestToken(claims, "wrong-secret")

	config := JWTConfig{SecretKey: "correct-secret"}
	handler := JWTAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong secret, got %d", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	config := JWTConfig{SecretKey: "test-secret"}
	claims := Claims{
		UserID: "user-1",
		Role:   "admin",
		Exp:    time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := generateTestToken(claims, "test-secret")

	handler := JWTAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

func TestJWTAuth_ValidToken_PassesContext(t *testing.T) {
	secret := "test-secret"
	config := JWTConfig{SecretKey: secret}
	claims := Claims{
		UserID: "user-1",
		Role:   "admin",
		OrgID:  "org-1",
		Scopes: []string{"devices:read", "devices:write"},
		Exp:    time.Now().Add(1 * time.Hour).Unix(),
	}
	token := generateTestToken(claims, secret)

	var extractedClaims *Claims
	handler := JWTAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedClaims = GetClaims(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if extractedClaims == nil {
		t.Fatal("expected claims in context")
	}
	if extractedClaims.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", extractedClaims.UserID)
	}
	if extractedClaims.Role != "admin" {
		t.Errorf("expected admin, got %s", extractedClaims.Role)
	}
	if extractedClaims.OrgID != "org-1" {
		t.Errorf("expected org-1, got %s", extractedClaims.OrgID)
	}
	if len(extractedClaims.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(extractedClaims.Scopes))
	}
}

func TestGetClaims_NoClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	claims := GetClaims(req)
	if claims != nil {
		t.Error("expected nil claims when none in context")
	}
}

func TestRBAC_AllowedRole(t *testing.T) {
	config := RBACConfig{
		RolePermissions: map[string]map[string][]string{
			"admin":  {"GET": {"*"}, "POST": {"*"}, "PUT": {"*"}, "DELETE": {"*"}},
			"viewer": {"GET": {"*"}},
		},
	}

	handler := RBAC(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "u1", Role: "viewer"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for viewer GET, got %d", w.Code)
	}
}

func TestRBAC_DeniedMethod(t *testing.T) {
	config := RBACConfig{
		RolePermissions: map[string]map[string][]string{
			"viewer": {"GET": {"*"}},
		},
	}

	handler := RBAC(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "u1", Role: "viewer"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for viewer POST, got %d", w.Code)
	}
}

func TestRBAC_DeniedPath(t *testing.T) {
	config := RBACConfig{
		RolePermissions: map[string]map[string][]string{
			"operator": {"GET": {"/api/v1/devices"}, "POST": {"/api/v1/devices"}},
		},
	}

	handler := RBAC(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/devices/123", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "u1", Role: "operator"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for operator DELETE, got %d", w.Code)
	}
}

func TestRBAC_NoClaims(t *testing.T) {
	config := RBACConfig{
		RolePermissions: map[string]map[string][]string{
			"admin": {"GET": {"*"}},
		},
	}

	handler := RBAC(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no claims, got %d", w.Code)
	}
}

func TestRBAC_UnknownRole(t *testing.T) {
	config := RBACConfig{
		RolePermissions: map[string]map[string][]string{
			"admin": {"GET": {"*"}},
		},
	}

	handler := RBAC(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "u1", Role: "hacker"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for unknown role, got %d", w.Code)
	}
}

func TestRateLimiter_AllowsUnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_BlocksOverLimit(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < 2 && w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
		if i == 2 && w.Code != http.StatusTooManyRequests {
			t.Errorf("request 3: expected 429, got %d", w.Code)
		}
	}
}

func TestRateLimiter_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ips := []string{"1.1.1.1:1111", "2.2.2.2:2222", "3.3.3.3:3333"}
	for i, ip := range ips {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("different IP request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_XForwardedFor(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	handler := rl.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request from IP via X-Forwarded-For
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "proxy:1234"
	req.Header.Set("X-Forwarded-For", "10.0.0.1, 172.16.0.1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("first request: expected 200, got %d", w.Code)
	}

	// Second request from same forwarded IP should be blocked
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "proxy:5678"
	req2.Header.Set("X-Forwarded-For", "10.0.0.1")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("second request: expected 429, got %d", w2.Code)
	}
}

func TestRequestValidation_TooLarge(t *testing.T) {
	validator := NewRequestValidator()
	validator.MaxBodySize = 100

	handler := validator.Validate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.ContentLength = 200
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", w.Code)
	}
}

func TestRequestValidation_BadContentType(t *testing.T) {
	validator := NewRequestValidator()

	handler := validator.Validate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "text/xml")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", w.Code)
	}
}

func TestRequestValidation_JSONAllowed(t *testing.T) {
	validator := NewRequestValidator()

	handler := validator.Validate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for JSON content type, got %d", w.Code)
	}
}

func TestRequestValidation_MultipartAllowed(t *testing.T) {
	validator := NewRequestValidator()

	handler := validator.Validate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=---")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for multipart, got %d", w.Code)
	}
}

func TestRequestValidation_GetNoContentType(t *testing.T) {
	validator := NewRequestValidator()

	handler := validator.Validate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for GET without content type, got %d", w.Code)
	}
}

func TestAuditLogger_RecordsEntries(t *testing.T) {
	al := NewAuditLogger()

	handler := al.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	req.Header.Set("X-Request-ID", "req-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	entries := al.GetEntries(10)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Method != "GET" {
		t.Errorf("expected GET, got %s", entries[0].Method)
	}
	if entries[0].Path != "/api/v1/devices" {
		t.Errorf("expected /api/v1/devices, got %s", entries[0].Path)
	}
	if entries[0].RequestID != "req-123" {
		t.Errorf("expected req-123, got %s", entries[0].RequestID)
	}
	if entries[0].StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", entries[0].StatusCode)
	}
	if entries[0].IP == "" {
		t.Error("expected IP to be set")
	}
}

func TestAuditLogger_WithClaims(t *testing.T) {
	al := NewAuditLogger()

	handler := al.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "user-1", Role: "admin", OrgID: "org-1"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	entries := al.GetEntries(10)
	if entries[0].UserID != "user-1" {
		t.Errorf("expected user-1, got %s", entries[0].UserID)
	}
	if entries[0].Role != "admin" {
		t.Errorf("expected admin, got %s", entries[0].Role)
	}
	if entries[0].OrgID != "org-1" {
		t.Errorf("expected org-1, got %s", entries[0].OrgID)
	}
}

func TestAuditLogger_CapturesStatusCode(t *testing.T) {
	al := NewAuditLogger()

	handler := al.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	entries := al.GetEntries(10)
	if entries[0].StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", entries[0].StatusCode)
	}
}

func TestAuditLogger_LimitEntries(t *testing.T) {
	al := NewAuditLogger()

	handler := al.Audit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	entries := al.GetEntries(3)
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestCORSSecurity_SetsHeaders(t *testing.T) {
	handler := CORSSecurity([]string{"http://localhost:3000"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("expected X-Content-Type-Options header")
	}
	if w.Header().Get("X-Frame-Options") != "DENY" {
		t.Error("expected X-Frame-Options header")
	}
	if w.Header().Get("Strict-Transport-Security") == "" {
		t.Error("expected HSTS header")
	}
	if w.Header().Get("Content-Security-Policy") == "" {
		t.Error("expected CSP header")
	}
	if w.Header().Get("Referrer-Policy") == "" {
		t.Error("expected Referrer-Policy header")
	}
}

func TestCORSSecurity_OptionsPreflight(t *testing.T) {
	handler := CORSSecurity([]string{"*"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler for OPTIONS")
	}))

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS, got %d", w.Code)
	}
}

func TestCORSSecurity_DisallowedOrigin(t *testing.T) {
	handler := CORSSecurity([]string{"http://allowed.com"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Error("should not set ACAO header for disallowed origin")
	}
}

func TestExtractBearerToken_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	token := extractBearerToken(req)
	if token != "token123" {
		t.Errorf("expected token123, got %s", token)
	}
}

func TestExtractBearerToken_Missing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	token := extractBearerToken(req)
	if token != "" {
		t.Errorf("expected empty, got %s", token)
	}
}

func TestExtractBearerToken_WrongScheme(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	token := extractBearerToken(req)
	if token != "" {
		t.Errorf("expected empty for non-Bearer, got %s", token)
	}
}
