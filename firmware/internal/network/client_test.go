package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegister(t *testing.T) {
	wantID := "test-device-uuid"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices" {
			t.Errorf("expected /api/v1/devices, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type")
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if body["serial_number"] != "SN-001" {
			t.Errorf("expected serial_number=SN-001, got %v", body["serial_number"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"device_id":     wantID,
			"serial_number": "SN-001",
			"status":        "ONLINE",
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	resp, err := client.Register(context.Background(), map[string]interface{}{
		"serial_number": "SN-001",
		"device_type":   "CAMERA",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if resp["device_id"] != wantID {
		t.Errorf("expected device_id=%s, got %v", wantID, resp["device_id"])
	}
}

func TestRegisterServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "")
	_, err := client.Register(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestSendHeartbeat(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/api/v1/devices/dev-123/heartbeat" {
			t.Errorf("expected /api/v1/devices/dev-123/heartbeat, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"status": "ok"}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "dev-123")
	err := client.SendHeartbeat(context.Background(), map[string]interface{}{
		"cpu_usage": 45.2,
	})
	if err != nil {
		t.Fatalf("SendHeartbeat failed: %v", err)
	}
	if !called {
		t.Fatal("expected heartbeat endpoint to be called")
	}
}

func TestSendHeartbeatEmptyDeviceID(t *testing.T) {
	client := NewClient("http://localhost:8080", "")
	err := client.SendHeartbeat(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for empty device ID")
	}
	if err.Error() != "cannot send heartbeat: device not registered (empty device ID)" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestSendHeartbeatNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "dev-123")
	err := client.SendHeartbeat(context.Background(), map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestSendEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/events" {
			t.Errorf("expected /api/v1/events, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"event_id": "evt-001"}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient("http://unused:8080", "dev-123")
	resp, err := client.SendEvent(context.Background(), srv.URL, map[string]interface{}{
		"device_id":  "dev-123",
		"event_type": "MOTION_DETECTED",
	})
	if err != nil {
		t.Fatalf("SendEvent failed: %v", err)
	}
	if resp["event_id"] != "evt-001" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestGetConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/dev-456" {
			t.Errorf("expected /api/v1/devices/dev-456, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]interface{}{
			"device_id": "dev-456",
			"config":    map[string]string{"resolution": "4K"},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "dev-456")
	resp, err := client.GetConfig(context.Background())
	if err != nil {
		t.Fatalf("GetConfig failed: %v", err)
	}
	if resp["device_id"] != "dev-456" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestGetConfigEmptyDeviceID(t *testing.T) {
	client := NewClient("http://localhost:8080", "")
	_, err := client.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected error for empty device ID")
	}
}

func TestReportFirmwareStatusEmptyDeviceID(t *testing.T) {
	client := NewClient("http://localhost:8080", "")
	err := client.ReportFirmwareStatus(context.Background(), "1.0.0", "COMPLETED")
	if err == nil {
		t.Fatal("expected error for empty device ID")
	}
}

func TestSetDeviceID(t *testing.T) {
	client := NewClient("http://localhost:8080", "")
	if client.DeviceID() != "" {
		t.Fatal("expected empty device ID initially")
	}
	client.SetDeviceID("new-id")
	if client.DeviceID() != "new-id" {
		t.Errorf("expected new-id, got %s", client.DeviceID())
	}
}

func TestDecodeResponseNonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "dev")
	_, err := client.GetConfig(context.Background())
	if err == nil {
		t.Fatal("expected error for non-JSON response")
	}
}
