package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewStructuredLogger(t *testing.T) {
	logger := NewStructuredLogger("test-service")
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	logger.Info("test message")
}

func TestCorrelationIDFromHeader(t *testing.T) {
	r := chi.NewRouter()
	r.Use(CorrelationIDMiddleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		cid := GetCorrelationID(r.Context())
		if cid != "test-correlation-123" {
			t.Errorf("expected test-correlation-123, got %s", cid)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Correlation-ID", "test-correlation-123")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Header().Get("X-Correlation-ID") != "test-correlation-123" {
		t.Error("correlation ID not in response header")
	}
}

func TestCorrelationIDFallbackToRequestID(t *testing.T) {
	r := chi.NewRouter()
	r.Use(CorrelationIDMiddleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		cid := GetCorrelationID(r.Context())
		if cid != "request-id-456" {
			t.Errorf("expected request-id-456, got %s", cid)
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "request-id-456")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Header().Get("X-Correlation-ID") != "request-id-456" {
		t.Error("should fall back to X-Request-ID")
	}
}

func TestCorrelationIDGenerated(t *testing.T) {
	r := chi.NewRouter()
	r.Use(CorrelationIDMiddleware)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		cid := GetCorrelationID(r.Context())
		if cid == "" {
			t.Error("expected generated correlation ID")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	cid := rr.Header().Get("X-Correlation-ID")
	if cid == "" {
		t.Error("expected correlation ID in response header")
	}
}

func TestRequestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	r := chi.NewRouter()
	r.Use(CorrelationIDMiddleware)
	r.Use(RequestLoggerMiddleware(logger))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response"))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	output := buf.String()
	if !strings.Contains(output, "http_request_completed") {
		t.Errorf("expected log message 'http_request_completed', got: %s", output)
	}
	if !strings.Contains(output, "correlation_id") {
		t.Error("expected correlation_id in log output")
	}
	if !strings.Contains(output, "status_code") {
		t.Error("expected status_code in log output")
	}
}

func TestRequestLoggerJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	r := chi.NewRouter()
	r.Use(RequestLoggerMiddleware(logger))
	r.Get("/json-test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodGet, "/json-test", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	line := strings.TrimSpace(buf.String())
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("log output is not valid JSON: %v, got: %s", err, line)
	}
	if entry["message"] != "http_request_completed" {
		t.Errorf("unexpected message: %v", entry["message"])
	}
}

func TestRequestLoggerRecordsStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	r := chi.NewRouter()
	r.Use(RequestLoggerMiddleware(logger))
	r.Get("/server-error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	req := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	output := buf.String()
	if !strings.Contains(output, `"status_code":500`) {
		t.Errorf("expected status_code 500 in log, got: %s", output)
	}
}

func TestGetLoggerFromContext(t *testing.T) {
	logger := NewStructuredLogger("test")
	ctx := context.WithValue(context.Background(), loggerKey, logger)
	extracted := GetLogger(ctx)
	if extracted == nil {
		t.Error("expected to extract logger from context")
	}
}

func TestGetLoggerNilContext(t *testing.T) {
	logger := GetLogger(context.Background())
	// Should return no-op logger, not nil
	logger.Info("this should not panic")
}

func TestGetCorrelationIDEmpty(t *testing.T) {
	cid := GetCorrelationID(context.Background())
	if cid != "" {
		t.Errorf("expected empty correlation ID, got %s", cid)
	}
}

// newTestLogger creates a zap logger that writes JSON to the provided buffer.
func newTestLogger(buf *bytes.Buffer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(buf),
		zapcore.DebugLevel,
	)
	return zap.New(core)
}
