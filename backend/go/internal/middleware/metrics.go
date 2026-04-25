// Package middleware provides HTTP middleware for the Sentinel Device Manager
// including Prometheus metrics collection.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metric naming convention:
//
//	sentinel_<subsystem>_<name>_<unit>
//
// Counters use _total suffix. Histograms include the unit. All labels are snake_case.

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_http_requests_total",
		Help: "Total HTTP requests partitioned by method, path and status code.",
	}, []string{"method", "path", "status_code"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sentinel_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sentinel_http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed.",
	}, []string{"method"})

	httpResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sentinel_http_response_size_bytes",
		Help:    "HTTP response body size in bytes.",
		Buckets: prometheus.ExponentialBuckets(100, 10, 6),
	}, []string{"method", "path"})

	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sentinel_db_query_duration_seconds",
		Help:    "Database query latency in seconds.",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"operation"})

	dbErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_db_errors_total",
		Help: "Total database operation errors.",
	}, []string{"operation"})

	grpcRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_grpc_requests_total",
		Help: "Total gRPC requests partitioned by method and status code.",
	}, []string{"method", "status_code"})

	grpcRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "sentinel_grpc_request_duration_seconds",
		Help:    "gRPC request latency in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
	}, []string{"method"})

	devicesManaged = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "sentinel_devices_managed_total",
		Help: "Current number of devices under management.",
	})

	complianceChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_compliance_checks_total",
		Help: "Total compliance checks partitioned by standard and status.",
	}, []string{"standard", "status"})

	incidentsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sentinel_incidents_created_total",
		Help: "Total incidents created partitioned by severity.",
	}, []string{"severity"})
)

// MetricsHandler returns an HTTP handler that serves Prometheus-formatted metrics.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// PrometheusMiddleware records HTTP request metrics for every inbound request.
// It captures method, route pattern, status code, duration, and response size.
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		path := routePattern(r)

		httpRequestsInFlight.WithLabelValues(r.Method).Inc()
		defer httpRequestsInFlight.WithLabelValues(r.Method).Dec()

		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		statusStr := strconv.Itoa(ww.statusCode)
		httpRequestsTotal.WithLabelValues(r.Method, path, statusStr).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path).Observe(time.Since(start).Seconds())
		httpResponseSize.WithLabelValues(r.Method, path).Observe(float64(ww.size))
	})
}

// RecordHTTPRequest records metrics for an HTTP request for manual instrumentation.
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, responseSize int) {
	httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	httpResponseSize.WithLabelValues(method, path).Observe(float64(responseSize))
}

// RecordDBQuery records a database operation metric. Increments error counter on non-nil err.
func RecordDBQuery(operation string, duration time.Duration, err error) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	if err != nil {
		dbErrorsTotal.WithLabelValues(operation).Inc()
	}
}

// RecordGRPCRequest records a gRPC method call metric.
func RecordGRPCRequest(method, statusCode string, duration time.Duration) {
	grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
	grpcRequestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// SetDevicesManaged updates the device count gauge.
func SetDevicesManaged(count float64) {
	devicesManaged.Set(count)
}

// RecordComplianceCheck records a compliance standard check.
func RecordComplianceCheck(standard, status string) {
	complianceChecksTotal.WithLabelValues(standard, status).Inc()
}

// RecordIncidentCreated records incident creation by severity.
func RecordIncidentCreated(severity string) {
	incidentsCreatedTotal.WithLabelValues(severity).Inc()
}

// routePattern extracts the registered chi route pattern from the request context.
func routePattern(r *http.Request) string {
	rctx := chi.RouteContext(r.Context())
	if rctx != nil && rctx.RoutePattern() != "" {
		return rctx.RoutePattern()
	}
	return r.URL.Path
}

// responseWriter wraps http.ResponseWriter to capture status code and bytes written.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Status returns the captured HTTP status code.
func (rw *responseWriter) Status() int { return rw.statusCode }

// Size returns the total bytes written to the response body.
func (rw *responseWriter) Size() int { return rw.size }
