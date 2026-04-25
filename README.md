# Sentinel Device Manager

An enterprise physical security device management platform for managing cameras, access control panels, alarm systems, and environmental sensors at scale. Built with a microservices architecture comprising a Go backend (REST + gRPC), Python analytics service (FastAPI), React frontend, DynamoDB storage, and a firmware simulator for embedded devices.

## Architecture

```
+----------------------------------------------------------------------+
|                        Kubernetes Cluster                            |
|                                                                      |
|  +---------------+  +-----------------+  +----------------------+   |
|  |  Frontend     |  |  Go Device Svc  |  |  Python Analytics    |   |
|  |  (React/TS)   |  |  gRPC + REST    |  |  Event Processing    |   |
|  |  Dark Mode     |  |  Port 8080/9090 |  |  Port 8081           |   |
|  +------+--------+  +-------+---------+  +----------+-----------+   |
|         |                   |                       |                |
|         |   +---------------+-----------------------+                |
|         |   |  Security Middleware Layer              |               |
|         |   |  JWT Auth | RBAC | Rate Limit | Audit   |              |
|         |   |  Input Validation | CORS | Security Hdrs|             |
|         |   +----------------------------------------+               |
|         |                                                              |
|         v   v                                                          |
|  +------------------+    +------------------+                         |
|  |  AWS DynamoDB     |    |  AWS SNS/SQS     |                         |
|  |  Device State     |    |  Event Bus       |                         |
|  +------------------+    +------------------+                         |
|                                                                      |
+----------------------------------------------------------------------+

+----------------------------------------------------------------------+
|                     Embedded Devices (Linux)                          |
|                                                                      |
|  +----------+  +----------+  +----------+  +----------+             |
|  |  Camera  |  |  Access   |  |  Alarm   |  |  Sensor  |             |
|  |  FW (Go) |  |  Control  |  |  Panel   |  |  Node    |             |
|  +----------+  +----------+  +----------+  +----------+             |
|       |              |              |              |                  |
|       +--------------+--------------+--------------+                  |
|                             | gRPC / Protobuf                        |
+----------------------------------------------------------------------+
```

The **Security Middleware Layer** sits between the public internet and the application handlers, enforcing authentication, authorization, input validation, rate limiting, and audit logging before any request reaches business logic. It implements controls mapped to ISO 27001, NIST CSF, SOC 2, and IEC 62443.

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend (Go)** | Go 1.22, chi router, gRPC, Protobuf, AWS SDK v2 |
| **Backend (Python)** | Python 3.12, FastAPI, boto3, Pydantic, uvicorn |
| **Frontend** | React 18, TypeScript, TailwindCSS (dark mode) |
| **Database** | AWS DynamoDB (local development via amazon/dynamodb-local) |
| **Containerization** | Docker, Docker Compose, Kubernetes (kubectl manifests) |
| **API Protocols** | gRPC + Protobuf, REST (JSON), OpenAPI 3.0 / Swagger |
| **Embedded Firmware** | Go, HAL abstraction, sensor drivers, firmware simulator |
| **Testing** | Go testing, pytest, Playwright (e2e), integration tests |

## Standards Compliance

The platform tracks compliance with the following international standards and frameworks:

| Standard | Name | Category |
|---|---|---|
| ISO 27001 | ISO/IEC 27001:2022 | Information Security Management (ISMS) |
| ISO 9001 | ISO 9001:2015 | Quality Management System (QMS) |
| ISO 27035 | ISO/IEC 27035:2023 | Information Security Incident Management |
| ISO 27017 | ISO/IEC 27017:2015 | Cloud Security Controls |
| ISO 20000 | ISO/IEC 20000-1:2018 | IT Service Management (ITSM) |
| ISO 22301 | ISO 22301:2019 | Business Continuity Management (BCMS) |
| IEC 62443 | IEC 62443:2018 | Industrial Automation and IoT Security |
| NIST CSF | NIST Cybersecurity Framework 2.0 | Cybersecurity Risk Management |
| SOC 2 | SOC 2 Type II (2017) | Trust Services Criteria (Audit & Assurance) |
| GDPR | EU Regulation 2016/679 | Data Protection and Privacy |

Compliance status is tracked at the individual control level with evidence mapping. Use the `GET /api/v1/compliance/report` endpoint for a full report, or `GET /api/v1/compliance/standards` for per-standard details. See `docs/standards/` for detailed compliance documentation.

## Features

