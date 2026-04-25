package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSanitizeInput_AllowsNormalJSON(t *testing.T) {
	handler := SanitizeInput(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	body := `{"serial_number":"VKD-001","device_type":"CAMERA","site_id":"site-1","organization_id":"org-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for normal JSON, got %d", w.Code)
	}
}

func TestSanitizeInput_BlocksNoSQLInjection(t *testing.T) {
	handler := SanitizeInput(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	tests := []struct {
		name string
		body string
	}{
		{"$where injection", `{"$where":"this.password == 'admin'"}`},
		{"$gt injection", `{"field":{"$gt":""}}`},
		{"$ne injection", `{"field":{"$ne":null}}`},
		{"$regex injection", `{"field":{"$regex":".*"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestSanitizeInput_BlocksXSS(t *testing.T) {
	handler := SanitizeInput(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	tests := []struct {
		name string
		body string
	}{
		{"script tag", `<script>alert('xss')</script>`},
		{"javascript URI", `javascript:alert(1)`},
		{"event handler", `onload=alert(1)`},
		{"data URI", `data:text/html,<script>alert(1)</script>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := `{"name":"` + tt.body + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/v1/devices", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for XSS %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestSanitizeInput_BlocksPathTraversal(t *testing.T) {
	handler := SanitizeInput(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	}))

	tests := []struct {
		name string
		path string
	}{
		{"double dot slash", "/api/v1/../../etc/passwd"},
		{"double dot backslash", "/api/v1/..\\windows\\system32"},
		{"encoded traversal", "/api/v1/%2e%2e/etc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400 for path traversal %s, got %d", tt.name, w.Code)
			}
		})
	}
}

func TestSanitizeInput_AllowsGetWithoutBody(t *testing.T) {
	handler := SanitizeInput(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/devices", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for GET, got %d", w.Code)
	}
}

func TestBruteForceProtection_AllowsUnderLimit(t *testing.T) {
	bf := NewBruteForceProtection(3, time.Minute, 5*time.Minute)
	handler := bf.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "1.2.3.4:1234"
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, w.Code)
		}
	}
}

func TestBruteForceProtection_BansAfterMaxFails(t *testing.T) {
	bf := NewBruteForceProtection(3, time.Minute, 5*time.Minute)
	handler := bf.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "1.2.3.4:1234"
	for i := 0; i < 3; i++ {
		bf.RecordFailure(ip)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = ip
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 after max failures, got %d", w.Code)
	}
}

func TestBruteForceProtection_SuccessResets(t *testing.T) {
	bf := NewBruteForceProtection(3, time.Minute, 5*time.Minute)
	handler := bf.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "1.2.3.4:1234"
	bf.RecordFailure(ip)
	bf.RecordFailure(ip)
	bf.RecordSuccess(ip) // reset

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = ip
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after success reset, got %d", w.Code)
	}
}

func TestBruteForceProtection_DifferentIPsIndependent(t *testing.T) {
	bf := NewBruteForceProtection(2, time.Minute, 5*time.Minute)
	handler := bf.Protect(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	bf.RecordFailure("1.1.1.1:1111")
	bf.RecordFailure("1.1.1.1:1111")

	// IP 1 should be banned
	req1 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req1.RemoteAddr = "1.1.1.1:1111"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusTooManyRequests {
		t.Errorf("IP 1 should be banned, got %d", w1.Code)
	}

	// IP 2 should still work
	req2 := httptest.NewRequest(http.MethodPost, "/login", nil)
	req2.RemoteAddr = "2.2.2.2:2222"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("IP 2 should be allowed, got %d", w2.Code)
	}
}

func TestSSRFValidation_BlocksInternalURLs(t *testing.T) {
	ssrf := NewSSRFProtection()

	tests := []struct {
		name string
		url  string
	}{
		{"localhost", "http://localhost:8080/admin"},
		{"127.0.0.1", "http://127.0.0.1/secrets"},
		{"AWS metadata", "http://169.254.169.254/latest/meta-data/"},
		{"GCP metadata", "http://metadata.google.internal/computeMetadata/v1/"},
		{"file scheme", "file:///etc/passwd"},
		{"ftp scheme", "ftp://internal.server/data"},
		{"dynamodb", "http://dynamodb:8000/admin"},
		{"s3", "http://s3.amazonaws.com/bucket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ssrf.ValidateURL(tt.url)
			if err == nil {
				t.Errorf("expected SSRF block for %s", tt.url)
			}
		})
	}
}

func TestSSRFValidation_AllowsExternalURLs(t *testing.T) {
	ssrf := NewSSRFProtection()

	urls := []string{
		"https://api.example.com/webhook",
		"https://maps.googleapis.com/geocode",
		"https://cdn.example.com/firmware/v2.bin",
	}

	for _, url := range urls {
		err := ssrf.ValidateURL(url)
		if err != nil {
			t.Errorf("expected %s to be allowed, got: %v", url, err)
		}
	}
}

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":   "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
		"Permissions-Policy":       "camera=(), microphone=(), geolocation=(self)",
		"Cache-Control":            "no-store",
	}

	for header, expected := range expectedHeaders {
		got := w.Header().Get(header)
		if got != expected {
			t.Errorf("header %s: expected %q, got %q", header, expected, got)
		}
	}

	csp := w.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'self'") {
		t.Errorf("CSP header missing default-src: %s", csp)
	}
}

func TestTenantIsolation_SetsHeaders(t *testing.T) {
	handler := TenantIsolation(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Organization-ID") != "org-42" {
			t.Error("expected X-Organization-ID header")
		}
		if r.Header.Get("X-User-ID") != "user-1" {
			t.Error("expected X-User-ID header")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := context.WithValue(req.Context(), claimsKey{}, &Claims{UserID: "user-1", Role: "admin", OrgID: "org-42"})
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOWASPHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/owasp", nil)
	w := httptest.NewRecorder()
	OWASPHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result["standard"] != "OWASP Top 10 (2021)" {
		t.Errorf("unexpected standard: %v", result["standard"])
	}
	controls := result["controls"].([]interface{})
	if len(controls) != 10 {
		t.Errorf("expected 10 OWASP controls, got %d", len(controls))
	}
}
