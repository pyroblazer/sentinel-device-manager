// Package handler implements the HTTP and gRPC transport layer.
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sentinel-device-manager/backend/go/internal/middleware"
	"github.com/sentinel-device-manager/backend/go/internal/service"
)

// AuthHandler exposes the login and session endpoints.
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// RegisterRoutes mounts the public login route. Rate limiting, brute-force
// protection and audit logging apply through the router-level middleware.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/auth/login", h.Login)
}

// RegisterProtectedRoutes mounts endpoints that require a valid token.
func (h *AuthHandler) RegisterProtectedRoutes(r chi.Router) {
	r.Get("/api/v1/auth/me", h.Me)
}

// Login handles POST /api/v1/auth/login {"username","password"}.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, user, err := h.auth.Login(r.Context(), input.Username, input.Password)
	if err != nil {
		// Generic message: do not reveal whether the account exists
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user": map[string]any{
			"user_id":  user.UserID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// Me handles GET /api/v1/auth/me and echoes the authenticated identity.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r)
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id":  claims.UserID,
		"role":     claims.Role,
		"org_id":   claims.OrgID,
		"scopes":   claims.Scopes,
		"issuer":   claims.Issuer,
		"exp":      claims.Exp,
	})
}