- **Device Lifecycle Management** -- Register, configure, monitor, update, and decommission physical security devices (cameras, access control panels, alarm systems, environmental sensors)
- **JWT Authentication & RBAC Authorization** -- HMAC-SHA256 signed tokens with role-based access control (admin, operator, viewer) and path/method-level enforcement
- **Rate Limiting** -- Per-IP token bucket rate limiting to prevent abuse and denial-of-service attacks
- **Input Validation** -- Request body size limits, content-type enforcement, and parameter validation
- **Audit Logging** -- Full request audit trail with user identity, timestamps, status codes, and request IDs (ISO 27001 A.8.15, SOC 2 CC7.2)
- **Compliance Reporting** -- Real-time compliance status across all 10 tracked standards with evidence mapping
- **Incident Management** -- Security incident tracking with severity classification, timeline entries, and status workflows (ISO 27035)
- **Data Retention Policies** -- Configurable per-category retention periods with GDPR legal basis tracking (GDPR Article 5(1)(e))
- **Firmware Management** -- gRPC server-streaming firmware delivery in 64KB chunks with SHA-256 checksums and deployment status tracking
- **Real-time Analytics Dashboard** -- React dark-mode dashboard with device status, event ingestion, alert management, and analytics summaries
- **Swagger/OpenAPI Documentation** -- Full OpenAPI 3.0 specification served at `GET /swagger.json` with schema definitions and authentication details
- **Comprehensive Test Suite** -- Unit tests (Go, Python), firmware tests, integration tests, and end-to-end tests (Playwright)

## Prerequisites

- Go 1.22+
- Python 3.12+
- Node.js 20+
- Docker & Docker Compose
- protoc (Protocol Buffer compiler)
- AWS CLI (for DynamoDB table creation when running without Docker)

## Running with Docker (Recommended)

1. Copy the environment template for each service:

```bash
cp backend/go/.env.example backend/go/.env
cp backend/python/.env.example backend/python/.env
cp firmware/.env.example firmware/.env
cp frontend/.env.example frontend/.env
```

2. Start all services:

```bash
docker-compose up -d
```

### Services

| Service | Container | Port(s) | Description |
|---|---|---|---|
| DynamoDB Local | `dynamodb` | 8000 | In-memory DynamoDB for local development |
| Go Backend | `go-service` | 8080 (REST), 9090 (gRPC) | Device management REST + gRPC API |
| Python Analytics | `python-service` | 8081 | Event processing and analytics service |
| Frontend | `frontend` | 3000 | React dark-mode dashboard |
| Firmware Simulator | `firmware-sim` | -- | Simulated embedded device (camera) |

### Health Check

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"healthy","version":"1.0.0"}
```

## Running Without Docker

### 1. Start DynamoDB Local

```bash
docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest
```

### 2. Create the DynamoDB Table

```bash
aws dynamodb create-table \
  --table-name sentinel-devices \
  --attribute-definitions AttributeName=device_id,AttributeType=S \
  --key-schema AttributeName=device_id,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --endpoint-url http://localhost:8000 \
  --region us-east-1
```

### 3. Start the Go Backend

```bash
cd backend/go
cp .env.example .env
go run cmd/server/main.go
```

The REST API will be available at `http://localhost:8080` and gRPC at `localhost:9090`.

### 4. Start the Python Analytics Service

```bash
cd backend/python
cp .env.example .env
pip install -r analytics_service/requirements.txt
uvicorn analytics_service.main:app --port 8081
```

The analytics API will be available at `http://localhost:8081`.

### 5. Start the Frontend

```bash
cd frontend
cp .env.example .env
npm install
npm start
```

The dashboard will be available at `http://localhost:3000`.

### 6. (Optional) Run the Firmware Simulator

```bash
cd firmware
cp .env.example .env
go run cmd/firmware-sim/main.go
```

The simulator registers a virtual camera device and begins sending heartbeats and events to the Go backend.

## Environment Variables

Each service has its own `.env.example` file with documented defaults. Copy the appropriate file to `.env` and adjust values for your environment.

### Go Backend (`backend/go/.env.example`)

