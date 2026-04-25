# Sentinel Device Manager - Requirements Specification

## 1. Functional Requirements

### 1.1 Device Management

| ID | Requirement | Priority |
|---|---|---|
| FR-001 | Register new devices (camera, access control, alarm, sensor) | Must |
| FR-002 | List devices with filtering by type, status, site, organization | Must |
| FR-003 | View device details including hardware info, firmware version, connectivity status | Must |
| FR-004 | Update device configuration remotely | Must |
| FR-005 | Decommission devices with graceful shutdown | Must |
| FR-006 | Monitor device health (CPU, memory, temperature, connectivity) | Must |
| FR-007 | Group devices by site, building, floor, zone | Should |
| FR-008 | Bulk operations on device groups | Should |

### 1.2 Firmware Management

| ID | Requirement | Priority |
|---|---|---|
| FW-001 | List available firmware versions per device type | Must |
| FW-002 | Deploy firmware to individual devices | Must |
| FW-003 | Rollout firmware to device groups with staged deployment | Must |
| FW-004 | Rollback firmware to previous version | Must |
| FW-005 | Stream firmware binary to devices via gRPC | Must |
| FW-006 | Track firmware deployment status per device | Should |

### 1.3 Event Processing & Alerting

| ID | Requirement | Priority |
|---|---|---|
| EV-001 | Ingest events from all device types | Must |
| EV-002 | Classify events by severity (info, warning, critical) | Must |
| EV-003 | Generate alerts from critical events | Must |
| EV-004 | Acknowledge and resolve alerts | Must |
| EV-005 | Provide event analytics and trend data | Should |
| EV-006 | Real-time event streaming via SSE or WebSocket | Could |

### 1.4 Analytics

| ID | Requirement | Priority |
|---|---|---|
| AN-001 | Dashboard with device status summary | Must |
| AN-002 | Alert volume trends over time | Should |
| AN-003 | Device health score aggregation | Should |
| AN-004 | Firmware compliance reporting | Should |

## 2. Non-Functional Requirements

### 2.1 Performance

| ID | Requirement | Target |
|---|---|---|
| NFR-001 | API response time (p99) | < 200ms |
| NFR-002 | Concurrent device connections | 10,000+ |
| NFR-003 | Event ingestion throughput | 50,000 events/sec |
| NFR-004 | Firmware deployment to 1,000 devices | < 30 min |

### 2.2 Reliability

| ID | Requirement | Target |
|---|---|---|
| NFR-005 | Service uptime | 99.95% |
| NFR-006 | Data durability | 99.9999% |
| NFR-007 | Graceful degradation on component failure | Must |
| NFR-008 | Automatic device reconnection | < 30 sec |

### 2.3 Security

| ID | Requirement | Target |
|---|---|---|
| NFR-009 | TLS for all communications | Must |
| NFR-010 | API authentication via JWT | Must |
| NFR-011 | Role-based access control | Must |
| NFR-012 | Audit logging for all management actions | Must |
| NFR-013 | Firmware binary integrity verification (SHA-256) | Must |

### 2.4 Scalability

| ID | Requirement | Target |
|---|---|---|
| NFR-014 | Horizontal scaling of all services | Must |
| NFR-015 | Multi-region deployment support | Should |
| NFR-016 | Support 100,000+ devices per organization | Should |

## 3. Technical Constraints

- **Languages**: Go (backend/firmware), Python (analytics), TypeScript (frontend)
- **Protocols**: REST (JSON), gRPC (Protobuf)
- **Database**: DynamoDB (NoSQL)
- **Cloud**: AWS (primary)
- **Containers**: Docker, orchestrated via Kubernetes
- **Embedded**: Linux-based devices, Go firmware with HAL abstraction

## 4. Data Model

### Device

```
Device {
  device_id:        string (UUID)
  serial_number:    string
  device_type:      enum [CAMERA, ACCESS_CONTROL, ALARM, SENSOR]
  model:            string
  firmware_version: string
  status:           enum [ONLINE, OFFLINE, MAINTENANCE, DECOMMISSIONED]
  site_id:          string
  organization_id:  string
  ip_address:       string
  mac_address:      string
  last_heartbeat:   timestamp
  config:           JSON (device-type-specific configuration)
  created_at:       timestamp
  updated_at:       timestamp
}
```

### Event

```
Event {
  event_id:     string (UUID)
  device_id:    string
  event_type:   string
  severity:     enum [INFO, WARNING, CRITICAL]
  payload:      JSON
  timestamp:    timestamp
}
```

### Alert

```
Alert {
  alert_id:       string (UUID)
  device_id:      string
  event_id:       string
  alert_type:     string
  severity:       enum [WARNING, CRITICAL]
  status:         enum [ACTIVE, ACKNOWLEDGED, RESOLVED]
  message:        string
  acknowledged_by: string (user_id)
  created_at:     timestamp
  updated_at:     timestamp
}
```

### Firmware

```
Firmware {
  version:       string (semver)
  device_type:   enum [CAMERA, ACCESS_CONTROL, ALARM, SENSOR]
  binary_url:    string (S3)
  checksum_sha256: string
  release_notes: string
  created_at:    timestamp
}
```
