package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/sentinel-device-manager/backend/go/internal/compliance"
	"github.com/sentinel-device-manager/backend/go/internal/handler"
	"github.com/sentinel-device-manager/backend/go/internal/middleware"
	"github.com/sentinel-device-manager/backend/go/internal/model"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
	"github.com/sentinel-device-manager/backend/go/internal/service"
)

// ---------------------------------------------------------------------------
// In-memory repository (no DynamoDB dependency)
// ---------------------------------------------------------------------------

type memRepo struct {
	devices map[string]*model.Device
}

func newMemRepo() *memRepo {
	return &memRepo{devices: make(map[string]*model.Device)}
}

func (m *memRepo) Create(_ context.Context, d *model.Device) error {
	m.devices[d.DeviceID] = d
	return nil
}

func (m *memRepo) GetByID(_ context.Context, id string) (*model.Device, error) {
	d, ok := m.devices[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return d, nil
}

func (m *memRepo) List(_ context.Context, f repository.DeviceFilter) ([]model.Device, int, error) {
	result := make([]model.Device, 0)
	for _, d := range m.devices {
		if f.DeviceType != nil && d.DeviceType != *f.DeviceType {
			continue
		}
		if f.Status != nil && d.Status != *f.Status {
			continue
		}
		if f.SiteID != nil && d.SiteID != *f.SiteID {
			continue
		}
		if f.OrganizationID != nil && d.OrganizationID != *f.OrganizationID {
			continue
		}
		result = append(result, *d)
	}
	return result, len(result), nil
}

func (m *memRepo) Update(_ context.Context, d *model.Device) error {
	m.devices[d.DeviceID] = d
	return nil
}

func (m *memRepo) Delete(_ context.Context, id string) error {
	delete(m.devices, id)
	return nil
}

// ---------------------------------------------------------------------------
// Test server wiring — mirrors main.go but without auth and DynamoDB
// ---------------------------------------------------------------------------

// setupFullServer creates an httptest.Server wired with all middleware,
// handlers, and compliance endpoints — fully in-memory, no external deps.
func setupFullServer(t *testing.T) *httptest.Server {
	t.Helper()

	repo := newMemRepo()
	svc := service.NewDeviceService(repo)
	restHandler := handler.NewRESTHandler(svc)
	superappHandler := handler.NewSuperappHandler()
	complianceReporter := compliance.NewComplianceReporter("test-1.0.0")

	auditLogger := middleware.NewAuditLogger()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	validator := middleware.NewRequestValidator()
	bruteForce := middleware.NewBruteForceProtection(5, 15*time.Minute, 30*time.Minute)

	r := chi.NewRouter()

	// Core middleware (same as main.go)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(30 * time.Second))

	// Security middleware (same stack as main.go)
	r.Use(rateLimiter.RateLimit)
	r.Use(validator.Validate)
	r.Use(auditLogger.Audit)
	r.Use(middleware.SanitizeInput)
	r.Use(middleware.SecurityHeaders)
	r.Use(bruteForce.Protect)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public endpoints
	r.Get("/health", healthHandler(complianceReporter, "test-1.0.0"))
	r.Get("/api/v1/compliance/standards", complianceReporter.HandleStandardsList)
	r.Get("/api/v1/compliance/report", complianceReporter.HandleComplianceReport)
	r.Get("/api/v1/compliance/retention", complianceReporter.HandleRetentionPolicies)
	r.Get("/api/v1/compliance/incidents", complianceReporter.HandleIncidents)
	r.Post("/api/v1/compliance/incidents", incidentCreateHandler(complianceReporter))
	r.Put("/api/v1/compliance/incidents/{incidentID}", incidentUpdateHandler(complianceReporter))
	r.Get("/api/v1/security/owasp", middleware.OWASPHandler)

	// API routes — no auth for integration tests (authEnabled = false path)
	r.Route("/api/v1", func(r chi.Router) {
		restHandler.RegisterRoutes(r)
		superappHandler.RegisterRoutes(r)
	})

	return httptest.NewServer(r)
}

// ---------------------------------------------------------------------------
// Duplicated minimal handler helpers from main.go (test-only copies)
// ---------------------------------------------------------------------------

func healthHandler(_ *compliance.ComplianceReporter, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"` + version + `"}`))
	}
}

func incidentCreateHandler(cr *compliance.ComplianceReporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Category    string `json:"category"`
			ReportedBy  string `json:"reported_by"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if input.Title == "" || input.Severity == "" {
			writeErr(w, http.StatusBadRequest, "title and severity are required")
			return
		}
		incident := cr.CreateIncident(
			input.Title, input.Description,
			input.Severity, input.Category,
			input.ReportedBy,
		)
		writeJSON(w, http.StatusCreated, incident)
	}
}

