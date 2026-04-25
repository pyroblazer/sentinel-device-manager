package compliance

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewComplianceReporter(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	if cr == nil {
		t.Fatal("expected non-nil reporter")
	}
	if cr.version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cr.version)
	}
}

func TestGetStandards(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	standards := cr.GetStandards()

	expectedIDs := []string{
		"ISO27001", "ISO9001", "ISO27035", "ISO27017",
		"ISO20000", "ISO22301", "IEC62443", "NIST-CSF", "SOC2", "GDPR",
	}

	if len(standards) != len(expectedIDs) {
		t.Errorf("expected %d standards, got %d", len(expectedIDs), len(standards))
	}

	found := make(map[string]bool)
	for _, s := range standards {
		found[s.ID] = true
		if s.Name == "" {
			t.Errorf("standard %s missing name", s.ID)
		}
		if len(s.Controls) == 0 {
			t.Errorf("standard %s has no controls", s.ID)
		}
	}

	for _, id := range expectedIDs {
		if !found[id] {
			t.Errorf("missing standard: %s", id)
		}
	}
}

func TestGetStandard(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	s := cr.GetStandard("ISO27001")
	if s == nil {
		t.Fatal("expected ISO27001 standard")
	}
	if s.Category != "Information Security" {
		t.Errorf("expected Information Security, got %s", s.Category)
	}

	s = cr.GetStandard("NONEXISTENT")
	if s != nil {
		t.Error("expected nil for nonexistent standard")
	}
}

func TestGetSummary(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	summary := cr.GetSummary()

	if summary.TotalControls == 0 {
		t.Error("expected non-zero total controls")
	}
	if summary.Implemented == 0 {
		t.Error("expected non-zero implemented controls")
	}
	if summary.CompliancePct <= 0 {
		t.Error("expected positive compliance percentage")
	}
	if len(summary.Standards) == 0 {
		t.Error("expected standards in summary")
	}

	for id, ss := range summary.Standards {
		if ss.CompliancePct < 0 || ss.CompliancePct > 100 {
			t.Errorf("standard %s: compliance %.1f out of range", id, ss.CompliancePct)
		}
	}
}

func TestCreateIncident(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	inc := cr.CreateIncident(
		"Test Incident", "Something went wrong",
		"HIGH", "SECURITY", "user-1",
	)

	if inc.IncidentID == "" {
		t.Error("expected incident ID")
	}
	if inc.Title != "Test Incident" {
		t.Errorf("expected 'Test Incident', got %s", inc.Title)
	}
	if inc.Severity != "HIGH" {
		t.Errorf("expected HIGH, got %s", inc.Severity)
	}
	if inc.Status != "OPEN" {
		t.Errorf("expected OPEN, got %s", inc.Status)
	}
	if len(inc.Timeline) != 1 {
		t.Errorf("expected 1 timeline entry, got %d", len(inc.Timeline))
	}
	if inc.ResolvedAt != nil {
		t.Error("new incident should not have resolved_at")
	}
}

func TestGetIncidents(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	cr.CreateIncident("Incident 1", "desc", "HIGH", "SECURITY", "user-1")
	cr.CreateIncident("Incident 2", "desc", "LOW", "OPERATIONAL", "user-2")

	all := cr.GetIncidents("")
	if len(all) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(all))
	}

	open := cr.GetIncidents("OPEN")
	if len(open) != 2 {
		t.Errorf("expected 2 open incidents, got %d", len(open))
	}

	resolved := cr.GetIncidents("RESOLVED")
	if len(resolved) != 0 {
		t.Errorf("expected 0 resolved, got %d", len(resolved))
	}
}

func TestUpdateIncidentStatus(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	inc := cr.CreateIncident("Test", "desc", "CRITICAL", "SECURITY", "user-1")

	err := cr.UpdateIncidentStatus(inc.IncidentID, "INVESTIGATING", "admin-1", "Looking into it")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	incidents := cr.GetIncidents("INVESTIGATING")
	if len(incidents) != 1 {
		t.Fatalf("expected 1 investigating incident, got %d", len(incidents))
	}
	if len(incidents[0].Timeline) != 2 {
		t.Errorf("expected 2 timeline entries, got %d", len(incidents[0].Timeline))
	}

	// Resolve the incident
	err = cr.UpdateIncidentStatus(inc.IncidentID, "RESOLVED", "admin-1", "Root cause found")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved := cr.GetIncidents("RESOLVED")
	if resolved[0].ResolvedAt == nil {
		t.Error("expected resolved_at to be set")
	}
}

func TestUpdateIncidentStatus_NotFound(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	err := cr.UpdateIncidentStatus("NONEXISTENT", "RESOLVED", "admin", "")
	if err == nil {
		t.Error("expected error for nonexistent incident")
	}
}

func TestGetRetentionPolicies(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	policies := cr.GetRetentionPolicies()

	if len(policies) == 0 {
		t.Error("expected retention policies")
	}

	for _, p := range policies {
		if p.Category == "" {
			t.Error("policy missing category")
		}
		if p.RetentionDays <= 0 {
			t.Errorf("policy %s: invalid retention days %d", p.Category, p.RetentionDays)
		}
		if p.Description == "" {
			t.Errorf("policy %s missing description", p.Category)
		}
	}
}

