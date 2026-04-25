package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// setupSuperappRouter creates a fresh SuperappHandler and chi router for each test.
func setupSuperappRouter() (*SuperappHandler, chi.Router) {
	h := NewSuperappHandler()
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return h, r
}

// doJSON is a helper that issues a request via the router and returns the response recorder.
func doJSON(r chi.Router, method, path string, body interface{}) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func decodeJSON(w *httptest.ResponseRecorder) map[string]interface{} {
	var m map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&m)
	return m
}

// ============================================================================
// Device Groups
// ============================================================================

func TestCreateGroup_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name":    "Building A",
		"type":    "BUILDING",
		"site_id": "site-1",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Building A" {
		t.Errorf("expected name 'Building A', got %v", resp["name"])
	}
	if resp["type"] != "BUILDING" {
		t.Errorf("expected type 'BUILDING', got %v", resp["type"])
	}
	if resp["group_id"] == nil || resp["group_id"] == "" {
		t.Error("expected group_id to be set")
	}
}

func TestCreateGroup_MissingName(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"type":    "ZONE",
		"site_id": "site-1",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "name is required" {
		t.Errorf("expected 'name is required' error, got %v", resp["error"])
	}
}

func TestCreateGroup_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/groups", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListGroups_Empty(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodGet, "/api/v1/groups", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(0) {
		t.Errorf("expected total 0, got %v", resp["total"])
	}
}

func TestListGroups_AfterCreate(t *testing.T) {
	_, r := setupSuperappRouter()

	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "Group 1", "type": "ZONE", "site_id": "s1",
	})
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "Group 2", "type": "FLOOR", "site_id": "s1",
	})

	w := doJSON(r, http.MethodGet, "/api/v1/groups", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(2) {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestGetGroup_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "TestGroup", "type": "CUSTOM", "site_id": "site-x",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	w := doJSON(r, http.MethodGet, "/api/v1/groups/"+groupID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "TestGroup" {
		t.Errorf("expected name 'TestGroup', got %v", resp["name"])
	}
}

func TestGetGroup_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodGet, "/api/v1/groups/nonexistent", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "group not found" {
		t.Errorf("expected 'group not found', got %v", resp["error"])
	}
}

func TestUpdateGroup_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "Old Name", "type": "ZONE", "site_id": "s1",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	w := doJSON(r, http.MethodPut, "/api/v1/groups/"+groupID, map[string]string{
		"name": "New Name",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "New Name" {
		t.Errorf("expected name 'New Name', got %v", resp["name"])
	}
}

func TestUpdateGroup_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPut, "/api/v1/groups/does-not-exist", map[string]string{
		"name": "Whatever",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDeleteGroup_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "To Delete", "type": "ZONE", "site_id": "s1",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	w := doJSON(r, http.MethodDelete, "/api/v1/groups/"+groupID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify it is gone
	getResp := doJSON(r, http.MethodGet, "/api/v1/groups/"+groupID, nil)
	if getResp.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.Code)
	}
}

func TestDeleteGroup_Idempotent(t *testing.T) {
	_, r := setupSuperappRouter()

	// Deleting a non-existent group should still return 204 (idempotent delete)
	w := doJSON(r, http.MethodDelete, "/api/v1/groups/nonexistent", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for idempotent delete, got %d", w.Code)
	}
}

func TestAddDevicesToGroup_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "Dev Group", "type": "BUILDING", "site_id": "s1",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	w := doJSON(r, http.MethodPost, "/api/v1/groups/"+groupID+"/devices", map[string]interface{}{
		"device_ids": []string{"dev-1", "dev-2", "dev-3"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)

	// device_ids in the response should have 3 entries
	devicesRaw, ok := resp["device_ids"].([]interface{})
	if !ok {
		t.Fatalf("expected device_ids to be an array, got %T", resp["device_ids"])
	}
	if len(devicesRaw) != 3 {
		t.Errorf("expected 3 device_ids, got %d", len(devicesRaw))
	}
}

func TestAddDevicesToGroup_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/groups/fake-group/devices", map[string]interface{}{
		"device_ids": []string{"dev-1"},
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAddDevicesToGroup_Appends(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "AppendGroup", "type": "CUSTOM", "site_id": "s1",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	// First addition
	doJSON(r, http.MethodPost, "/api/v1/groups/"+groupID+"/devices", map[string]interface{}{
		"device_ids": []string{"dev-1"},
	})
	// Second addition
	w := doJSON(r, http.MethodPost, "/api/v1/groups/"+groupID+"/devices", map[string]interface{}{
		"device_ids": []string{"dev-2", "dev-3"},
	})

	resp := decodeJSON(w)
	devicesRaw := resp["device_ids"].([]interface{})
	if len(devicesRaw) != 3 {
		t.Errorf("expected 3 devices after two appends, got %d", len(devicesRaw))
	}
}

// ============================================================================
// Config Templates
// ============================================================================

func TestCreateTemplate_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/templates", map[string]interface{}{
		"name":        "Camera Config",
		"device_type": "CAMERA",
		"config":      map[string]string{"resolution": "4K", "fps": "30"},
		"description": "Default camera config",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Camera Config" {
		t.Errorf("expected name 'Camera Config', got %v", resp["name"])
	}
	if resp["template_id"] == nil || resp["template_id"] == "" {
		t.Error("expected template_id to be set")
	}
}

