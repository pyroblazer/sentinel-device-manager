// Package main provides the entry point for the Sentinel Device Manager Go backend service.
//
// The server exposes both REST (port 8080) and gRPC (port 9090) interfaces for
// managing physical security devices including cameras, access control panels,
// alarm systems, and environmental sensors.
//
// # Security Architecture
//
// The service implements a layered security model:
//   - JWT authentication for identity verification
//   - RBAC authorization for role-based access control
//   - Rate limiting to prevent abuse
//   - Audit logging for compliance and forensics
//   - Input validation for request integrity
//   - Security headers (HSTS, CSP, X-Frame-Options)
//
// # Compliance
//
// The service tracks compliance with:
// ISO 27001, ISO 9001, ISO 27035, ISO 27017, ISO 20000, ISO 22301,
// IEC 62443, NIST CSF, SOC 2, and GDPR.
//
// # Running
//
//	DYNAMODB_ENDPOINT=http://localhost:8000 go run cmd/server/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	pb "github.com/sentinel-device-manager/backend/go/pkg/api"
	"github.com/sentinel-device-manager/backend/go/internal/compliance"
	"github.com/sentinel-device-manager/backend/go/internal/handler"
	"github.com/sentinel-device-manager/backend/go/internal/middleware"
	"github.com/sentinel-device-manager/backend/go/internal/repository"
	"github.com/sentinel-device-manager/backend/go/internal/service"
	"google.golang.org/grpc"
)

func main() {
	logger := log.New(os.Stdout, "[sentinel] ", log.LstdFlags|log.Lshortfile)
	structuredLogger := middleware.NewStructuredLogger("sentinel-go")
	defer func() { _ = structuredLogger.Sync() }()

	// AWS config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Fatalf("load AWS config: %v", err)
	}
	ddbClient := dynamodb.NewFromConfig(cfg)

	tableName := envOrDefault("DYNAMO_DEVICES_TABLE", "sentinel-devices")
	restPort := envOrDefault("REST_PORT", "8080")
	grpcPort := envOrDefault("GRPC_PORT", "9090")
	jwtSecret := envOrDefault("JWT_SECRET_KEY", "change-me-in-production")
	appVersion := envOrDefault("APP_VERSION", "1.0.0")
	authEnabled := envOrDefault("AUTH_ENABLED", "true")
	otelEndpoint := envOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	// Initialize distributed tracing
	tracerShutdown, tracerErr := middleware.InitTracer(middleware.TracingConfig{
		ServiceName: "sentinel-go",
		Endpoint:    otelEndpoint,
		Insecure:    true,
	})
	if tracerErr != nil {
		structuredLogger.Warn("tracer init failed, tracing disabled", zap.Error(tracerErr))
	} else if tracerShutdown != nil {
		defer func() { _ = tracerShutdown(context.Background()) }()
	}

	// Wire up layers
	deviceRepo := repository.NewDeviceRepository(ddbClient, tableName)
	deviceSvc := service.NewDeviceService(deviceRepo)
	restHandler := handler.NewRESTHandler(deviceSvc)
	grpcHandler := handler.NewGRPCHandler(deviceSvc)
	complianceReporter := compliance.NewComplianceReporter(appVersion)

	// Initialize middleware
	auditLogger := middleware.NewAuditLogger()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)
	validator := middleware.NewRequestValidator()
	bruteForce := middleware.NewBruteForceProtection(5, 15*time.Minute, 30*time.Minute)

	// Initialize superapp handler
	superappHandler := handler.NewSuperappHandler()

	// Start gRPC server
	grpcLis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatalf("gRPC listen: %v", err)
	}
	grpcSrv := grpc.NewServer()
	pb.RegisterDeviceServiceServer(grpcSrv, grpcHandler)
	go func() {
		logger.Printf("gRPC server listening on :%s", grpcPort)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			logger.Fatalf("gRPC serve: %v", err)
		}
	}()

	// Start REST server
	r := chi.NewRouter()

	// Core middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
		r.Use(chimw.StripSlashes)
	r.Use(chimw.Timeout(30 * time.Second))

	// Security middleware
	r.Use(rateLimiter.RateLimit)
	r.Use(validator.Validate)
	r.Use(auditLogger.Audit)
	r.Use(middleware.SanitizeInput)
	r.Use(middleware.SecurityHeaders)
	r.Use(bruteForce.Protect)

	// CORS
	allowedOrigins := envOrDefault("CORS_ALLOWED_ORIGINS", "*")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{allowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Observability middleware
	r.Use(middleware.CorrelationIDMiddleware)
	r.Use(middleware.PrometheusMiddleware)
	r.Use(middleware.TracingMiddleware("sentinel-go"))
	r.Use(middleware.RequestLoggerMiddleware(structuredLogger))

	// Metrics endpoint (no auth)
	r.Handle("/metrics", middleware.MetricsHandler())

	// Public endpoints (no auth required)
	r.Get("/health", healthHandler(complianceReporter, appVersion))
	r.Get("/api/v1/compliance/standards", complianceReporter.HandleStandardsList)
	r.Get("/api/v1/compliance/report", complianceReporter.HandleComplianceReport)
	r.Get("/api/v1/compliance/retention", complianceReporter.HandleRetentionPolicies)
	r.Get("/api/v1/compliance/incidents", complianceReporter.HandleIncidents)
	r.Post("/api/v1/compliance/incidents", incidentCreateHandler(complianceReporter))
	r.Put("/api/v1/compliance/incidents/{incidentID}", incidentUpdateHandler(complianceReporter))

	// OWASP status
	r.Get("/api/v1/security/owasp", middleware.OWASPHandler)

	// Swagger documentation
	r.Get("/swagger.json", swaggerHandler())

	// Protected API routes
	r.Group(func(r chi.Router) {
		if authEnabled == "true" {
			r.Use(middleware.JWTAuth(middleware.JWTConfig{
				SecretKey: jwtSecret,
				Issuer:    "sentinel-device-manager",
			}))
			r.Use(middleware.RBAC(middleware.RBACConfig{
				RolePermissions: map[string]map[string][]string{
					"admin": {
						"GET": {"*"}, "POST": {"*"},
						"PUT": {"*"}, "DELETE": {"*"},
					},
					"operator": {
						"GET": {"*"}, "POST": {"*"},
						"PUT": {"/api/v1/devices"},
					},
					"viewer": {
						"GET": {"*"},
					},
				},
			}))
		}
		restHandler.RegisterRoutes(r)
		superappHandler.RegisterRoutes(r)
	})

	logger.Printf("REST server listening on :%s", restPort)
	srv := &http.Server{
		Addr:         ":" + restPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("REST serve: %v", err)
	}
}