| Variable | Default | Description |
|---|---|---|
| `AWS_REGION` | `us-east-1` | AWS region for DynamoDB |
| `AWS_ACCESS_KEY_ID` | `test` | AWS access key (use `test` for DynamoDB Local) |
| `AWS_SECRET_ACCESS_KEY` | `test` | AWS secret key (use `test` for DynamoDB Local) |
| `DYNAMO_DEVICES_TABLE` | `sentinel-devices` | DynamoDB table name for device records |
| `DYNAMODB_ENDPOINT` | `http://localhost:8000` | DynamoDB endpoint override (leave empty for AWS-managed) |
| `REST_PORT` | `8080` | REST API listen port |
| `GRPC_PORT` | `9090` | gRPC server listen port |
| `AUTH_ENABLED` | `true` | Enable/disable JWT authentication (`false` for development) |
| `JWT_SECRET_KEY` | `change-me-in-production` | HMAC-SHA256 secret for JWT signing (change in production) |
| `APP_VERSION` | `1.0.0` | Application version reported in health checks |
| `CORS_ALLOWED_ORIGINS` | `*` | Comma-separated allowed origins, or `*` for all |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

### Python Analytics (`backend/python/.env.example`)

| Variable | Default | Description |
|---|---|---|
| `AWS_REGION` | `us-east-1` | AWS region for DynamoDB |
| `AWS_ACCESS_KEY_ID` | `test` | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | `test` | AWS secret key |
| `DYNAMODB_ENDPOINT` | `http://localhost:8000` | DynamoDB endpoint override |
| `PORT` | `8081` | Analytics service listen port |

### Firmware Simulator (`firmware/.env.example`)

| Variable | Default | Description |
|---|---|---|
| `DEVICE_SERIAL` | `VKD-CAM-SIM-001` | Simulated device serial number |
| `DEVICE_MAC` | `AA:BB:CC:DD:EE:FF` | Simulated device MAC address |
| `DEVICE_MODEL` | `D30-SIM` | Simulated device hardware model |
| `API_URL` | `http://localhost:8080` | Go backend API base URL |

### Frontend (`frontend/.env.example`)

| Variable | Default | Description |
|---|---|---|
| `REACT_APP_API_BASE` | `http://localhost:8080` | Go backend API base URL |
| `REACT_APP_ANALYTICS_BASE` | `http://localhost:8081` | Python analytics service base URL |

## API Reference

### Swagger / OpenAPI

The full OpenAPI 3.0 specification is available at:

```
GET /swagger.json
```

This includes all endpoint definitions, request/response schemas, authentication requirements, and compliance metadata.

### REST Endpoints -- Go Service (Port 8080)

#### Device Management (requires JWT authentication)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/devices` | List devices with filtering and pagination |
| `POST` | `/api/v1/devices` | Register a new device |
| `GET` | `/api/v1/devices/{deviceID}` | Get device details by ID |
| `PUT` | `/api/v1/devices/{deviceID}` | Update device configuration (partial) |
| `DELETE` | `/api/v1/devices/{deviceID}` | Delete / decommission a device |

Query parameters for `GET /api/v1/devices`: `type`, `status`, `site_id`, `organization_id`, `page`, `limit`.

#### Compliance Endpoints (public, no authentication required)

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | System health check with component status |
| `GET` | `/api/v1/compliance/report` | Full compliance report across all standards |
| `GET` | `/api/v1/compliance/standards` | List all tracked compliance standards and controls |
| `GET` | `/api/v1/compliance/retention` | Data retention policies (GDPR Article 5) |
| `GET` | `/api/v1/compliance/incidents` | List security incidents (ISO 27035) |
| `POST` | `/api/v1/compliance/incidents` | Create a new security incident |
| `PUT` | `/api/v1/compliance/incidents/{incidentID}` | Update incident status with timeline entry |

### REST Endpoints -- Python Service (Port 8081)

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/v1/events` | List security events (filterable by `device_id`, `severity`, `limit`) |
| `POST` | `/api/v1/events` | Ingest a security event (auto-generates alerts for CRITICAL severity) |
| `GET` | `/api/v1/alerts` | List active alerts (filterable by `status`, `severity`, `limit`) |
| `POST` | `/api/v1/alerts/{alert_id}/acknowledge` | Acknowledge an alert |
| `GET` | `/api/v1/analytics/summary` | Get analytics summary |

### gRPC Methods -- Go Service (Port 9090)

| Method | Request | Response | Description |
|---|---|---|---|
| `Register` | `RegisterRequest` | `RegisterResponse` | Enroll a new device into the platform |
| `SendHeartbeat` | `HeartbeatRequest` | `HeartbeatResponse` | Receive periodic health updates from devices |
| `SendEvent` | `EventRequest` | `EventResponse` | Ingest a security or telemetry event |
| `StreamFirmware` | `FirmwareRequest` | `stream FirmwareChunk` | Stream firmware binary to a device in 64KB chunks |
| `ReportFirmwareStatus` | `FirmwareStatusRequest` | `FirmwareStatusResponse` | Receive firmware update status from a device |
| `GetConfig` | `ConfigRequest` | `ConfigResponse` | Retrieve current device configuration |

Service and message definitions are in `proto/device/device.proto`.

## Testing

```bash
# Go unit tests (backend)
cd backend/go
go test ./...