func TestCreateTemplate_MissingName(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/templates", map[string]interface{}{
		"device_type": "CAMERA",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "name is required" {
		t.Errorf("expected 'name is required', got %v", resp["error"])
	}
}

func TestCreateTemplate_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListTemplates_Empty(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodGet, "/api/v1/templates", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(0) {
		t.Errorf("expected total 0, got %v", resp["total"])
	}
}

func TestGetTemplate_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/templates", map[string]interface{}{
		"name":        "Sensor Config",
		"device_type": "SENSOR",
	})
	created := decodeJSON(createResp)
	tmplID := created["template_id"].(string)

	w := doJSON(r, http.MethodGet, "/api/v1/templates/"+tmplID, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Sensor Config" {
		t.Errorf("expected 'Sensor Config', got %v", resp["name"])
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodGet, "/api/v1/templates/no-such-template", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "template not found" {
		t.Errorf("expected 'template not found', got %v", resp["error"])
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/templates", map[string]interface{}{
		"name":        "Temp",
		"device_type": "CAMERA",
	})
	created := decodeJSON(createResp)
	tmplID := created["template_id"].(string)

	w := doJSON(r, http.MethodDelete, "/api/v1/templates/"+tmplID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify gone
	getResp := doJSON(r, http.MethodGet, "/api/v1/templates/"+tmplID, nil)
	if getResp.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", getResp.Code)
	}
}

func TestDeleteTemplate_Idempotent(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodDelete, "/api/v1/templates/nonexistent", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for idempotent delete, got %d", w.Code)
	}
}

// ============================================================================
// Bulk Operations
// ============================================================================

func TestBulkDeleteDevices_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/devices/bulk-delete", map[string]interface{}{
		"device_ids": []string{"dev-1", "dev-2"},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["action"] != "bulk_delete" {
		t.Errorf("expected action 'bulk_delete', got %v", resp["action"])
	}
	if resp["status"] != "queued" {
		t.Errorf("expected status 'queued', got %v", resp["status"])
	}
	if resp["total"] != float64(2) {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestBulkDeleteDevices_EmptyIDs(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/devices/bulk-delete", map[string]interface{}{
		"device_ids": []string{},
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "device_ids required" {
		t.Errorf("expected 'device_ids required', got %v", resp["error"])
	}
}

func TestBulkDeleteDevices_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bulk-delete", bytes.NewReader([]byte("xxx")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBulkUpdateDevices_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/devices/bulk-update", map[string]interface{}{
		"device_ids": []string{"dev-1"},
		"updates":    map[string]string{"firmware_version": "2.0.0"},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["action"] != "bulk_update" {
		t.Errorf("expected action 'bulk_update', got %v", resp["action"])
	}
	if resp["status"] != "queued" {
		t.Errorf("expected status 'queued', got %v", resp["status"])
	}
}

func TestBulkTagDevices_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/devices/bulk-tag", map[string]interface{}{
		"device_ids": []string{"dev-1", "dev-2", "dev-3"},
		"tags":       []string{"production", "critical"},
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["action"] != "bulk_tag" {
		t.Errorf("expected action 'bulk_tag', got %v", resp["action"])
	}
	if resp["status"] != "queued" {
		t.Errorf("expected status 'queued', got %v", resp["status"])
	}
}

func TestExportDevices_JSON(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/export?format=json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected application/json content-type, got %s", ct)
	}
	cd := w.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "devices.json") {
		t.Errorf("expected devices.json in content-disposition, got %s", cd)
	}
}

func TestExportDevices_CSV(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/export?format=csv", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Errorf("expected text/csv content-type, got %s", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "device_id,serial_number") {
		t.Errorf("expected CSV header in body, got %s", body)
	}
}

