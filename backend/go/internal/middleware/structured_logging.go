package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	loggerKey        contextKey = "structured_logger"
)

// NewStructuredLogger creates a zap.Logger configured for JSON output to stdout.
// Every log entry includes the service name, ISO 8601 timestamp, caller, and level.
func NewStructuredLogger(serviceName string) *zap.Logger {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service": serviceName,
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to create structured logger: " + err.Error())
	}
	return logger
}

// CorrelationIDMiddleware injects a correlation ID into the request context.
// Priority: X-Correlation-ID header > X-Request-ID header > generated timestamp ID.
// The correlation ID is set on the response header for end-to-end traceability.
func CorrelationIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			if rid := r.Header.Get("X-Request-ID"); rid != "" {
				correlationID = rid
			} else {
				correlationID = fmt.Sprintf("%d", time.Now().UnixNano())
			}
		}

		ctx := context.WithValue(r.Context(), correlationIDKey, correlationID)
		w.Header().Set("X-Correlation-ID", correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLoggerMiddleware logs each HTTP request as structured JSON.
// Entries include correlation_id, trace_id, method, path, status, duration, and size.
func RequestLoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			if cid, ok := r.Context().Value(correlationIDKey).(string); ok {
				fields = append(fields, zap.String("correlation_id", cid))
			}
			if traceID := TraceIDFromContext(r.Context()); traceID != "" {
				fields = append(fields, zap.String("trace_id", traceID))
			}

			reqLogger := logger.With(fields...)
			ctx := context.WithValue(r.Context(), loggerKey, reqLogger)

			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(ww, r.WithContext(ctx))

			reqLogger.Info("http_request_completed",
				zap.Int("status_code", ww.statusCode),
				zap.Duration("duration", time.Since(start)),
				zap.Int("response_size_bytes", ww.size),
			)
		})
	}
}

// GetLogger extracts the request-scoped structured logger from context.
// Returns a no-op logger if none is found.
func GetLogger(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}

// GetCorrelationID extracts the correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if cid, ok := ctx.Value(correlationIDKey).(string); ok {
		return cid
	}
	return ""
}
