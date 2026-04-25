# API Specification

## REST API - Go Device Service (Port 8080)

### Authentication

All requests require `Authorization: Bearer <JWT>` header.

### Device Endpoints

#### List Devices
```
GET /api/v1/devices?type={type}&status={status}&site_id={site_id}&page={page}&limit={limit}
```
Response:
```json
{
  "devices": [
    {
      "device_id": "uuid",
      "serial_number": "VKD-CAM-001",
      "device_type": "CAMERA",
      "model": "D30",
      "firmware_version": "1.2.3",
      "status": "ONLINE",
      "site_id": "site-uuid",
      "organization_id": "org-uuid",
      "ip_address": "192.168.1.100",
      "mac_address": "AA:BB:CC:DD:EE:FF",
      "last_heartbeat": "2026-04-24T10:00:00Z",
      "config": {},
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-04-24T10:00:00Z"
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 50
}
```

#### Register Device
```
POST /api/v1/devices
```
Request:
```json
{
  "serial_number": "VKD-CAM-001",
  "device_type": "CAMERA",
  "model": "D30",
  "site_id": "site-uuid",
  "organization_id": "org-uuid",
  "config": {
    "resolution": "4K",
    "retention_days": 30
  }
}
```

#### Get Device
```
GET /api/v1/devices/{device_id}
```

#### Update Device
```
PUT /api/v1/devices/{device_id}
```

#### Delete Device
```
DELETE /api/v1/devices/{device_id}
```

#### Get Device Health
```
GET /api/v1/devices/{device_id}/health
```
Response:
```json
{
  "device_id": "uuid",
  "cpu_usage": 45.2,
  "memory_usage": 62.1,
  "temperature_c": 52.0,
  "uptime_seconds": 864000,
  "network_latency_ms": 12,
  "last_reported": "2026-04-24T10:00:00Z"
}
```

### Firmware Endpoints

#### List Firmware Versions
```
GET /api/v1/firmware/versions?device_type={type}
```

#### Deploy Firmware
```
POST /api/v1/firmware/{version}/deploy
```
Request:
```json
{
  "device_ids": ["uuid1", "uuid2"],
  "staged_rollout": {
    "percentage_per_stage": 25,
    "delay_between_stages_minutes": 30
  }
}
```

#### Get Deployment Status
```
GET /api/v1/firmware/deployments/{deployment_id}
```

---

## REST API - Python Analytics Service (Port 8081)

### Events

#### List Events
```
GET /api/v1/events?device_id={id}&severity={sev}&from={ts}&to={ts}&page={page}&limit={limit}
```

### Alerts

#### List Alerts
```
GET /api/v1/alerts?status={status}&severity={sev}
```

#### Acknowledge Alert
```
POST /api/v1/alerts/{alert_id}/acknowledge
```

### Analytics

#### Get Summary
```
GET /api/v1/analytics/summary?organization_id={id}&period={7d|30d|90d}
```
Response:
```json
{
  "total_devices": 250,
  "online_devices": 240,
  "offline_devices": 10,
  "active_alerts": 5,
  "events_last_24h": 15000,
  "critical_alerts": 1,
  "firmware_compliance_pct": 92.5
}
```

---

## gRPC Service - Device Communication (Port 9090)

```protobuf
service DeviceService {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc SendHeartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc SendEvent(EventRequest) returns (EventResponse);
  rpc StreamFirmware(FirmwareRequest) returns (stream FirmwareChunk);
  rpc ReportFirmwareStatus(FirmwareStatusRequest) returns (FirmwareStatusResponse);
  rpc GetConfig(ConfigRequest) returns (ConfigResponse);
}
```