func TestExportDevices_DefaultFormat(t *testing.T) {
	_, r := setupSuperappRouter()

	// No format query param -> should default to JSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/export", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected default to be JSON, got content-type %s", ct)
	}
}

// ============================================================================
// Webhooks
// ============================================================================

func TestCreateWebhook_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name":   "Alert Hook",
		"url":    "https://example.com/hook",
		"events": []string{"DEVICE_CREATED", "ALERT_TRIGGERED"},
		"headers": map[string]string{"Authorization": "Bearer xyz"},
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Alert Hook" {
		t.Errorf("expected name 'Alert Hook', got %v", resp["name"])
	}
	if resp["active"] != true {
		t.Errorf("expected active=true, got %v", resp["active"])
	}
	if resp["webhook_id"] == nil || resp["webhook_id"] == "" {
		t.Error("expected webhook_id to be set")
	}
}

func TestCreateWebhook_MissingURL(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name":   "No URL",
		"events": []string{"DEVICE_CREATED"},
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "url is required" {
		t.Errorf("expected 'url is required', got %v", resp["error"])
	}
}

func TestCreateWebhook_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks", bytes.NewReader([]byte("garbage")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListWebhooks_AfterCreate(t *testing.T) {
	_, r := setupSuperappRouter()

	doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name": "Hook 1", "url": "https://a.com",
		"events": []string{"DEVICE_CREATED"},
	})
	doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name": "Hook 2", "url": "https://b.com",
		"events": []string{"ALERT_TRIGGERED"},
	})

	w := doJSON(r, http.MethodGet, "/api/v1/webhooks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(2) {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestDeleteWebhook_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name": "To Delete", "url": "https://c.com", "events": []string{},
	})
	created := decodeJSON(createResp)
	whID := created["webhook_id"].(string)

	w := doJSON(r, http.MethodDelete, "/api/v1/webhooks/"+whID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestTestWebhook_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name": "Testable", "url": "https://example.com/test",
		"events": []string{"DEVICE_CREATED"},
	})
	created := decodeJSON(createResp)
	whID := created["webhook_id"].(string)

	w := doJSON(r, http.MethodPost, "/api/v1/webhooks/"+whID+"/test", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["status"] != "test_sent" {
		t.Errorf("expected status 'test_sent', got %v", resp["status"])
	}
}

func TestTestWebhook_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/webhooks/nonexistent/test", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "webhook not found" {
		t.Errorf("expected 'webhook not found', got %v", resp["error"])
	}
}

// ============================================================================
// API Keys
// ============================================================================

func TestCreateAPIKey_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/api-keys", map[string]interface{}{
		"name":  "Service Key",
		"role":  "admin",
		"org_id": "org-1",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Service Key" {
		t.Errorf("expected name 'Service Key', got %v", resp["name"])
	}
	if resp["active"] != true {
		t.Errorf("expected active=true, got %v", resp["active"])
	}
	key, ok := resp["key"].(string)
	if !ok || !strings.HasPrefix(key, "sk-sentinel-") {
		t.Errorf("expected key to start with 'sk-sentinel-', got %v", resp["key"])
	}
}

func TestCreateAPIKey_WithExpiration(t *testing.T) {
	_, r := setupSuperappRouter()

	expiresIn := int64(3600)
	w := doJSON(r, http.MethodPost, "/api/v1/api-keys", map[string]interface{}{
		"name":                "Temp Key",
		"role":                "viewer",
		"org_id":              "org-2",
		"expires_in_seconds":  expiresIn,
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["expires_at"] == nil {
		t.Error("expected expires_at to be set when expires_in_seconds is provided")
	}
}

func TestCreateAPIKey_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/api-keys", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListAPIKeys_Redacted(t *testing.T) {
	_, r := setupSuperappRouter()

	doJSON(r, http.MethodPost, "/api/v1/api-keys", map[string]interface{}{
		"name": "Key 1", "role": "admin", "org_id": "org-1",
	})

	w := doJSON(r, http.MethodGet, "/api/v1/api-keys", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(1) {
		t.Errorf("expected total 1, got %v", resp["total"])
	}

	// The key in the listing should be redacted (contains asterisks)
	keysRaw, ok := resp["api_keys"].([]interface{})
	if !ok || len(keysRaw) != 1 {
		t.Fatalf("expected api_keys array with 1 entry")
	}
	keyEntry := keysRaw[0].(map[string]interface{})
	listedKey := keyEntry["key"].(string)
	if !strings.Contains(listedKey, "*") {
		t.Errorf("expected redacted key to contain '*', got %s", listedKey)
	}
}

func TestDeleteAPIKey_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/api-keys", map[string]interface{}{
		"name": "To Delete", "role": "viewer", "org_id": "o1",
	})
	created := decodeJSON(createResp)
	keyID := created["key_id"].(string)

	w := doJSON(r, http.MethodDelete, "/api/v1/api-keys/"+keyID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify listing no longer has it
	listResp := doJSON(r, http.MethodGet, "/api/v1/api-keys", nil)
	listBody := decodeJSON(listResp)
	if listBody["total"] != float64(0) {
		t.Errorf("expected total 0 after delete, got %v", listBody["total"])
	}
}

