# Architecture Design

## System Overview

Sentinel Device Manager follows a microservices architecture with clear domain boundaries:

1. **Go Device Service** - Core device lifecycle management, firmware deployment
2. **Python Analytics Service** - Event processing, alerting, analytics
3. **React Frontend** - Operator dashboard with high-contrast dark mode
4. **Embedded Firmware** - Go-based firmware for Linux-embedded devices

## Component Diagram

```
                            ┌─────────────────────────────┐
                            │        React Frontend        │
                            │   (High-Contrast Dark Mode)  │
                            └──────────┬──────────────────┘
                                       │ HTTPS
                    ┌──────────────────┴──────────────────┐
                    │                                      │
          ┌─────────▼──────────┐              ┌───────────▼─────────┐
          │   Go Device Service │              │ Python Analytics    │
          │                     │              │ Service             │
          │  ┌───────────────┐ │              │                     │
          │  │ REST Handler  │ │              │  ┌───────────────┐  │
          │  │ (chi router)  │ │              │  │ FastAPI       │  │
          │  └───────┬───────┘ │              │  │ Routes        │  │
          │          │         │              │  └───────┬───────┘  │
          │  ┌───────▼───────┐ │              │          │          │
          │  │ gRPC Handler  │ │              │  ┌───────▼───────┐  │
          │  │ (protobuf)    │ │              │  │ Event         │  │
          │  └───────┬───────┘ │              │  │ Processor     │  │
          │          │         │              │  └───────┬───────┘  │
          │  ┌───────▼───────┐ │              │          │          │
          │  │ Device Service│◄├──────────────┤  ┌───────▼───────┐  │
          │  │ (business     │ │  gRPC        │  │ Alert Engine  │  │
          │  │  logic)       │ │  internal    │  │               │  │
          │  └───────┬───────┘ │              │  └───────────────┘  │
          │          │         │              │                     │
          │  ┌───────▼───────┐ │              │  ┌───────────────┐  │
          │  │ DynamoDB      │ │              │  │ DynamoDB      │  │
          │  │ Repository    │ │              │  │ Repository    │  │
          │  └───────────────┘ │              │  └───────────────┘  │
          └────────────────────┘              └─────────────────────┘
                    │                                      │
                    └──────────────┬───────────────────────┘
                                   │ AWS SDK
                          ┌────────▼────────┐
                          │   DynamoDB       │
                          │                  │
                          │ Tables:          │
                          │  - Devices       │
                          │  - Events        │
                          │  - Alerts        │
                          │  - Firmware      │
                          └─────────────────┘

          ┌──────────────────────────────────────────────┐
          │           Embedded Devices (Linux)            │
          │                                              │
          │  ┌─────────────────────────────────────────┐ │
          │  │        Go Firmware Runtime              │ │
          │  │                                         │ │
          │  │  ┌──────┐  ┌──────┐  ┌──────────────┐  │ │
          │  │  │ HAL  │  │Sensor│  │ gRPC Client  │  │ │
          │  │  │ Layer│  │Drivers│ │ (device →    │  │ │
          │  │  └──────┘  └──────┘  │  cloud comm) │  │ │
          │  │           ┌──────┐   └──────────────┘  │ │
          │  │           │Network│                     │ │
          │  │           │Stack  │                     │ │
          │  │           └──────┘                     │ │
          │  └─────────────────────────────────────────┘ │
          └──────────────────────────────────────────────┘
```

## Data Flow

### Device Registration
```
Device → gRPC Register() → Device Service → DynamoDB PUT → Response
```

### Event Ingestion
```
Device → gRPC SendEvent() → Device Service → SQS Queue → Python Analytics
                                                              │
                                                              ├── Store Event
                                                              ├── Evaluate Alert Rules
                                                              └── Generate Alert (if matched)
```

### Firmware Deployment
```
Operator → REST POST /firmware/deploy → Device Service → gRPC StreamFirmware() → Device
                                                                         │
                                                                  Device verifies checksum
                                                                  Device applies update
                                                                  Device reports status
```

## API Design

### Dual Protocol Strategy

- **REST (JSON)**: Used by frontend and external integrations. Human-readable, wide tooling support.
- **gRPC (Protobuf)**: Used by embedded devices and inter-service communication. Efficient binary serialization, streaming support.

### Versioning

All REST endpoints prefixed with `/api/v1/`. gRPC services use package versioning in protobuf definitions.

## Deployment Architecture

```
                    ┌─────────────── Internet ──────────────┐
                    │                                       │
               ┌────▼─────┐                          ┌─────▼────┐
               │ CDN/S3   │                          │ API GW   │
               │ (Frontend)│                          │ (ALB)    │
               └──────────┘                          └────┬─────┘
                                                         │
                                          ┌──────────────┼──────────────┐
                                          │              │              │
                                    ┌─────▼─────┐ ┌─────▼─────┐ ┌────▼────┐
                                    │ K8s Pod    │ │ K8s Pod    │ │ K8s Pod │
                                    │ Go Svc x3  │ │ Python x2  │ │ Frontend│
                                    └─────┬──────┘ └─────┬─────┘ └─────────┘
                                          │              │
                                    ┌─────▼──────────────▼─────┐
                                    │       DynamoDB            │
                                    └──────────────────────────┘
```

## Security Architecture

1. **TLS Everywhere**: All service-to-service and device-to-cloud communication encrypted
2. **JWT Authentication**: Frontend authenticates via JWT tokens
3. **mTLS for Devices**: Embedded devices use mutual TLS for strong identity
4. **RBAC**: Role-based access control (Admin, Operator, Viewer)
5. **Audit Trail**: All management operations logged with actor, action, timestamp
6. **Firmware Signing**: Firmware binaries signed and verified via SHA-256 checksum
