package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMiddlewareRecordsSuccess(t *testing.T) {
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	count, err := testutil.GatherAndCount(
		prometheus.DefaultGatherer,
		"sentinel_http_requests_total",
	)
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	if count == 0 {
		t.Error("expected at least one http_requests_total metric")
	}
}

func TestPrometheusMiddlewareRecordsDuration(t *testing.T) {
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_http_request_duration_seconds" {
			if len(f.GetMetric()) == 0 {
				t.Error("expected duration observations")
			}
			return
		}
	}
	t.Error("sentinel_http_request_duration_seconds not found")
}

func TestPrometheusMiddlewareRecordsErrorStatus(t *testing.T) {
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/fail", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_http_requests_total" {
			for _, m := range f.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "status_code" && l.GetValue() == "500" {
						return
					}
				}
			}
		}
	}
	t.Error("500 status code not found in metrics")
}

func TestRecordDBQuery(t *testing.T) {
	RecordDBQuery("get_device", 50*time.Millisecond, nil)
	RecordDBQuery("put_device", 100*time.Millisecond, fmt.Errorf("timeout"))

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	var foundDuration, foundError bool
	for _, f := range families {
		switch f.GetName() {
		case "sentinel_db_query_duration_seconds":
			foundDuration = true
		case "sentinel_db_errors_total":
			foundError = true
		}
	}
	if !foundDuration {
		t.Error("db_query_duration_seconds not recorded")
	}
	if !foundError {
		t.Error("db_errors_total not recorded after error")
	}
}

func TestRecordGRPCRequest(t *testing.T) {
	RecordGRPCRequest("GetDevice", "OK", 25*time.Millisecond)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_grpc_requests_total" {
			return
		}
	}
	t.Error("sentinel_grpc_requests_total not found")
}

func TestMetricsHandler(t *testing.T) {
	handler := MetricsHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "sentinel_http_requests_total") {
		t.Error("metrics output missing sentinel_http_requests_total")
	}
}

func TestResponseWriterCapture(t *testing.T) {
	inner := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: inner, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	if rw.Status() != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rw.Status())
	}

	n, err := rw.Write([]byte("test data"))
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if n != 9 {
		t.Errorf("expected 9 bytes, got %d", n)
	}
	if rw.Size() != 9 {
		t.Errorf("expected size 9, got %d", rw.Size())
	}
}

func TestSetDevicesManaged(t *testing.T) {
	SetDevicesManaged(42)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_devices_managed_total" {
			if f.GetMetric()[0].GetGauge().GetValue() != 42 {
				t.Error("devices_managed should be 42")
			}
			return
		}
	}
	t.Error("sentinel_devices_managed_total not found")
}

func TestRecordIncidentCreated(t *testing.T) {
	RecordIncidentCreated("CRITICAL")
	RecordIncidentCreated("LOW")

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_incidents_created_total" {
			var foundCritical, foundLow bool
			for _, m := range f.GetMetric() {
				for _, l := range m.GetLabel() {
					if l.GetName() == "severity" {
						switch l.GetValue() {
						case "CRITICAL":
							foundCritical = true
						case "LOW":
							foundLow = true
						}
					}
				}
			}
			if !foundCritical || !foundLow {
				t.Error("missing severity labels in incidents metric")
			}
			return
		}
	}
	t.Error("sentinel_incidents_created_total not found")
}

func TestRecordComplianceCheck(t *testing.T) {
	RecordComplianceCheck("ISO27001", "PASS")
	RecordComplianceCheck("GDPR", "FAIL")

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_compliance_checks_total" {
			return
		}
	}
	t.Error("sentinel_compliance_checks_total not found")
}

func TestInFlightRequestsGauge(t *testing.T) {
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/long", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		req := httptest.NewRequest(http.MethodGet, "/long", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}()

	time.Sleep(10 * time.Millisecond)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_http_requests_in_flight" {
			return
		}
	}
	t.Error("sentinel_http_requests_in_flight not found")
}

func TestResponseSizeTracking(t *testing.T) {
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware)
	r.Get("/big", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(strings.Repeat("x", 1000)))
	})

	req := httptest.NewRequest(http.MethodGet, "/big", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	for _, f := range families {
		if f.GetName() == "sentinel_http_response_size_bytes" {
			return
		}
	}
	t.Error("sentinel_http_response_size_bytes not found")
}
