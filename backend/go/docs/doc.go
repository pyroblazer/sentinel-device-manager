// Sentinel Device Manager - Go Backend Documentation
//
// This package contains the Go backend for the Sentinel Device Manager platform.
//
// # Architecture Overview
//
// The backend follows a layered architecture:
//
//   handler (transport) -> service (business logic) -> repository (data access)
//
// # Packages
//
//   - cmd/server: Application entry point and server configuration
//   - internal/handler: REST and gRPC request handlers
//   - internal/service: Business logic layer
//   - internal/repository: Data persistence (DynamoDB)
//   - internal/model: Domain entities
//   - internal/middleware: Security middleware (JWT, RBAC, audit, rate limit)
//   - internal/compliance: Standards compliance tracking and reporting
//   - pkg/api: Protobuf/gRPC generated code
//
// # API Documentation
//
// Access the OpenAPI/Swagger specification at: GET /swagger.json
//
// # Running
//
//	export DYNAMODB_ENDPOINT=http://localhost:8000
//	export JWT_SECRET_KEY=your-secret-key
//	go run cmd/server/main.go
//
// # Standards Compliance
//
// See the compliance package (internal/compliance) for the full list of
// tracked standards and controls. The compliance report endpoint provides
// a machine-readable overview: GET /api/v1/compliance/report
package docs