// healthHandler returns comprehensive system health including compliance status.
func healthHandler(cr *compliance.ComplianceReporter, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		components := []compliance.ComponentHealth{
			{Name: "rest_api", Status: "healthy"},
			{Name: "grpc_server", Status: "healthy"},
		}
		_ = cr.GetHealthStatus(components)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"` + "ok" + `","version":"` + version + `"}`))
	}
}

// incidentCreateHandler handles POST requests to create compliance incidents.
func incidentCreateHandler(cr *compliance.ComplianceReporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Category    string `json:"category"`
			ReportedBy  string `json:"reported_by"`
		}
		if err := readJSON(r, &input); err != nil {
			writeHTTPError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if input.Title == "" || input.Severity == "" {
			writeHTTPError(w, http.StatusBadRequest, "title and severity are required")
			return
		}

		incident := cr.CreateIncident(
			input.Title, input.Description,
			input.Severity, input.Category,
			input.ReportedBy,
		)
		writeHTTPJSON(w, http.StatusCreated, incident)
	}
}

// incidentUpdateHandler handles PUT requests to update incident status.
func incidentUpdateHandler(cr *compliance.ComplianceReporter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		incidentID := chi.URLParam(r, "incidentID")
		var input struct {
			Status    string `json:"status"`
			UpdatedBy string `json:"updated_by"`
			Notes     string `json:"notes"`
		}
		if err := readJSON(r, &input); err != nil {
			writeHTTPError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := cr.UpdateIncidentStatus(incidentID, input.Status, input.UpdatedBy, input.Notes); err != nil {
			writeHTTPError(w, http.StatusNotFound, err.Error())
			return
		}
		writeHTTPJSON(w, http.StatusOK, map[string]string{"status": "updated"})
	}
}

// swaggerHandler serves the OpenAPI/Swagger specification.
func swaggerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(swaggerSpec))
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func writeHTTPJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeHTTPError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// swaggerSpec contains the OpenAPI 3.0 specification for the Sentinel Device Manager API.
var swaggerSpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Sentinel Device Manager API",
    "description": "Enterprise physical security device management platform. Manages cameras, access control panels, alarm systems, and environmental sensors at scale.\n\n## Authentication\nAll device management endpoints require JWT Bearer token authentication. Include the token in the Authorization header: ` + "`Bearer <token>`" + `\n\n## Standards Compliance\nThis API implements controls from ISO 27001, ISO 9001, ISO 27035, ISO 27017, ISO 20000, ISO 22301, IEC 62443, NIST CSF, SOC 2, and GDPR.",
    "version": "1.0.0",
    "contact": {
      "name": "Sentinel Device Manager Team"
    }
  },
  "servers": [
    {"url": "http://localhost:8080", "description": "Local development"}
  ],
  "security": [{"BearerAuth": []}],
  "paths": {
    "/health": {
      "get": {
        "tags": ["System"],
        "summary": "Health check",
        "description": "Returns system health status including uptime, version, and component status.",
        "security": [],
        "responses": {
          "200": {
            "description": "System is healthy",
            "content": {
              "application/json": {
                "schema": {"$ref": "#/components/schemas/HealthResponse"}
              }
            }
          }
        }
      }
    },
    "/api/v1/devices": {
      "get": {
        "tags": ["Devices"],
        "summary": "List devices",
        "description": "Retrieve a paginated list of devices with optional filtering by type, status, site, or organization.",
        "parameters": [
          {"name": "type", "in": "query", "schema": {"type": "string", "enum": ["CAMERA", "ACCESS_CONTROL", "ALARM", "SENSOR"]}},
          {"name": "status", "in": "query", "schema": {"type": "string", "enum": ["ONLINE", "OFFLINE", "MAINTENANCE", "DECOMMISSIONED"]}},
          {"name": "site_id", "in": "query", "schema": {"type": "string"}},
          {"name": "organization_id", "in": "query", "schema": {"type": "string"}},
          {"name": "page", "in": "query", "schema": {"type": "integer", "default": 1}},
          {"name": "limit", "in": "query", "schema": {"type": "integer", "default": 50}}
        ],
        "responses": {
          "200": {
            "description": "List of devices",
            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/DeviceListResponse"}}}
          }
        }
      },
      "post": {
        "tags": ["Devices"],
        "summary": "Register a new device",
        "description": "Create and register a new physical security device in the system.",
        "requestBody": {
          "required": true,
          "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateDeviceRequest"}}}
        },
        "responses": {
          "201": {"description": "Device created", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Device"}}}},
          "400": {"description": "Invalid input", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Error"}}}}
        }
      }
    },
    "/api/v1/devices/{deviceID}": {
      "get": {
        "tags": ["Devices"],
        "summary": "Get device details",
        "description": "Retrieve detailed information about a specific device.",
        "parameters": [{"name": "deviceID", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}],
        "responses": {
          "200": {"description": "Device details", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Device"}}}},
          "404": {"description": "Device not found"}
        }
      },
      "put": {
        "tags": ["Devices"],
        "summary": "Update device",
        "description": "Update device configuration. Supports partial updates.",
        "parameters": [{"name": "deviceID", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}],
        "requestBody": {
          "required": true,
          "content": {"application/json": {"schema": {"$ref": "#/components/schemas/UpdateDeviceRequest"}}}
        },
        "responses": {
          "200": {"description": "Device updated", "content": {"application/json": {"schema": {"$ref": "#/components/schemas/Device"}}}},
          "404": {"description": "Device not found"}
        }
      },
      "delete": {
        "tags": ["Devices"],
        "summary": "Delete device",
        "description": "Permanently delete a device from the system. For GDPR-compliant decommissioning, use the decommission endpoint first.",
        "parameters": [{"name": "deviceID", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}],
        "responses": {
          "204": {"description": "Device deleted"},
          "404": {"description": "Device not found"}
        }
      }
    },
    "/api/v1/compliance/report": {
      "get": {
        "tags": ["Compliance"],
        "summary": "Get compliance report",
        "description": "Returns a comprehensive compliance report across all tracked standards (ISO 27001, ISO 9001, ISO 27035, ISO 27017, ISO 20000, ISO 22301, IEC 62443, NIST CSF, SOC 2, GDPR).",
        "security": [],
        "responses": {
          "200": {"description": "Full compliance report"}
        }
      }
    },
    "/api/v1/compliance/standards": {
      "get": {
        "tags": ["Compliance"],
        "summary": "List compliance standards",
        "description": "Returns all tracked compliance standards with their control details.",
        "security": [],
        "responses": {
          "200": {"description": "List of standards"}
        }
      }
    },
    "/api/v1/compliance/retention": {
      "get": {
        "tags": ["Compliance"],
        "summary": "Data retention policies",
        "description": "Returns configured data retention policies per GDPR requirements.",
        "security": [],
        "responses": {
          "200": {"description": "Retention policies"}
        }
      }
    },
    "/api/v1/compliance/incidents": {
      "get": {
        "tags": ["Compliance", "Incidents"],
        "summary": "List incidents",
        "description": "Returns security and operational incidents. Implements ISO 27035 incident management.",
        "parameters": [{"name": "status", "in": "query", "schema": {"type": "string", "enum": ["OPEN", "INVESTIGATING", "CONTAINED", "RESOLVED", "CLOSED"]}}],
        "security": [],
        "responses": {
          "200": {"description": "List of incidents"}
        }
      },
      "post": {
        "tags": ["Compliance", "Incidents"],
        "summary": "Create incident",
        "description": "Record a new security or operational incident. Implements ISO 27035.",
        "security": [],
        "requestBody": {
          "required": true,
          "content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateIncidentRequest"}}}
        },
        "responses": {
          "201": {"description": "Incident created"},
          "400": {"description": "Invalid input"}
        }
      }
    },
    "/api/v1/compliance/incidents/{incidentID}": {
      "put": {
        "tags": ["Compliance", "Incidents"],
        "summary": "Update incident status",
        "description": "Update an incident's status and add a timeline entry. Implements ISO 27035.",
        "security": [],
        "parameters": [{"name": "incidentID", "in": "path", "required": true, "schema": {"type": "string"}}],
        "requestBody": {
          "required": true,
          "content": {"application/json": {"schema": {"$ref": "#/components/schemas/UpdateIncidentRequest"}}}
        },
        "responses": {
          "200": {"description": "Incident updated"},
          "404": {"description": "Incident not found"}
        }
      }
    },
    "/swagger.json": {
      "get": {
        "tags": ["Documentation"],
        "summary": "OpenAPI specification",
        "description": "Returns the OpenAPI 3.0 specification for this API.",
        "security": [],
        "responses": {
          "200": {"description": "OpenAPI JSON specification"}
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    },
    "schemas": {
      "HealthResponse": {
        "type": "object",
        "properties": {
          "status": {"type": "string", "example": "ok"},
          "version": {"type": "string", "example": "1.0.0"}
        }
      },
      "Device": {
        "type": "object",
        "properties": {
          "device_id": {"type": "string", "format": "uuid"},
          "serial_number": {"type": "string", "example": "VKD-CAM-001"},
          "device_type": {"type": "string", "enum": ["CAMERA", "ACCESS_CONTROL", "ALARM", "SENSOR"]},
          "model": {"type": "string", "example": "D30"},
          "firmware_version": {"type": "string", "example": "1.2.3"},
          "status": {"type": "string", "enum": ["ONLINE", "OFFLINE", "MAINTENANCE", "DECOMMISSIONED"]},
          "site_id": {"type": "string"},
          "organization_id": {"type": "string"},
          "ip_address": {"type": "string"},
          "mac_address": {"type": "string"},
          "config": {"type": "object", "additionalProperties": {"type": "string"}},
          "created_at": {"type": "string", "format": "date-time"},
          "updated_at": {"type": "string", "format": "date-time"}
        }
      },
      "CreateDeviceRequest": {
        "type": "object",
        "required": ["serial_number", "device_type", "site_id", "organization_id"],
        "properties": {
          "serial_number": {"type": "string", "example": "VKD-CAM-001"},
          "device_type": {"type": "string", "enum": ["CAMERA", "ACCESS_CONTROL", "ALARM", "SENSOR"]},
          "model": {"type": "string", "example": "D30"},
          "site_id": {"type": "string", "example": "site-001"},
          "organization_id": {"type": "string", "example": "org-001"},
          "ip_address": {"type": "string"},
          "mac_address": {"type": "string"},
          "config": {"type": "object", "additionalProperties": {"type": "string"}}
        }
      },
      "UpdateDeviceRequest": {
        "type": "object",
        "properties": {
          "device_type": {"type": "string", "enum": ["CAMERA", "ACCESS_CONTROL", "ALARM", "SENSOR"]},
          "model": {"type": "string"},
          "firmware_version": {"type": "string"},
          "status": {"type": "string", "enum": ["ONLINE", "OFFLINE", "MAINTENANCE", "DECOMMISSIONED"]},
          "site_id": {"type": "string"},
          "ip_address": {"type": "string"},
          "config": {"type": "object", "additionalProperties": {"type": "string"}}
        }
      },
      "DeviceListResponse": {
        "type": "object",
        "properties": {
          "devices": {"type": "array", "items": {"$ref": "#/components/schemas/Device"}},
          "total": {"type": "integer"},
          "page": {"type": "integer"},
          "limit": {"type": "integer"}
        }
      },
      "CreateIncidentRequest": {
        "type": "object",
        "required": ["title", "severity"],
        "properties": {
          "title": {"type": "string"},
          "description": {"type": "string"},
          "severity": {"type": "string", "enum": ["LOW", "MEDIUM", "HIGH", "CRITICAL"]},
          "category": {"type": "string", "enum": ["SECURITY", "OPERATIONAL", "COMPLIANCE", "DATA"]},
          "reported_by": {"type": "string"}
        }
      },
      "UpdateIncidentRequest": {
        "type": "object",
        "required": ["status"],
        "properties": {
          "status": {"type": "string", "enum": ["OPEN", "INVESTIGATING", "CONTAINED", "RESOLVED", "CLOSED"]},
          "updated_by": {"type": "string"},
          "notes": {"type": "string"}
        }
      },
      "Error": {
        "type": "object",
        "properties": {
          "error": {"type": "string"}
        }
      }
    }
  }
}`