// ============================================================================
// Notifications
// ============================================================================

func TestListNotifications_TriggeredByGroupCreation(t *testing.T) {
	_, r := setupSuperappRouter()

	// Creating a group triggers a notification
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "NotifyGroup", "type": "ZONE", "site_id": "s1",
	})

	w := doJSON(r, http.MethodGet, "/api/v1/notifications", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(1) {
		t.Errorf("expected total 1 notification after group creation, got %v", resp["total"])
	}
}

func TestMarkNotificationRead_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	// Create a group to trigger a notification
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "ReadNotify", "type": "ZONE", "site_id": "s1",
	})

	// Get the notification ID
	listResp := doJSON(r, http.MethodGet, "/api/v1/notifications", nil)
	listBody := decodeJSON(listResp)
	notifs := listBody["notifications"].([]interface{})
	if len(notifs) == 0 {
		t.Fatal("expected at least one notification")
	}
	notif := notifs[0].(map[string]interface{})
	notifID := notif["notification_id"].(string)

	// Mark as read
	w := doJSON(r, http.MethodPost, "/api/v1/notifications/"+notifID+"/read", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["read"] != true {
		t.Errorf("expected read=true, got %v", resp["read"])
	}
}

func TestMarkNotificationRead_NotFound(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/notifications/nonexistent/read", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["error"] != "notification not found" {
		t.Errorf("expected 'notification not found', got %v", resp["error"])
	}
}

func TestDeleteNotification_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	// Create a group to trigger a notification
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "DelNotify", "type": "ZONE", "site_id": "s1",
	})

	// Get the notification ID
	listResp := doJSON(r, http.MethodGet, "/api/v1/notifications", nil)
	listBody := decodeJSON(listResp)
	notifs := listBody["notifications"].([]interface{})
	notif := notifs[0].(map[string]interface{})
	notifID := notif["notification_id"].(string)

	// Delete it
	w := doJSON(r, http.MethodDelete, "/api/v1/notifications/"+notifID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify list is empty
	listResp2 := doJSON(r, http.MethodGet, "/api/v1/notifications", nil)
	listBody2 := decodeJSON(listResp2)
	if listBody2["total"] != float64(0) {
		t.Errorf("expected total 0 after deletion, got %v", listBody2["total"])
	}
}

func TestDeleteNotification_Idempotent(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodDelete, "/api/v1/notifications/nonexistent", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for idempotent delete, got %d", w.Code)
	}
}

// ============================================================================
// Geofences
// ============================================================================

func TestCreateGeofence_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/geofences", map[string]interface{}{
		"name":          "HQ Perimeter",
		"center_lat":    37.7749,
		"center_lng":    -122.4194,
		"radius_meters": 500,
		"site_id":       "site-1",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "HQ Perimeter" {
		t.Errorf("expected name 'HQ Perimeter', got %v", resp["name"])
	}
	if resp["zone_id"] == nil || resp["zone_id"] == "" {
		t.Error("expected zone_id to be set")
	}
	if resp["center_lat"] != 37.7749 {
		t.Errorf("expected center_lat 37.7749, got %v", resp["center_lat"])
	}
}

func TestCreateGeofence_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/geofences", bytes.NewReader([]byte("notjson")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListGeofences_AfterCreate(t *testing.T) {
	_, r := setupSuperappRouter()

	doJSON(r, http.MethodPost, "/api/v1/geofences", map[string]interface{}{
		"name": "Zone A", "center_lat": 1.0, "center_lng": 2.0,
		"radius_meters": 100, "site_id": "s1",
	})
	doJSON(r, http.MethodPost, "/api/v1/geofences", map[string]interface{}{
		"name": "Zone B", "center_lat": 3.0, "center_lng": 4.0,
		"radius_meters": 200, "site_id": "s2",
	})

	w := doJSON(r, http.MethodGet, "/api/v1/geofences", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["total"] != float64(2) {
		t.Errorf("expected total 2, got %v", resp["total"])
	}
}