func TestGetHealthStatus(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	components := []ComponentHealth{
		{Name: "rest_api", Status: "healthy"},
		{Name: "grpc", Status: "healthy"},
		{Name: "database", Status: "degraded", Details: "high latency"},
	}

	health := cr.GetHealthStatus(components)

	if health.Status != "degraded" {
		t.Errorf("expected degraded, got %s", health.Status)
	}
	if health.Version != "1.0.0" {
		t.Errorf("expected 1.0.0, got %s", health.Version)
	}
	if health.Uptime == "" {
		t.Error("expected uptime to be set")
	}
	if health.System.GoVersion == "" {
		t.Error("expected go version")
	}
	if health.System.NumCPU == 0 {
		t.Error("expected CPU count")
	}
}

func TestGetHealthStatus_UnhealthyComponent(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	components := []ComponentHealth{
		{Name: "rest_api", Status: "healthy"},
		{Name: "database", Status: "unhealthy"},
	}

	health := cr.GetHealthStatus(components)
	if health.Status != "unhealthy" {
		t.Errorf("expected unhealthy, got %s", health.Status)
	}
}

func TestHandleComplianceReport(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/report", nil)
	w := httptest.NewRecorder()
	cr.HandleComplianceReport(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)

	if result["version"] == nil {
		t.Error("expected version in report")
	}
	if result["standards"] == nil {
		t.Error("expected standards in report")
	}
	if result["summary"] == nil {
		t.Error("expected summary in report")
	}
}

func TestHandleStandardsList(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/standards", nil)
	w := httptest.NewRecorder()
	cr.HandleStandardsList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)
	standards := result["standards"].([]interface{})
	if len(standards) != 10 {
		t.Errorf("expected 10 standards, got %d", len(standards))
	}
}

func TestHandleIncidents(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	cr.CreateIncident("Test", "desc", "HIGH", "SECURITY", "user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/incidents?status=OPEN", nil)
	w := httptest.NewRecorder()
	cr.HandleIncidents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)
	incidents := result["incidents"].([]interface{})
	if len(incidents) != 1 {
		t.Errorf("expected 1 incident, got %d", len(incidents))
	}
}

func TestHandleRetentionPolicies(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/compliance/retention", nil)
	w := httptest.NewRecorder()
	cr.HandleRetentionPolicies(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result map[string]interface{}
	_ = json.NewDecoder(w.Body).Decode(&result)
	policies := result["policies"].([]interface{})
	if len(policies) == 0 {
		t.Error("expected retention policies")
	}
}

func TestControlStatuses_AreValid(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	validStatuses := map[string]bool{
		"IMPLEMENTED":    true,
		"PARTIAL":        true,
		"PLANNED":        true,
		"NOT_APPLICABLE": true,
	}

	for _, s := range cr.GetStandards() {
		for _, c := range s.Controls {
			if !validStatuses[c.Status] {
				t.Errorf("standard %s, control %s: invalid status %q", s.ID, c.ID, c.Status)
			}
		}
	}
}

func TestIncidentSeverity_Values(t *testing.T) {
	severities := []string{"LOW", "MEDIUM", "HIGH", "CRITICAL"}
	cr := NewComplianceReporter("1.0.0")

	for _, sev := range severities {
		inc := cr.CreateIncident("Test "+sev, "desc", sev, "SECURITY", "user-1")
		if inc.Severity != sev {
			t.Errorf("expected %s, got %s", sev, inc.Severity)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(n int) {
			cr.CreateIncident("Incident", "desc", "HIGH", "SECURITY", "user")
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			cr.GetSummary()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	incidents := cr.GetIncidents("")
	if len(incidents) != 5 {
		t.Errorf("expected 5 incidents, got %d", len(incidents))
	}
}

func TestIncidentTimelineOrder(t *testing.T) {
	cr := NewComplianceReporter("1.0.0")
	inc := cr.CreateIncident("Test", "desc", "HIGH", "SECURITY", "user-1")

	time.Sleep(10 * time.Millisecond)
	_ = cr.UpdateIncidentStatus(inc.IncidentID, "INVESTIGATING", "admin", "step 2")

	time.Sleep(10 * time.Millisecond)
	_ = cr.UpdateIncidentStatus(inc.IncidentID, "RESOLVED", "admin", "done")

	incidents := cr.GetIncidents("")
	if len(incidents) != 1 {
		t.Fatal("expected 1 incident")
	}

	timeline := incidents[0].Timeline
	if len(timeline) != 3 {
		t.Fatalf("expected 3 timeline entries, got %d", len(timeline))
	}

	for i := 1; i < len(timeline); i++ {
		if timeline[i].Timestamp.Before(timeline[i-1].Timestamp) {
			t.Errorf("timeline entry %d before %d", i, i-1)
		}
	}
}