# Python tests (analytics service)
cd backend/python
pytest analytics_service/tests/ -v

# Firmware tests
cd firmware
go test ./...

# Integration tests (cross-service)
cd tests/integration
go test ./...

# End-to-end tests (Playwright)
cd tests/e2e
go test ./...
```

## Project Structure

```
sentinel-device-manager/
├── docs/                       # Architecture, API specs, compliance documentation
│   ├── architecture.md         # Architecture decision records
│   ├── api-specification.md    # Detailed API specification
│   ├── requirements.md         # System requirements document
│   └── standards/              # Compliance documentation per standard
├── proto/                      # Protobuf service definitions
│   └── device/
│       └── device.proto        # DeviceService gRPC definition
├── backend/
│   ├── go/                     # Go device management service
│   │   ├── cmd/server/         # Entry point (main.go)
│   │   ├── internal/
│   │   │   ├── handler/        # REST and gRPC request handlers
│   │   │   ├── middleware/      # JWT auth, RBAC, audit logging, rate limiting
│   │   │   ├── compliance/     # Standards tracking and reporting
│   │   │   ├── model/          # Domain models (Device, Event, etc.)
│   │   │   ├── repository/     # DynamoDB data access layer
│   │   │   └── service/        # Business logic layer
│   │   ├── pkg/api/            # Generated protobuf/gRPC Go code
│   │   └── .env.example        # Go service environment template
│   └── python/                 # Python analytics service
│       ├── analytics_service/
│       │   ├── main.py         # FastAPI application entry point
│       │   ├── handlers.py     # Route handlers
│       │   ├── models.py       # Pydantic models
│       │   ├── store.py        # In-memory event/alert store
│       │   └── requirements.txt
│       └── .env.example        # Python service environment template
├── firmware/                   # Embedded device firmware (Go)
│   ├── cmd/firmware-sim/       # Firmware simulator entry point
│   ├── internal/               # HAL, sensors, network layers
│   └── .env.example            # Firmware simulator environment template
├── frontend/                   # React dark-mode dashboard
│   ├── src/
│   │   ├── components/         # UI components (Dashboard, DeviceList, AlertViewer, etc.)
│   │   ├── pages/              # Page views
│   │   ├── services/           # API client (api.ts)
│   │   ├── styles/             # Dark theme CSS
│   │   ├── App.tsx             # Application root
│   │   └── index.tsx           # Entry point
│   └── .env.example            # Frontend environment template
├── deploy/                     # Docker & Kubernetes configurations
│   └── docker/                 # Dockerfiles (Go, Python, Frontend, Firmware)
├── tests/
│   ├── e2e/                    # End-to-end tests (Playwright)
│   └── integration/            # Cross-service integration tests
└── docker-compose.yml          # Local development stack
```

## Documentation

| Resource | Location |
|---|---|
| Compliance Documentation | `docs/standards/` -- per-standard control details and evidence |
| Architecture Decisions | `docs/architecture.md` |
| API Specification | `docs/api-specification.md` |
| System Requirements | `docs/requirements.md` |
| OpenAPI / Swagger | `GET /swagger.json` on the Go backend (port 8080) |
| Go Package Documentation | Run `go doc ./...` from `backend/go/` |

## Security

The Sentinel Device Manager implements a layered security model across the middleware stack:

- **Authentication** -- HMAC-SHA256 signed JWT tokens (HS256) with expiration enforcement and Bearer token extraction from the `Authorization` header
- **Authorization** -- Role-based access control (RBAC) with three roles (admin, operator, viewer) and path/method-level enforcement per role
- **Rate Limiting** -- Per-IP token bucket rate limiting (default: 100 requests/minute) with automatic cleanup and `Retry-After` headers
- **Input Validation** -- Request body size limits (1MB maximum), content-type enforcement, and required field validation
- **Audit Logging** -- Full request audit trail capturing method, path, status code, user identity, IP address, user agent, request ID, and response duration
- **Security Headers** -- `Strict-Transport-Security`, `Content-Security-Policy`, `X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, and `Referrer-Policy` set on every response
- **CORS** -- Configurable allowed origins with credentials support and preflight handling
- **Data Retention** -- Per-category retention policies with GDPR legal basis tracking and automatic policy enforcement

## License

Proprietary -- Internal use only.