func TestDeleteGeofence_Success(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/geofences", map[string]interface{}{
		"name": "To Delete", "center_lat": 0, "center_lng": 0,
		"radius_meters": 50, "site_id": "s1",
	})
	created := decodeJSON(createResp)
	zoneID := created["zone_id"].(string)

	w := doJSON(r, http.MethodDelete, "/api/v1/geofences/"+zoneID, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify list is empty
	listResp := doJSON(r, http.MethodGet, "/api/v1/geofences", nil)
	listBody := decodeJSON(listResp)
	if listBody["total"] != float64(0) {
		t.Errorf("expected total 0 after delete, got %v", listBody["total"])
	}
}

func TestDeleteGeofence_Idempotent(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodDelete, "/api/v1/geofences/nonexistent", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for idempotent delete, got %d", w.Code)
	}
}

// ============================================================================
// Cross-cutting / Edge Cases
// ============================================================================

func TestCreateGroup_NotificationGenerated(t *testing.T) {
	h := NewSuperappHandler()
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Create two groups
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "G1", "type": "ZONE", "site_id": "s1",
	})
	doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "G2", "type": "BUILDING", "site_id": "s1",
	})

	// Should have 2 notifications
	w := doJSON(r, http.MethodGet, "/api/v1/notifications", nil)
	resp := decodeJSON(w)
	if resp["total"] != float64(2) {
		t.Errorf("expected 2 notifications from 2 group creates, got %v", resp["total"])
	}

	// Verify notification content
	notifs := resp["notifications"].([]interface{})
	titles := map[string]bool{}
	for _, n := range notifs {
		notif := n.(map[string]interface{})
		titles[notif["title"].(string)] = true
		if notif["type"] != "SYSTEM" {
			t.Errorf("expected notification type SYSTEM, got %v", notif["type"])
		}
		if notif["read"] != false {
			t.Errorf("expected new notification to be unread, got read=%v", notif["read"])
		}
	}
	if !titles["Group Created"] {
		t.Error("expected notification title 'Group Created'")
	}
}

func TestCreateAPIKey_WithoutExpiration(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/api-keys", map[string]interface{}{
		"name": "No Expire", "role": "viewer", "org_id": "o1",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["expires_at"] != nil {
		t.Errorf("expected expires_at to be nil when not provided, got %v", resp["expires_at"])
	}
}

func TestCreateWebhook_WithHeaders(t *testing.T) {
	_, r := setupSuperappRouter()

	w := doJSON(r, http.MethodPost, "/api/v1/webhooks", map[string]interface{}{
		"name":    "Header Hook",
		"url":     "https://example.com/h",
		"events":  []string{"DEVICE_CREATED"},
		"headers": map[string]string{"X-Custom": "value", "Authorization": "Bearer t"},
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	headers, ok := resp["headers"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected headers to be a map, got %T", resp["headers"])
	}
	if headers["X-Custom"] != "value" {
		t.Errorf("expected X-Custom header 'value', got %v", headers["X-Custom"])
	}
}

func TestBulkUpdateDevices_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bulk-update", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBulkTagDevices_InvalidBody(t *testing.T) {
	_, r := setupSuperappRouter()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/devices/bulk-tag", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateGroup_EmptyBody(t *testing.T) {
	_, r := setupSuperappRouter()

	createResp := doJSON(r, http.MethodPost, "/api/v1/groups", map[string]string{
		"name": "Original", "type": "ZONE", "site_id": "s1",
	})
	created := decodeJSON(createResp)
	groupID := created["group_id"].(string)

	// Empty body should succeed and not change anything
	w := doJSON(r, http.MethodPut, "/api/v1/groups/"+groupID, map[string]string{})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["name"] != "Original" {
		t.Errorf("expected name to remain 'Original', got %v", resp["name"])
	}
}

func TestCreateGeofence_ZeroCoordinates(t *testing.T) {
	_, r := setupSuperappRouter()

	// Zero coordinates are valid (e.g., null island); the handler does not validate ranges
	w := doJSON(r, http.MethodPost, "/api/v1/geofences", map[string]interface{}{
		"name":          "Null Island",
		"center_lat":    0,
		"center_lng":    0,
		"radius_meters": 100,
		"site_id":       "site-0",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
	resp := decodeJSON(w)
	if resp["center_lat"] != float64(0) {
		t.Errorf("expected center_lat 0, got %v", resp["center_lat"])
	}
}