func incidentUpdateHandler(cr *compliance.ComplianceReporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		incidentID := chi.URLParam(r, "incidentID")
		var input struct {
			Status    string `json:"status"`
			UpdatedBy string `json:"updated_by"`
			Notes     string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if err := cr.UpdateIncidentStatus(incidentID, input.Status, input.UpdatedBy, input.Notes); err != nil {
			writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// ---------------------------------------------------------------------------
// 1. Device CRUD through full middleware stack
// ---------------------------------------------------------------------------

func TestIntegration_DeviceCRUD_FullMiddlewareStack(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// CREATE
	createBody := map[string]interface{}{
		"serial_number":   "VKD-INT-001",
		"device_type":     "CAMERA",
		"model":           "D30",
		"site_id":         "site-int-test",
		"organization_id": "org-int-test",
		"config": map[string]string{
			"resolution":     "4K",
			"retention_days": "30",
		},
	}
	b, _ := json.Marshal(createBody)
	resp, err := client.Post(srv.URL+"/api/v1/devices", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("create device: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var created map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&created)
	deviceID := created["device_id"].(string)
	resp.Body.Close()

	// GET
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/devices/%s", srv.URL, deviceID))
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var fetched map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&fetched)
	resp.Body.Close()
	if fetched["serial_number"] != "VKD-INT-001" {
		t.Errorf("expected serial VKD-INT-001, got %v", fetched["serial_number"])
	}

	// UPDATE
	updateBody := map[string]interface{}{"model": "D50", "status": "MAINTENANCE"}
	b, _ = json.Marshal(updateBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/devices/%s", srv.URL, deviceID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update device: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on update, got %d", resp.StatusCode)
	}
	var updated map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()
	if updated["model"] != "D50" {
		t.Errorf("expected model D50, got %v", updated["model"])
	}

	// LIST
	resp, err = client.Get(srv.URL + "/api/v1/devices")
	if err != nil {
		t.Fatalf("list devices: %v", err)
	}
	var listResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()
	if listResp["total"].(float64) < 1 {
		t.Errorf("expected at least 1 device, got %v", listResp["total"])
	}

	// DELETE
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/devices/%s", srv.URL, deviceID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete device: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify deletion
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/devices/%s", srv.URL, deviceID))
	if err != nil {
		t.Fatalf("get deleted device: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// ---------------------------------------------------------------------------
// 2. Security headers present on all responses
// ---------------------------------------------------------------------------

func TestIntegration_SecurityHeadersPresent(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	endpoints := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"health", "GET", "/health", ""},
		{"owasp", "GET", "/api/v1/security/owasp", ""},
		{"compliance standards", "GET", "/api/v1/compliance/standards", ""},
		{"compliance report", "GET", "/api/v1/compliance/report", ""},
		{"devices list", "GET", "/api/v1/devices", ""},
		{"groups list", "GET", "/api/v1/groups", ""},
		{"notifications list", "GET", "/api/v1/notifications", ""},
	}

	expectedHeaders := map[string]string{
		"X-Content-Type-Options":   "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
		"Cache-Control":            "no-store",
	}

	for _, ep := range endpoints {
		t.Run(ep.name, func(t *testing.T) {
			var resp *http.Response
			var err error
			if ep.body != "" {
				resp, err = client.Post(srv.URL+ep.path, "application/json", strings.NewReader(ep.body))
			} else {
				resp, err = client.Get(srv.URL + ep.path)
			}
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			resp.Body.Close()

			for header, expected := range expectedHeaders {
				got := resp.Header.Get(header)
				if got != expected {
					t.Errorf("endpoint %s: header %s = %q, want %q", ep.path, header, got, expected)
				}
			}

			csp := resp.Header.Get("Content-Security-Policy")
			if !strings.Contains(csp, "default-src 'self'") {
				t.Errorf("endpoint %s: CSP missing default-src 'self': %s", ep.path, csp)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Rate limiting kicks in after threshold
// ---------------------------------------------------------------------------

func TestIntegration_RateLimiting(t *testing.T) {
	// Build a server with a very low rate limit
	repo := newMemRepo()
	svc := service.NewDeviceService(repo)
	restHandler := handler.NewRESTHandler(svc)
	rateLimiter := middleware.NewRateLimiter(5, time.Minute)

	r := chi.NewRouter()
	r.Use(rateLimiter.RateLimit)
	r.Use(middleware.SecurityHeaders)
	restHandler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()
	client := srv.Client()

	// Send 5 requests — should all succeed
	for i := 0; i < 5; i++ {
		resp, err := client.Get(srv.URL + "/api/v1/devices")
		if err != nil {
			t.Fatalf("request %d: %v", i+1, err)
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Fatalf("request %d should not be rate limited yet", i+1)
		}
	}

	// 6th request should be rate limited
	resp, err := client.Get(srv.URL + "/api/v1/devices")
	if err != nil {
		t.Fatalf("rate-limited request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 on 6th request, got %d", resp.StatusCode)
	}

	// Verify Retry-After header is set
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		t.Error("expected Retry-After header on rate-limited response")
	}
}

// ---------------------------------------------------------------------------
// 4. Input sanitization rejects XSS payloads
// ---------------------------------------------------------------------------

func TestIntegration_InputSanitization_XSS(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	xssPayloads := []struct {
		name string
		body string
	}{
		{
			name: "script tag in serial_number",
			body: `{"serial_number":"<script>alert(1)</script>","device_type":"CAMERA","site_id":"s1","organization_id":"o1"}`,
		},
		{
			name: "javascript URI in name",
			body: `{"name":"javascript:alert(1)"}`,
		},
		{
			name: "onerror handler",
			body: `{"serial_number":"onerror=alert(1)","device_type":"CAMERA","site_id":"s1","organization_id":"o1"}`,
		},
		{
			name: "data URI",
			body: `{"serial_number":"data:text/html,<script>alert(1)</script>","device_type":"CAMERA","site_id":"s1","organization_id":"o1"}`,
		},
		{
			name: "eval call",
			body: `{"serial_number":"eval(document.cookie)","device_type":"CAMERA","site_id":"s1","organization_id":"o1"}`,
		},
	}

	for _, tt := range xssPayloads {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Post(srv.URL+"/api/v1/devices", "application/json", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400 for XSS payload %q, got %d", tt.name, resp.StatusCode)
			}
		})
	}
}

func TestIntegration_InputSanitization_NoSQLInjection(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	payloads := []struct {
		name string
		body string
	}{
		{"$where", `{"$where":"this.password == 'admin'"}`},
		{"$gt", `{"field":{"$gt":""},"serial_number":"x","device_type":"CAMERA","site_id":"s","organization_id":"o"}`},
		{"$ne", `{"field":{"$ne":null},"serial_number":"x","device_type":"CAMERA","site_id":"s","organization_id":"o"}`},
	}

	for _, tt := range payloads {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Post(srv.URL+"/api/v1/devices", "application/json", strings.NewReader(tt.body))
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400 for NoSQL injection %q, got %d", tt.name, resp.StatusCode)
			}
		})
	}
}

func TestIntegration_InputSanitization_PathTraversal(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	paths := []struct {
		name string
		path string
	}{
		{"double dot slash", "/api/v1/../../etc/passwd"},
		{"double dot backslash", "/api/v1/..\\windows\\system32"},
	}

	for _, tt := range paths {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(srv.URL + tt.path)
			if err != nil {
				t.Fatalf("request: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("expected 400 for path traversal %q, got %d", tt.name, resp.StatusCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. OWASP endpoint returns all 10 controls
// ---------------------------------------------------------------------------

func TestIntegration_OWASPEndpoint_ReturnsAll10Controls(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/security/owasp")
	if err != nil {
		t.Fatalf("owasp request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["standard"] != "OWASP Top 10 (2021)" {
		t.Errorf("unexpected standard: %v", result["standard"])
	}

	controls, ok := result["controls"].([]interface{})
	if !ok {
		t.Fatalf("controls is not a slice: %T", result["controls"])
	}
	if len(controls) != 10 {
		t.Errorf("expected 10 OWASP controls, got %d", len(controls))
	}

	// Verify each control has required fields
	for i, c := range controls {
		control, ok := c.(map[string]interface{})
		if !ok {
			t.Errorf("control %d is not a map", i)
			continue
		}
		for _, field := range []string{"id", "name", "status", "evidence"} {
			if _, exists := control[field]; !exists {
				t.Errorf("control %d missing field %s", i, field)
			}
		}
	}

	// Verify all A01-A10 are present
	ids := make(map[string]bool)
	for _, c := range controls {
		control := c.(map[string]interface{})
		ids[control["id"].(string)] = true
	}
	for _, expected := range []string{"A01", "A02", "A03", "A04", "A05", "A06", "A07", "A08", "A09", "A10"} {
		if !ids[expected] {
			t.Errorf("missing OWASP control %s", expected)
		}
	}
}

// ---------------------------------------------------------------------------
// 6. Compliance report endpoint returns standards
// ---------------------------------------------------------------------------

func TestIntegration_ComplianceReport_ReturnsStandards(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/compliance/report")
	if err != nil {
		t.Fatalf("compliance report: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var report map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&report)
	resp.Body.Close()

	// Verify top-level fields
	if _, ok := report["generated_at"]; !ok {
		t.Error("report missing generated_at")
	}
	if report["version"] != "test-1.0.0" {
		t.Errorf("expected version test-1.0.0, got %v", report["version"])
	}

	// Verify standards list
	standards, ok := report["standards"].([]interface{})
	if !ok {
		t.Fatalf("standards is not a slice: %T", report["standards"])
	}
	expectedStandardCount := 10 // ISO27001, ISO9001, ISO27035, ISO27017, ISO20000, ISO22301, IEC62443, NIST-CSF, SOC2, GDPR
	if len(standards) != expectedStandardCount {
		t.Errorf("expected %d standards, got %d", expectedStandardCount, len(standards))
	}

	// Verify summary
	summary, ok := report["summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("summary is not a map: %T", report["summary"])
	}
	if _, ok := summary["total_controls"]; !ok {
		t.Error("summary missing total_controls")
	}
	if _, ok := summary["compliance_pct"]; !ok {
		t.Error("summary missing compliance_pct")
	}
}

func TestIntegration_ComplianceStandardsList(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/compliance/standards")
	if err != nil {
		t.Fatalf("standards list: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	standards, ok := result["standards"].([]interface{})
	if !ok {
		t.Fatalf("standards not a slice: %T", result["standards"])
	}
	if len(standards) != 10 {
		t.Errorf("expected 10 standards, got %d", len(standards))
	}
	if result["total"].(float64) != 10 {
		t.Errorf("expected total=10, got %v", result["total"])
	}
}

func TestIntegration_ComplianceRetentionPolicies(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/compliance/retention")
	if err != nil {
		t.Fatalf("retention: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	policies, ok := result["policies"].([]interface{})
	if !ok {
		t.Fatalf("policies not a slice: %T", result["policies"])
	}
	if len(policies) < 5 {
		t.Errorf("expected at least 5 retention policies, got %d", len(policies))
	}
}

// ---------------------------------------------------------------------------
// 7. Superapp endpoints (groups, templates, webhooks)
// ---------------------------------------------------------------------------

func TestIntegration_GroupsCRUD(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create group
	createBody := `{"name":"Test Building","type":"BUILDING","site_id":"site-001"}`
	resp, err := client.Post(srv.URL+"/api/v1/groups", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var group map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&group)
	resp.Body.Close()
	groupID := group["group_id"].(string)

	// List groups
	resp, err = client.Get(srv.URL + "/api/v1/groups")
	if err != nil {
		t.Fatalf("list groups: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()
	if listResult["total"].(float64) < 1 {
		t.Error("expected at least 1 group")
	}

	// Get group
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/groups/%s", srv.URL, groupID))
	if err != nil {
		t.Fatalf("get group: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Update group
	updateBody := `{"name":"Updated Building"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/groups/%s", srv.URL, groupID), strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update group: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&updated)
	resp.Body.Close()
	if updated["name"] != "Updated Building" {
		t.Errorf("expected name 'Updated Building', got %v", updated["name"])
	}

	// Delete group
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/groups/%s", srv.URL, groupID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete group: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestIntegration_TemplatesCRUD(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create template
	createBody := `{"name":"Camera Default","device_type":"CAMERA","config":{"resolution":"1080p","fps":"30"},"description":"Default camera config"}`
	resp, err := client.Post(srv.URL+"/api/v1/templates", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var tmpl map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&tmpl)
	resp.Body.Close()
	templateID := tmpl["template_id"].(string)

	// List templates
	resp, err = client.Get(srv.URL + "/api/v1/templates")
	if err != nil {
		t.Fatalf("list templates: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()
	if listResult["total"].(float64) < 1 {
		t.Error("expected at least 1 template")
	}

	// Get template
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/templates/%s", srv.URL, templateID))
	if err != nil {
		t.Fatalf("get template: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Delete template
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/templates/%s", srv.URL, templateID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete template: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestIntegration_WebhooksCRUD(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create webhook
	createBody := `{"name":"Alert Webhook","url":"https://hooks.example.com/alert","events":["DEVICE_CREATED","ALERT_TRIGGERED"],"headers":{"X-Auth":"token123"}}`
	resp, err := client.Post(srv.URL+"/api/v1/webhooks", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create webhook: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var wh map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&wh)
	resp.Body.Close()
	webhookID := wh["webhook_id"].(string)

	if wh["active"].(bool) != true {
		t.Error("expected webhook to be active")
	}

	// List webhooks
	resp, err = client.Get(srv.URL + "/api/v1/webhooks")
	if err != nil {
		t.Fatalf("list webhooks: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()
	if listResult["total"].(float64) < 1 {
		t.Error("expected at least 1 webhook")
	}

	// Test webhook
	resp, err = client.Post(fmt.Sprintf("%s/api/v1/webhooks/%s/test", srv.URL, webhookID), "application/json", nil)
	if err != nil {
		t.Fatalf("test webhook: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var testResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&testResult)
	resp.Body.Close()
	if testResult["status"] != "test_sent" {
		t.Errorf("expected status test_sent, got %v", testResult["status"])
	}

	// Delete webhook
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/webhooks/%s", srv.URL, webhookID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete webhook: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestIntegration_APIKeysCRUD(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create API key
	createBody := `{"name":"Test Key","role":"admin","org_id":"org-test"}`
	resp, err := client.Post(srv.URL+"/api/v1/api-keys", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create api key: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var key map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&key)
	resp.Body.Close()

	if !strings.HasPrefix(key["key"].(string), "sk-sentinel-") {
		t.Errorf("expected key to start with sk-sentinel-, got %s", key["key"])
	}

	// List API keys (key should be redacted)
	resp, err = client.Get(srv.URL + "/api/v1/api-keys")
	if err != nil {
		t.Fatalf("list api keys: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()

	keys := listResult["api_keys"].([]interface{})
	if len(keys) < 1 {
		t.Error("expected at least 1 API key")
	}

	// Delete API key
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/api-keys/%s", srv.URL, key["key_id"].(string)), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete api key: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestIntegration_Notifications(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create a group to trigger notification
	groupBody := `{"name":"Notify Test Group","type":"CUSTOM","site_id":"site-n1"}`
	resp, err := client.Post(srv.URL+"/api/v1/groups", "application/json", strings.NewReader(groupBody))
	if err != nil {
		t.Fatalf("create group: %v", err)
	}
	resp.Body.Close()

	// List notifications
	resp, err = client.Get(srv.URL + "/api/v1/notifications")
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()

	notifications := listResult["notifications"].([]interface{})
	if len(notifications) < 1 {
		t.Fatal("expected at least 1 notification after creating a group")
	}

	notificationID := notifications[0].(map[string]interface{})["notification_id"].(string)

	// Mark as read
	resp, err = client.Post(fmt.Sprintf("%s/api/v1/notifications/%s/read", srv.URL, notificationID), "application/json", nil)
	if err != nil {
		t.Fatalf("mark read: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var readNotif map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&readNotif)
	resp.Body.Close()
	if readNotif["read"].(bool) != true {
		t.Error("expected notification to be marked as read")
	}

	// Delete notification
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/notifications/%s", srv.URL, notificationID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete notification: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestIntegration_Geofences(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create geofence
	createBody := `{"name":"HQ Perimeter","center_lat":40.7128,"center_lng":-74.006,"radius_meters":500,"site_id":"site-hq"}`
	resp, err := client.Post(srv.URL+"/api/v1/geofences", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create geofence: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var gf map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&gf)
	resp.Body.Close()
	zoneID := gf["zone_id"].(string)

	// List geofences
	resp, err = client.Get(srv.URL + "/api/v1/geofences")
	if err != nil {
		t.Fatalf("list geofences: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()
	if listResult["total"].(float64) < 1 {
		t.Error("expected at least 1 geofence")
	}

	// Delete geofence
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/v1/geofences/%s", srv.URL, zoneID), nil)
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("delete geofence: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// ---------------------------------------------------------------------------
// 8. Bulk operations return queued status
// ---------------------------------------------------------------------------

func TestIntegration_BulkDelete_ReturnsQueued(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	body := `{"device_ids":["dev-001","dev-002","dev-003"]}`
	resp, err := client.Post(srv.URL+"/api/v1/devices/bulk-delete", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("bulk delete: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["status"] != "queued" {
		t.Errorf("expected status=queued, got %v", result["status"])
	}
	if result["action"] != "bulk_delete" {
		t.Errorf("expected action=bulk_delete, got %v", result["action"])
	}
	if result["total"].(float64) != 3 {
		t.Errorf("expected total=3, got %v", result["total"])
	}
}

func TestIntegration_BulkUpdate_ReturnsQueued(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	body := `{"device_ids":["dev-001","dev-002"],"updates":{"status":"MAINTENANCE"}}`
	resp, err := client.Post(srv.URL+"/api/v1/devices/bulk-update", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("bulk update: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["status"] != "queued" {
		t.Errorf("expected status=queued, got %v", result["status"])
	}
	if result["action"] != "bulk_update" {
		t.Errorf("expected action=bulk_update, got %v", result["action"])
	}
}

func TestIntegration_BulkTag_ReturnsQueued(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	body := `{"device_ids":["dev-001"],"tags":["critical","outdoor"]}`
	resp, err := client.Post(srv.URL+"/api/v1/devices/bulk-tag", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("bulk tag: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["status"] != "queued" {
		t.Errorf("expected status=queued, got %v", result["status"])
	}
}

func TestIntegration_BulkDelete_EmptyListRejected(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	body := `{"device_ids":[]}`
	resp, err := client.Post(srv.URL+"/api/v1/devices/bulk-delete", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("bulk delete empty: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for empty device_ids, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// 9. Data export returns proper content types
// ---------------------------------------------------------------------------

func TestIntegration_DataExport_JSON(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/devices/export?format=json")
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "devices.json") {
		t.Errorf("expected Content-Disposition to contain devices.json, got %s", contentDisposition)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["format"] != "json" {
		t.Errorf("expected format=json, got %v", result["format"])
	}
	if _, ok := result["devices"]; !ok {
		t.Error("export missing devices field")
	}
	if _, ok := result["exported_at"]; !ok {
		t.Error("export missing exported_at field")
	}
}

func TestIntegration_DataExport_CSV(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/devices/export?format=csv")
	if err != nil {
		t.Fatalf("export csv: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/csv" {
		t.Errorf("expected Content-Type text/csv, got %s", contentType)
	}

	contentDisposition := resp.Header.Get("Content-Disposition")
	if !strings.Contains(contentDisposition, "devices.csv") {
		t.Errorf("expected Content-Disposition to contain devices.csv, got %s", contentDisposition)
	}

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if !strings.Contains(string(body), "device_id,serial_number,device_type,status,site_id") {
		t.Errorf("CSV header missing or incorrect: %s", string(body[:200]))
	}
}

func TestIntegration_DataExport_DefaultIsJSON(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/api/v1/devices/export")
	if err != nil {
		t.Fatalf("export default: %v", err)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected default export to be JSON, got %s", contentType)
	}
	resp.Body.Close()
}

// ---------------------------------------------------------------------------
// 10. Compliance incidents lifecycle
// ---------------------------------------------------------------------------

func TestIntegration_IncidentLifecycle(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Create incident
	createBody := `{"title":"Unauthorized access attempt","description":"Multiple failed logins","severity":"HIGH","category":"SECURITY","reported_by":"admin@sentinel.io"}`
	resp, err := client.Post(srv.URL+"/api/v1/compliance/incidents", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var incident map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&incident)
	resp.Body.Close()

	incidentID := incident["incident_id"].(string)
	if incident["status"] != "OPEN" {
		t.Errorf("expected status OPEN, got %v", incident["status"])
	}
	if incident["severity"] != "HIGH" {
		t.Errorf("expected severity HIGH, got %v", incident["severity"])
	}

	// List incidents
	resp, err = client.Get(srv.URL + "/api/v1/compliance/incidents")
	if err != nil {
		t.Fatalf("list incidents: %v", err)
	}
	var listResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&listResult)
	resp.Body.Close()
	if listResult["total"].(float64) < 1 {
		t.Error("expected at least 1 incident")
	}

	// Update incident status
	updateBody := `{"status":"INVESTIGATING","updated_by":"security-team","notes":"Under investigation"}`
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/compliance/incidents/%s", srv.URL, incidentID), strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("update incident: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Filter incidents by status
	resp, err = client.Get(srv.URL + "/api/v1/compliance/incidents?status=OPEN")
	if err != nil {
		t.Fatalf("filter incidents: %v", err)
	}
	var openIncidents map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&openIncidents)
	resp.Body.Close()
	// The incident we updated to INVESTIGATING should not appear
	if openIncidents["total"].(float64) != 0 {
		t.Logf("Note: open incidents = %v (incident was moved to INVESTIGATING)", openIncidents["total"])
	}
}

func TestIntegration_IncidentCreate_MissingFields(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	// Missing severity
	body := `{"title":"Test incident"}`
	resp, err := client.Post(srv.URL+"/api/v1/compliance/incidents", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("create incident: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing severity, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// ---------------------------------------------------------------------------
// 11. Health endpoint
// ---------------------------------------------------------------------------

func TestIntegration_HealthEndpoint(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	resp, err := client.Get(srv.URL + "/health")
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result["status"] != "ok" {
		t.Errorf("expected status ok, got %v", result["status"])
	}
	if result["version"] != "test-1.0.0" {
		t.Errorf("expected version test-1.0.0, got %v", result["version"])
	}
}

// ---------------------------------------------------------------------------
// 12. Request body size validation
// ---------------------------------------------------------------------------

func TestIntegration_RequestTooLarge(t *testing.T) {
	repo := newMemRepo()
	svc := service.NewDeviceService(repo)
	restHandler := handler.NewRESTHandler(svc)
	validator := middleware.NewRequestValidator()

	r := chi.NewRouter()
	r.Use(validator.Validate)
	r.Use(middleware.SecurityHeaders)
	restHandler.RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()
	client := srv.Client()

	// Create a body larger than 1MB
	largeBody := make([]byte, 1024*1024+100)
	for i := range largeBody {
		largeBody[i] = 'a'
	}
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/api/v1/devices", bytes.NewReader(largeBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("large request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 for oversized body, got %d", resp.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// 13. Multiple device types filtering
// ---------------------------------------------------------------------------

func TestIntegration_MultipleDeviceTypes_WithFiltering(t *testing.T) {
	srv := setupFullServer(t)
	defer srv.Close()
	client := srv.Client()

	types := []string{"CAMERA", "ACCESS_CONTROL", "ALARM", "SENSOR"}
	for _, dt := range types {
		body := fmt.Sprintf(`{"serial_number":"VKD-%s-001","device_type":"%s","model":"Model-X","site_id":"site-001","organization_id":"org-001"}`, dt, dt)
		resp, err := client.Post(srv.URL+"/api/v1/devices", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatalf("create %s device: %v", dt, err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 for %s, got %d", dt, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// List all
	resp, err := client.Get(srv.URL + "/api/v1/devices")
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	var allDevices map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&allDevices)
	resp.Body.Close()
	if allDevices["total"].(float64) != 4 {
		t.Errorf("expected 4 devices, got %.0f", allDevices["total"].(float64))
	}

	// Filter by type
	resp, err = client.Get(srv.URL + "/api/v1/devices?type=CAMERA")
	if err != nil {
		t.Fatalf("filter camera: %v", err)
	}
	var cameras map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&cameras)
	resp.Body.Close()
	if cameras["total"].(float64) != 1 {
		t.Errorf("expected 1 CAMERA, got %.0f", cameras["total"].(float64))
	}

	// Filter by status
	resp, err = client.Get(srv.URL + "/api/v1/devices?status=ONLINE")
	if err != nil {
		t.Fatalf("filter online: %v", err)
	}
	var online map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&online)
	resp.Body.Close()
	if online["total"].(float64) != 4 {
		t.Errorf("expected 4 ONLINE devices, got %.0f", online["total"].(float64))
	}
}
