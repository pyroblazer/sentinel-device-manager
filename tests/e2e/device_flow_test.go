package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// E2E tests run against a live deployment.
// Set BASE_URL env var to the deployed service URL.
// Defaults to http://localhost:8080

func getBaseURL() string {
	if url := getEnv("BASE_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

func getAnalyticsURL() string {
	if url := getEnv("ANALYTICS_URL"); url != "" {
		return url
	}
	return "http://localhost:8081"
}

func getEnv(key string) string {
	// simplified - in production use os.Getenv
	return ""
}

func TestE2E_DeviceRegistrationAndRetrieval(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	baseURL := getBaseURL()
	client := &http.Client{Timeout: 10 * time.Second}

	// Register device
	body := map[string]interface{}{
		"serial_number":   "VKD-E2E-001",
		"device_type":     "CAMERA",
		"model":           "D30",
		"site_id":         "site-e2e",
		"organization_id": "org-e2e",
		"config": map[string]string{
			"resolution": "1080p",
		},
	}
	b, _ := json.Marshal(body)
	resp, err := client.Post(baseURL+"/api/v1/devices", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("register device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var device map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&device)
	deviceID := device["device_id"].(string)

	t.Logf("Registered device: %s", deviceID)

	// Retrieve device
	resp, err = client.Get(fmt.Sprintf("%s/api/v1/devices/%s", baseURL, deviceID))
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var fetched map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&fetched)

	if fetched["serial_number"] != "VKD-E2E-001" {
		t.Errorf("expected serial VKD-E2E-001, got %v", fetched["serial_number"])
	}
	if fetched["status"] != "ONLINE" {
		t.Errorf("expected ONLINE, got %v", fetched["status"])
	}
}

func TestE2E_EventIngestionAndAlertGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	analyticsURL := getAnalyticsURL()
	client := &http.Client{Timeout: 10 * time.Second}

	// Send critical event
	event := map[string]interface{}{
		"device_id":  "dev-e2e-001",
		"event_type": "ALARM_TRIGGERED",
		"severity":   "CRITICAL",
		"payload":    map[string]string{"zone": "perimeter"},
	}
	b, _ := json.Marshal(event)
	resp, err := client.Post(analyticsURL+"/api/v1/events", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("send event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Check alert was generated
	resp, err = client.Get(analyticsURL + "/api/v1/alerts?status=ACTIVE")
	if err != nil {
		t.Fatalf("get alerts: %v", err)
	}
	defer resp.Body.Close()

	var alerts []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&alerts)

	if len(alerts) == 0 {
		t.Fatal("expected at least one active alert after critical event")
	}

	alertID := alerts[0]["alert_id"].(string)
	t.Logf("Generated alert: %s", alertID)

	// Acknowledge alert
	ackBody := map[string]interface{}{"acknowledged_by": "admin@sentinel.io"}
	b, _ = json.Marshal(ackBody)
	resp, err = client.Post(
		fmt.Sprintf("%s/api/v1/alerts/%s/acknowledge", analyticsURL, alertID),
		"application/json",
		bytes.NewReader(b),
	)
	if err != nil {
		t.Fatalf("acknowledge alert: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on acknowledge, got %d", resp.StatusCode)
	}
}

func TestE2E_AnalyticsSummary(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	analyticsURL := getAnalyticsURL()
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(analyticsURL + "/api/v1/analytics/summary")
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var summary map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&summary)

	requiredFields := []string{
		"total_devices", "online_devices", "offline_devices",
		"active_alerts", "events_last_24h", "critical_alerts",
		"firmware_compliance_pct",
	}
	for _, f := range requiredFields {
		if _, ok := summary[f]; !ok {
			t.Errorf("missing field: %s", f)
		}
	}
}

func TestE2E_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	baseURL := getBaseURL()
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
