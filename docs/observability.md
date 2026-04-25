# Sentinel Device Manager - Observability Architecture

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Client / Browser                            │
└──────────────┬──────────────────────────────┬───────────────────────┘
               │ HTTP                         │ HTTP
               ▼                              ▼
┌──────────────────────┐          ┌──────────────────────┐
│   Go Backend :8080   │          │ Python Service :8081 │
│  ┌────────────────┐  │          │  ┌────────────────┐  │
│  │  REST + gRPC   │  │          │  │    FastAPI      │  │
│  │  ┌──────────┐  │  │          │  │  ┌──────────┐  │  │
│  │  │Prometheus│  │  │          │  │  │Prometheus│  │  │
│  │  │Middleware│  │  │          │  │  │Middleware│  │  │
│  │  └────┬─────┘  │  │          │  │  └────┬─────┘  │  │
│  │  ┌────┴─────┐  │  │          │  │  ┌────┴─────┐  │  │
│  │  │OTel Trace│  │  │          │  │  │OTel Trace│  │  │
│  │  │Middleware│  │  │          │  │  │Middleware│  │  │
│  │  └────┬─────┘  │  │          │  │  └────┬─────┘  │  │
│  │  ┌────┴─────┐  │  │          │  │  ┌────┴─────┐  │  │
│  │  │ Struct   │  │  │          │  │  │ Struct   │  │  │
│  │  │ Logging  │  │  │          │  │  │ Logging  │  │  │
│  │  └────┬─────┘  │  │          │  │  └────┬─────┘  │  │
│  └───────┼────────┘  │          │  └───────┼────────┘  │
│          │           │          │          │           │
│  /metrics│  JSON logs│          │  /metrics│  JSON logs│
└──────────┼───────────┘          └──────────┼───────────┘
           │                                 │
     ┌─────┴─────┐                     ┌─────┴─────┐
     │  Scrape   │                     │  Scrape   │
     ▼           │                     ▼           │
┌─────────┐     │                  ┌─────────┐    │
│Prometheus│     │                  │Prometheus│    │
│  :9091   │     │                  │  :9091   │    │
└────┬─────┘     │                  └────┬─────┘    │
     │           │                       │          │
     │ Query     │                       │ Query    │
     ▼           │                       ▼          │
┌─────────┐     │                  ┌─────────┐     │
│ Grafana │     │                  │ Grafana │     │
│  :3001  │     │                  │  :3001  │     │
└─────────┘     │                  └─────────┘     │
                │                                   │
                │  Traces (OTLP gRPC)               │  Traces (OTLP gRPC)
                ▼                                   ▼
          ┌──────────────────────────────────────────┐
          │        OpenTelemetry Collector           │
          │   :4317 (gRPC)  :4318 (HTTP)  :8889     │
          └─────┬──────────────────┬────────────────┘
                │                  │
                │ Traces           │ Logs
                ▼                  ▼
          ┌───────────┐    ┌──────────────┐
          │Elastic-   │    │ Logstash     │◄──── Filebeat
          │search     │    │  :5044       │      (Docker logs)
          │  :9200    │    └──────┬───────┘
          └─────┬─────┘           │
                │                 │
                ▼                 ▼
          ┌───────────┐    ┌──────────────┐
          │  Kibana   │    │Elasticsearch │
          │  :5601    │    │  indices     │
          └───────────┘    └──────────────┘
                │
                ▼
          ┌───────────┐
          │Alert-     │
          │manager    │
          │  :9093    │
          └───────────┘
