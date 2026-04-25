// Package middleware provides HTTP middleware for the Sentinel Device Manager
// including JWT authentication, RBAC authorization, audit logging, rate limiting,
// and request validation.
//
// Security Controls:
//   - ISO 27001 A.9: Access Control (JWT + RBAC)
//   - ISO 27001 A.12.4: Logging and Monitoring (Audit Trail)
//   - NIST CSF PR.AC: Access Control
//   - SOC 2 CC6.1: Logical Access Security
//   - IEC 62443 SR 2.1: Authorization Enforcement
package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Claims represents the parsed JWT payload.
type Claims struct {
	UserID   string   `json:"sub"`
	Role     string   `json:"role"`
	OrgID    string   `json:"org_id"`
	Scopes   []string `json:"scopes"`
	Exp      int64    `json:"exp"`
	IssuedAt int64    `json:"iat"`
}

// JWTConfig holds JWT validation parameters.
type JWTConfig struct {
	SecretKey string
	Issuer    string
}

// JWTAuth returns middleware that validates JWT tokens from the Authorization header.
// Tokens must be HMAC-SHA256 signed (HS256).
func JWTAuth(config JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractBearerToken(r)
			if tokenStr == "" {
				writeUnauthorized(w, "missing or malformed authorization header")
				return
			}

			claims, err := parseJWT(tokenStr, config.SecretKey)
			if err != nil {
				writeUnauthorized(w, "invalid token: "+err.Error())
				return
			}

			if claims.Exp < time.Now().Unix() {
				writeUnauthorized(w, "token expired")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims extracts claims from request context.
func GetClaims(r *http.Request) *Claims {
	c, _ := r.Context().Value(claimsKey{}).(*Claims)
	return c
}

type claimsKey struct{}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}

func parseJWT(tokenStr string, secret string) (*Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	headerB64, payloadB64, sigB64 := parts[0], parts[1], parts[2]

	signingInput := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigB64), []byte(expectedSig)) {
		return nil, ErrInvalidSignature
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

// RBACConfig configures role-based access control.
type RBACConfig struct {
	// RolePermissions maps roles to allowed HTTP methods and path prefixes.
	// Example: {"admin": {"GET": ["*"], "POST": ["*"]}, "viewer": {"GET": ["*"]}}
	RolePermissions map[string]map[string][]string
}

// RBAC returns middleware that checks the user's role against required permissions.
func RBAC(config RBACConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r)
			if claims == nil {
				writeUnauthorized(w, "no authentication context")
				return
			}

			rolePerms, ok := config.RolePermissions[claims.Role]
			if !ok {
				writeForbidden(w, "role has no permissions")
				return
			}

			allowedPaths, ok := rolePerms[r.Method]
			if !ok {
				writeForbidden(w, "method not allowed for role")
				return
			}

			if !isPathAllowed(r.URL.Path, allowedPaths) {
				writeForbidden(w, "path not allowed for role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isPathAllowed(path string, patterns []string) bool {
	for _, p := range patterns {
		if p == "*" {
			return true
		}
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// AuditLogger logs every request for compliance auditing.
// Implements ISO 27001 A.12.4, SOC 2 CC7.2, ISO 27035.
type AuditLogger struct {
	mu     sync.Mutex
	entries []AuditEntry
}

// AuditEntry represents a single auditable event.
type AuditEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	StatusCode  int       `json:"status_code"`
	UserID      string    `json:"user_id,omitempty"`
	Role        string    `json:"role,omitempty"`
	OrgID       string    `json:"org_id,omitempty"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	RequestID   string    `json:"request_id"`
	Duration    string    `json:"duration"`
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{entries: make([]AuditEntry, 0, 1000)}
}

// Audit returns middleware that logs request details for audit purposes.
func (al *AuditLogger) Audit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		entry := AuditEntry{
			Timestamp:  start.UTC(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: rw.statusCode,
			IP:         r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			RequestID:  requestID,
			Duration:   time.Since(start).String(),
		}

		if claims := GetClaims(r); claims != nil {
			entry.UserID = claims.UserID
			entry.Role = claims.Role
			entry.OrgID = claims.OrgID
		}

		al.mu.Lock()
		al.entries = append(al.entries, entry)
		if len(al.entries) > 10000 {
			al.entries = al.entries[len(al.entries)-5000:]
		}
		al.mu.Unlock()
	})
}

// GetEntries returns recent audit entries.
func (al *AuditLogger) GetEntries(limit int) []AuditEntry {
	al.mu.Lock()
	defer al.mu.Unlock()

	if limit <= 0 || limit > len(al.entries) {
		limit = len(al.entries)
	}
	start := len(al.entries) - limit
	result := make([]AuditEntry, limit)
	copy(result, al.entries[start:])
	return result
}

// RateLimiter provides per-IP rate limiting using a token bucket algorithm.
// Implements ISO 27001 A.12.6 (technical vulnerability management),
// SOC 2 CC7.1, NIST CSF PR.PT.
type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     int
	window   time.Duration
	cleanup  time.Duration
}

type clientBucket struct {
	count    int
	lastSeen time.Time
}

// NewRateLimiter creates a rate limiter allowing `rate` requests per `window`.
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*clientBucket),
		rate:    rate,
		window:  window,
		cleanup: window * 2,
	}
	go rl.cleanupLoop()
	return rl
}

// RateLimit returns middleware that enforces per-IP rate limiting.
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r)

		if !rl.allow(ip) {
			w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
			writeError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	c, ok := rl.clients[ip]
	if !ok || now.Sub(c.lastSeen) > rl.window {
		rl.clients[ip] = &clientBucket{count: 1, lastSeen: now}
		return true
	}

	c.count++
	c.lastSeen = now
	return c.count <= rl.rate
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, c := range rl.clients {
			if now.Sub(c.lastSeen) > rl.cleanup {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RequestValidator provides input validation middleware.
// Implements ISO 27001 A.14.1 (security in development), NIST CSF PR.DS.
type RequestValidator struct {
	MaxBodySize int64
}

// NewRequestValidator creates a validator with sensible defaults.
func NewRequestValidator() *RequestValidator {
	return &RequestValidator{
		MaxBodySize: 1 << 20, // 1MB
	}
}

// Validate returns middleware that enforces request size and content-type checks.
func (rv *RequestValidator) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ContentLength > rv.MaxBodySize {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}

		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && !strings.HasPrefix(contentType, "application/json") &&
				!strings.HasPrefix(contentType, "multipart/form-data") {
				writeError(w, http.StatusUnsupportedMediaType, "content-type must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CORSSecurity provides production-grade CORS configuration.
// Implements ISO 27001 A.13 (communications security).
func CORSSecurity(allowedOrigins []string) func(http.Handler) http.Handler {
	originsMap := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originsMap[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if originsMap[origin] || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "300")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


func generateRequestID() string {
	b := make([]byte, 16)
	t := time.Now().UnixNano()
	for i := 0; i < 8; i++ {
		b[i] = byte(t >> (i * 8)) //nolint:gosec // intentional byte extraction from int64
	}
	for i := 8; i < 16; i++ {
		b[i] = byte(t >> ((i - 8) * 8)) //nolint:gosec // intentional byte extraction from int64
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusUnauthorized, msg)
}

func writeForbidden(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusForbidden, msg)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