```

## Data Flow

### Metrics Pipeline
1. Application exposes `/metrics` endpoint with Prometheus-formatted metrics
2. Prometheus scrapes metrics every 15 seconds from Go and Python services
3. Grafana queries Prometheus for dashboard visualization
4. Alertmanager evaluates alert rules and sends notifications

### Logging Pipeline
1. Applications write structured JSON logs to stdout (zap / structlog)
2. Filebeat collects Docker container logs
3. Logstash parses JSON, extracts correlation_id and trace_id
4. Elasticsearch stores logs in daily indices (`sentinel-logs-YYYY.MM.dd`)
5. Kibana provides search, filtering, and visualization

### Tracing Pipeline
1. OpenTelemetry SDK creates spans for each HTTP/gRPC request
2. Spans exported via OTLP gRPC to the Collector
3. Collector batches and forwards traces to Elasticsearch
4. Trace IDs correlate with logs and metrics

## Metric Naming Convention

All custom metrics follow `sentinel_<subsystem>_<name>_<unit>`:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `sentinel_http_requests_total` | Counter | method, path, status_code | Total HTTP requests |
| `sentinel_http_request_duration_seconds` | Histogram | method, path | Request latency |
| `sentinel_http_requests_in_flight` | Gauge | method | Active requests |
| `sentinel_http_response_size_bytes` | Histogram | method, path | Response size |
| `sentinel_db_query_duration_seconds` | Histogram | operation | DB query latency |
| `sentinel_db_errors_total` | Counter | operation | DB errors |
| `sentinel_grpc_requests_total` | Counter | method, status_code | gRPC requests |
| `sentinel_grpc_request_duration_seconds` | Histogram | method | gRPC latency |
| `sentinel_devices_managed_total` | Gauge | - | Device count |
| `sentinel_compliance_checks_total` | Counter | standard, status | Compliance checks |
| `sentinel_incidents_created_total` | Counter | severity | Incidents |

Python service adds `sentinel_python_` prefixed metrics for events, alerts, and analytics queries.

## Alert Rules

| Alert | Condition | Severity | Window |
|-------|-----------|----------|--------|
| HighErrorRate | 5xx rate > 5% | Critical | 5m |
| HighLatencyP95 | P95 latency > 2s | Warning | 5m |
| HighLatencyP99 | P99 latency > 5s | Critical | 5m |
| HighRequestRate | RPS > 1000 | Warning | 3m |
| HighMemoryUsage | RSS > 1GB | Warning | 5m |
| HighGoroutines | Goroutines > 1000 | Warning | 5m |
| ServiceDown | up == 0 | Critical | 1m |
| HighDBLatency | DB P95 > 500ms | Warning | 3m |
| DBErrorRateSpike | DB error rate > 10% | Critical | 2m |
| IncidentSpike | >10 incidents in 5m | Warning | 5m |
| ComplianceFailure | >5 failures in 1h | Warning | 1h |

## Grafana Dashboards

### System Overview (`sentinel-system-overview`)
Go runtime metrics (goroutines, GC, memory, CPU, FDs) + HTTP/gRPC overview.

### API Performance (`sentinel-api-performance`)
Request rates by method/path/status, latency percentiles (P50/P95/P99), throughput,
database performance, business metrics (devices, compliance, incidents).

### Error Tracking (`sentinel-error-tracking`)
Error rates, 5xx distribution, top error endpoints, database errors, gRPC errors,
correlated Elasticsearch logs.

## Correlation

Every request carries a **correlation ID** (X-Correlation-ID header) that links:
- Structured log entries (correlation_id field)
- OpenTelemetry traces (trace_id attribute)
- Prometheus exemplars (trace_id label)

This enables jumping from a Grafana metric spike to the exact log entries and traces
that explain the anomaly.

## Quick Start

```bash
# Start everything (app + observability)
make obs-up

# Start just the app
make dev

# Access dashboards
make obs-grafana    # http://localhost:3001
make obs-prometheus # http://localhost:9091
make obs-kibana     # http://localhost:5601
```

## Log Format

JSON logs follow this structure:

```json
{
  "timestamp": "2026-04-25T10:30:00.000Z",
  "level": "info",
  "caller": "middleware/structured_logging.go:95",
  "message": "http_request_completed",
  "service": "sentinel-go",
  "method": "GET",
  "path": "/api/v1/devices",
  "remote_addr": "10.0.0.1:54321",
  "user_agent": "Mozilla/5.0",
  "correlation_id": "1714041000123456789",
  "trace_id": "abc123def456",
  "status_code": 200,
  "duration": 45.123,
  "response_size_bytes": 1024
}
```

## Index Strategy

Elasticsearch indices use daily rotation:
- `sentinel-logs-2026.04.25` - Application logs
- `sentinel-traces` - Distributed traces
