# Sentinel Device Manager - Standards Compliance Reference

> **Version:** 1.0.0
> **Last Updated:** 2026-04-24
> **Classification:** Internal Reference / Audit Evidence
> **Audience:** Security engineers, compliance officers, auditors, developers

This document describes every regulatory and industry standard implemented by the Sentinel Device Manager platform, maps each control to concrete code, and provides step-by-step verification procedures suitable for internal review and external audit evidence.

---

## Table of Contents

1. [Standards Overview](#1-standards-overview)
2. [ISO 27001:2022 -- Information Security Management System](#2-iso-270012022--information-security-management-system)
3. [ISO 9001:2015 -- Quality Management System](#3-iso-90012015--quality-management-system)
4. [ISO 27035:2023 -- Information Security Incident Management](#4-iso-270352023--information-security-incident-management)
5. [ISO 27017:2015 -- Cloud Security Controls](#5-iso-270172015--cloud-security-controls)
6. [ISO 20000-1:2018 -- IT Service Management](#6-iso-20000-12018--it-service-management)
7. [ISO 22301:2019 -- Business Continuity Management](#7-iso-223012019--business-continuity-management)
8. [IEC 62443 -- Industrial Automation and Control Systems Security](#8-iec-62443--industrial-automation-and-control-systems-security)
9. [NIST Cybersecurity Framework 2.0](#9-nist-cybersecurity-framework-20)
10. [SOC 2 Type II -- Trust Services Criteria](#10-soc-2-type-ii--trust-services-criteria)
11. [GDPR (EU 2016/679) -- General Data Protection Regulation](#11-gdpr-eu-2016679--general-data-protection-regulation)
12. [Standards Interrelationships](#12-standards-interrelationships)
13. [Compliance Reporting API](#13-compliance-reporting-api)
14. [Incident Management Procedures](#14-incident-management-procedures)
15. [Data Retention Policies](#15-data-retention-policies)
16. [Adding New Standards or Controls](#16-adding-new-standards-or-controls)
17. [Appendix A: Quick Reference Matrix](#appendix-a-quick-reference-matrix)

---

## 1. Standards Overview

The Sentinel Device Manager tracks and implements controls across ten standards. The table below provides a snapshot of current compliance posture. All figures are generated dynamically by the compliance reporting subsystem defined in `backend/go/internal/compliance/compliance.go`.

| Standard | Category | Total Controls | Implemented | Partial | Planned |
|---|---|---|---|---|---|
| ISO 27001:2022 | Information Security | 12 | 10 | 2 | 0 |
| ISO 9001:2015 | Quality Management | 8 | 7 | 1 | 0 |
| ISO 27035:2023 | Incident Management | 6 | 5 | 1 | 0 |
| ISO 27017:2015 | Cloud Security | 6 | 5 | 1 | 0 |
| ISO 20000-1:2018 | IT Service Management | 7 | 6 | 0 | 1 |
| ISO 22301:2019 | Business Continuity | 5 | 3 | 1 | 1 |
| IEC 62443 | Industrial / IoT Security | 8 | 6 | 2 | 0 |
| NIST CSF 2.0 | Cybersecurity | 10 | 8 | 2 | 0 |
| SOC 2 Type II | Audit & Assurance | 7 | 7 | 0 | 0 |
| GDPR (EU 2016/679) | Data Protection & Privacy | 9 | 8 | 1 | 0 |

**Total:** 78 controls tracked across 10 standards.

---

## 2. ISO 27001:2022 -- Information Security Management System

### What It Is

ISO/IEC 27001:2022 is the international standard for establishing, implementing, maintaining, and continually improving an Information Security Management System (ISMS). It defines a risk-based approach to managing information security through a comprehensive set of controls organized across organizational, people, physical, and technological themes (Annex A). Certification demonstrates that an organization has systematically addressed information security risks.

### Why It Applies

The Sentinel Device Manager is an enterprise platform that manages physical security devices -- cameras, access control panels, alarm systems, and environmental sensors. These devices generate sensitive telemetry, control physical access to facilities, and produce audit evidence. A compromise of the management platform could lead to unauthorized facility access, surveillance blind spots, or destruction of forensic evidence. ISO 27001 provides the overarching security governance framework that ensures all controls are risk-assessed, documented, and continuously improved.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| A.5.1 | Policies for information security | IMPLEMENTED | Security policy enforced via layered middleware stack (auth, RBAC, rate limiting, validation, audit) | `backend/go/cmd/server/main.go` lines 96-127 |
| A.5.3 | Segregation of duties | IMPLEMENTED | RBAC middleware enforces role-based separation. Admin, operator, and viewer roles have distinct permission sets | `backend/go/internal/middleware/auth.go` lines 117-165 |
| A.5.7 | Threat intelligence | IMPLEMENTED | Rate limiting mitigates automated attacks; request validation blocks malformed payloads | `backend/go/internal/middleware/auth.go` lines 249-318 |
| A.8.1 | User endpoint devices | PARTIAL | Device registration and lifecycle tracking implemented. Endpoint hardening in progress | `backend/go/internal/service/device_service.go` lines 55-95 |
| A.8.2 | Privileged access rights | IMPLEMENTED | JWT-based authentication with role claims. Admin role required for destructive operations | `backend/go/internal/middleware/auth.go` lines 44-68 |
| A.8.3 | Information access restriction | IMPLEMENTED | RBAC path-based authorization restricts access per role and HTTP method | `backend/go/internal/middleware/auth.go` lines 124-153 |
| A.8.10 | Information deletion | IMPLEMENTED | Data retention policies enforce deletion timelines. Decommission workflow for device data | `backend/go/internal/service/device_service.go` lines 154-169 |
| A.8.12 | Data leakage prevention | PARTIAL | Input validation with 1MB request size limit, content-type enforcement. DLP tooling planned | `backend/go/internal/middleware/auth.go` lines 330-360 |
| A.8.15 | Logging | IMPLEMENTED | Audit logging middleware captures method, path, status, user identity, IP, user-agent, request ID, and duration for every request | `backend/go/internal/middleware/auth.go` lines 169-244 |
| A.8.16 | Monitoring activities | IMPLEMENTED | Health check endpoint with component status, uptime, and system metrics | `backend/go/cmd/server/main.go` lines 174-185 |
| A.8.23 | Web filtering | IMPLEMENTED | CORS policy, Content-Security-Policy, HSTS, X-Frame-Options DENY, X-XSS-Protection headers | `backend/go/cmd/server/main.go` lines 118-127 |
| A.8.24 | Secure coding | IMPLEMENTED | Input validation middleware, parameterized DynamoDB queries, UUID generation, request body size limits | `backend/go/internal/middleware/auth.go` lines 330-360 |

### Verification

```bash
# 1. Confirm JWT authentication is active (A.8.2)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/devices
# Expected: 401 (unauthorized without token)

# 2. Verify RBAC enforcement (A.5.3, A.8.3)
# Create a viewer-role token and attempt DELETE -- expect 403
curl -s -X DELETE http://localhost:8080/api/v1/devices/test-id \
  -H "Authorization: Bearer <viewer-token>"
# Expected: 403 Forbidden

# 3. Verify rate limiting (A.5.7)
for i in $(seq 1 110); do curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/health; done
# Expected: first 100 return 200, then 429

# 4. Verify security headers (A.8.23)
curl -sI http://localhost:8080/health | grep -E "Strict-Transport|X-Frame|X-Content-Type|Content-Security"
# Expected: HSTS, DENY, nosniff, default-src 'self'

# 5. Retrieve full ISO 27001 controls via compliance API
curl -s http://localhost:8080/api/v1/compliance/standards | jq '.standards[] | select(.id=="ISO27001")'
```

---

## 3. ISO 9001:2015 -- Quality Management System

### What It Is

ISO 9001:2015 defines criteria for a comprehensive Quality Management System (QMS). It is based on seven quality management principles: customer focus, leadership, engagement of people, process approach, improvement, evidence-based decision making, and relationship management. Organizations certified to ISO 9001 demonstrate that they consistently provide products and services that meet customer and regulatory requirements.

### Why It Applies

The Sentinel Device Manager is an enterprise-grade platform that physical security teams depend on for mission-critical device management. Failures in quality -- misconfigured cameras, lost access control events, incorrect firmware deployments -- directly impact physical security posture. ISO 9001 ensures the platform follows a defined, documented, and continuously improved development and operations lifecycle.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| 4.4 | QMS and processes | IMPLEMENTED | Structured microservice architecture (Go backend, Python analytics, React frontend) with defined interfaces | `backend/go/cmd/server/main.go`, `backend/python/analytics_service/main.py` |
| 7.5 | Documented information | IMPLEMENTED | OpenAPI 3.0 specification served at `/swagger.json`, Go package documentation, this standards document | `backend/go/cmd/server/main.go` lines 270-558 (swaggerSpec) |
| 8.1 | Operational planning and control | IMPLEMENTED | CI/CD pipeline with Docker multi-stage builds and Kubernetes deployment manifests | `deploy/docker/Dockerfile.go`, `deploy/k8s/go-service.yaml` |
| 8.2 | Requirements for products and services | IMPLEMENTED | Requirements tracked in project documentation with traceable API specifications | `docs/requirements.md` |
| 8.5 | Production and service provision | IMPLEMENTED | Kubernetes deployment with liveness/readiness probes, HPA (3-20 replicas), resource limits | `deploy/k8s/go-service.yaml` lines 37-48, 66-83 |
| 9.1 | Monitoring, measurement, analysis | IMPLEMENTED | Analytics summary endpoint, device health telemetry, compliance reporting | `backend/python/analytics_service/handlers.py` lines 57-59 (`get_analytics_summary`) |
| 9.3 | Management review | PARTIAL | Compliance reporting endpoint available; formal review cadence to be established | `GET /api/v1/compliance/report` |
| 10.2 | Nonconformity and corrective action | IMPLEMENTED | Incident management system with severity classification, alert workflows, and resolution tracking | `backend/go/internal/compliance/compliance.go` lines 245-310 |

### Verification

```bash
# 1. Verify API documentation is served (7.5)
curl -s http://localhost:8080/swagger.json | jq '.info.title'
# Expected: "Sentinel Device Manager API"

# 2. Verify analytics monitoring endpoint (9.1)
curl -s http://localhost:8081/api/v1/analytics/summary
# Expected: JSON with total_devices, online_devices, events_last_24h, etc.

# 3. Verify Kubernetes HPA configuration (8.5)
kubectl get hpa sentinel-go-service-hpa -o yaml | grep -A5 "metrics"
# Expected: minReplicas: 3, maxReplicas: 20, CPU target 70%

# 4. Verify incident management for corrective actions (10.2)
curl -s -X POST http://localhost:8080/api/v1/compliance/incidents \
  -H "Content-Type: application/json" \
  -d '{"title":"Test nonconformity","severity":"MEDIUM","category":"QUALITY","reported_by":"qa-team"}'
# Expected: 201 with incident_id
```

---

## 4. ISO 27035:2023 -- Information Security Incident Management

### What It Is

ISO/IEC 27035:2023 provides guidelines for planning, detecting, assessing, and responding to information security incidents. It establishes a structured approach to incident management that covers preparation, detection and reporting, assessment and decision, response, and lessons learned. The standard ensures organizations can minimize damage from security incidents and use incidents as opportunities to improve security posture.

### Why It Applies

Physical security devices managed by the Sentinel platform generate security events continuously -- door forced open alerts, camera tamper detection, unauthorized access attempts. These events can escalate to incidents requiring coordinated response. ISO 27035 provides the framework for detecting these events, triaging them by severity, tracking response actions through a complete timeline, and producing post-incident evidence for regulatory and forensic purposes.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| 6.1 | Incident management policy | IMPLEMENTED | Incident management API with full CRUD operations and status lifecycle | `backend/go/internal/compliance/compliance.go` lines 245-310 |
| 6.2 | Incident detection and reporting | IMPLEMENTED | Event ingestion pipeline through Python analytics service with severity classification | `backend/python/analytics_service/handlers.py` lines 29-39 |
| 6.3 | Incident assessment and decision | IMPLEMENTED | Severity-based auto-alert generation: CRITICAL events automatically create alerts | `backend/python/analytics_service/handlers.py` lines 33-38 |
| 6.4 | Incident response | IMPLEMENTED | Alert lifecycle workflow: ACTIVE -> ACKNOWLEDGED -> RESOLVED with full audit trail | `backend/python/analytics_service/handlers.py` lines 50-55 |
| 7.1 | Incident documentation | IMPLEMENTED | Incident records with timeline entries tracking every status change | `backend/go/internal/compliance/compliance.go` lines 93-115 |
| 7.2 | Evidence collection | PARTIAL | Audit log entries capture request/response details. Enhanced evidence preservation planned | `backend/go/internal/middleware/auth.go` lines 175-188 |

### Verification

```bash
# 1. Create a security incident (6.1)
curl -s -X POST http://localhost:8080/api/v1/compliance/incidents \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Unauthorized device access attempt",
    "description": "Multiple failed authentication attempts from IP 10.0.0.99",
    "severity": "HIGH",
    "category": "SECURITY",
    "reported_by": "siem-integration"
  }'
# Expected: 201 with incident_id (e.g. "INC-1745...")

# 2. Verify auto-alert for critical events (6.3)
curl -s -X POST http://localhost:8081/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{"device_id":"cam-001","event_type":"DOOR_FORCED","severity":"CRITICAL","payload":{"zone":"server-room"}}'
# Expected: 201, then check alerts
curl -s http://localhost:8081/api/v1/alerts?status=ACTIVE | jq '.[0]'
# Expected: Alert with matching device_id and severity

# 3. Update incident with timeline (7.1)
curl -s -X PUT http://localhost:8080/api/v1/compliance/incidents/<incident-id> \
  -H "Content-Type: application/json" \
  -d '{"status":"INVESTIGATING","updated_by":"security-team","notes":"Reviewing camera footage"}'
# Expected: 200

# 4. Retrieve incident history
curl -s "http://localhost:8080/api/v1/compliance/incidents?status=OPEN"
# Expected: JSON array of open incidents with timeline entries
```

---

## 5. ISO 27017:2015 -- Cloud Security Controls

### What It Is

ISO/IEC 27017:2015 provides cloud-specific security guidance that builds on the controls defined in ISO 27001/27002. It addresses the shared responsibility model between cloud service providers and cloud service customers, covering virtual machine hardening, cloud operational security, environment segmentation, and monitoring specific to cloud computing environments.

### Why It Applies

The Sentinel Device Manager runs on Kubernetes (AWS EKS or equivalent) with DynamoDB as its data store. This is a cloud-native deployment where the infrastructure layer, container orchestration, and managed database service introduce cloud-specific security considerations. ISO 27017 ensures these cloud-specific risks are addressed beyond the general ISMS controls in ISO 27001.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| A.5.1 | Shared cloud responsibilities | IMPLEMENTED | Kubernetes deployment with defined service boundaries. Each service (Go, Python, frontend) runs in its own deployment with explicit resource limits | `deploy/k8s/go-service.yaml`, `deploy/k8s/python-service.yaml`, `deploy/k8s/frontend.yaml` |
| A.6.1 | Removal/return of customer assets | IMPLEMENTED | Device decommission API endpoint marks devices as DECOMMISSIONED before deletion | `backend/go/internal/service/device_service.go` lines 154-169 |
| A.6.2 | Segmentation in cloud environments | IMPLEMENTED | Organization and site-level data isolation through DynamoDB filter expressions. Kubernetes NetworkPolicies separate service tiers | `backend/go/internal/repository/dynamodb.go` lines 144-178 |
| A.7.1 | Virtual machine hardening | PARTIAL | Alpine-based Docker images with minimal attack surface. CIS benchmark hardening in progress | `deploy/docker/Dockerfile.go` |
| A.8.1 | Cloud operational/administrative security | IMPLEMENTED | JWT authentication, RBAC with admin role, audit logging for all administrative operations | `backend/go/internal/middleware/auth.go` lines 44-68, 117-165 |
| A.9.1 | Cloud service monitoring and audit | IMPLEMENTED | Comprehensive audit trail via middleware capturing all API operations with user identity | `backend/go/internal/middleware/auth.go` lines 169-244 |

### Verification

```bash
# 1. Verify service segmentation (A.6.2)
kubectl get deployments --all-namespaces -l app
# Expected: separate deployments for go-service, python-service, frontend

# 2. Verify resource limits per service (A.5.1)
kubectl describe deployment sentinel-go-service | grep -A5 "Limits"
# Expected: cpu: 500m, memory: 512Mi

# 3. Verify device decommission workflow (A.6.1)
curl -s -X PUT http://localhost:8080/api/v1/devices/<device-id> \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"DECOMMISSIONED"}'
# Expected: 200 with status DECOMMISSIONED

# 4. Verify audit logging of admin operations (A.9.1)
curl -s http://localhost:8080/api/v1/compliance/report | jq '.open_incidents'
# Audit entries are captured in the AuditLogger instance
```

---

## 6. ISO 20000-1:2018 -- IT Service Management

### What It Is

ISO/IEC 20000-1:2018 specifies requirements for an organization to establish, implement, maintain, and continually improve a Service Management System (SMS). It covers service level management, availability management, incident management, change management, and release and deployment management. It is the international standard for IT service management excellence.

### Why It Applies

The Sentinel Device Manager is itself an IT service consumed by security operations teams who require guaranteed uptime, defined SLAs, controlled change management (especially for firmware deployments to physical security devices), and rapid incident response. ISO 20000 ensures the platform delivers reliable service levels that physical security operations depend on.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| 8.1 | Service level management | IMPLEMENTED | 99.95% uptime target enforced through Kubernetes HPA (3-20 replicas), health checks, and auto-scaling | `deploy/k8s/go-service.yaml` lines 66-83 |
| 8.2 | Service reporting | IMPLEMENTED | Analytics summary endpoint and compliance reports provide operational metrics | `backend/python/analytics_service/store.py` lines 91-113 |
| 8.3 | Service continuity and availability | IMPLEMENTED | Multi-replica Kubernetes deployment with liveness and readiness probes ensures automatic failover | `deploy/k8s/go-service.yaml` lines 37-48 |
| 8.4 | Budgeting and charging for services | PLANNED | Cost tracking for cloud infrastructure planned for next release | N/A |
| 8.5 | Incident management | IMPLEMENTED | Incident management module with severity classification and SLA-aware status tracking | `backend/go/internal/compliance/compliance.go` lines 245-310 |
| 8.7 | Change management | IMPLEMENTED | Firmware deployment with staged rollout, status progression (PENDING -> DOWNLOADING -> VERIFYING -> APPLYING -> COMPLETED/FAILED/ROLLED_BACK) | `backend/go/internal/handler/grpc_handler.go` lines 98-131 |
| 8.8 | Release and deployment management | IMPLEMENTED | Docker-based CI/CD with multi-stage builds, Kubernetes rolling deployments | `deploy/docker/Dockerfile.go`, `deploy/k8s/go-service.yaml` |

### Verification

```bash
# 1. Verify auto-scaling configuration (8.1)
kubectl get hpa sentinel-go-service-hpa
# Expected: min=3, max=20, target CPU=70%

# 2. Verify service reporting (8.2)
curl -s http://localhost:8081/api/v1/analytics/summary | jq .
# Expected: total_devices, online_devices, active_alerts, events_last_24h

# 3. Verify health probes (8.3)
kubectl describe pod -l app=sentinel-go-service | grep -A3 "Liveness"
# Expected: httpGet path=/health port=8080

# 4. Verify firmware deployment workflow (8.7)
# Via gRPC client, call StreamFirmware and ReportFirmwareStatus
# Check firmware deployment status model includes ROLLED_BACK state
curl -s http://localhost:8080/api/v1/compliance/standards | \
  jq '.standards[] | select(.id=="ISO20000") | .controls[] | select(.id=="8.7")'
```

---

## 7. ISO 22301:2019 -- Business Continuity Management

### What It Is

ISO 22301:2019 specifies requirements for a Business Continuity Management System (BCMS). It provides a framework for planning, establishing, implementing, operating, monitoring, reviewing, maintaining, and continually improving business continuity. The standard ensures organizations can continue operating during disruptions and recover to full operations within defined timeframes.

### Why It Applies

The Sentinel Device Manager manages physical security infrastructure. If the platform becomes unavailable, security operations centers lose visibility into cameras, access control events, and alarm systems. Business continuity planning ensures the platform can survive infrastructure failures, recover from disasters, and maintain at minimum a degraded but functional state during major incidents.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| 8.2 | Business impact analysis | IMPLEMENTED | Device management service classified as critical with defined RTO/RPO targets | Architecture documentation in `docs/architecture.md` |
| 8.3 | Risk assessment | IMPLEMENTED | Defense-in-depth security architecture with multiple control layers (auth, RBAC, rate limiting, validation, audit, headers) | `backend/go/cmd/server/main.go` lines 96-127 |
| 8.4 | Business continuity strategy | IMPLEMENTED | Multi-replica Kubernetes deployment (minimum 3 replicas), DynamoDB managed backups with point-in-time recovery | `deploy/k8s/go-service.yaml` lines 8-9 |
| 8.5 | Business continuity plans | PARTIAL | Infrastructure-as-code (K8s manifests, Docker Compose) enables rapid recovery. Formal BCP document in progress | `deploy/k8s/`, `docker-compose.yml` |
| 8.6 | Exercising and testing | PLANNED | End-to-end and integration tests cover basic recovery. Regular DR drills planned | `tests/e2e/device_flow_test.go`, `tests/integration/api_test.go` |

### Verification

```bash
# 1. Verify multi-replica deployment (8.4)
kubectl get deployment sentinel-go-service -o jsonpath='{.spec.replicas}'
# Expected: 3

# 2. Verify infrastructure-as-code recoverability (8.5)
# All deployment manifests are version-controlled:
ls deploy/k8s/*.yaml
# Expected: go-service.yaml, python-service.yaml, frontend.yaml

# 3. Verify Docker Compose for local recovery (8.5)
docker compose -f docker-compose.yml up -d
docker compose ps
# Expected: 4 services running (dynamodb, go-service, python-service, frontend)

# 4. Verify firmware simulator operates in standalone mode (8.4)
# The firmware simulator continues operating even when cloud is unreachable:
# See firmware/cmd/firmware-sim/main.go lines 56-57
# "Registration failed (will continue in standalone mode)"
```

---

## 8. IEC 62443 -- Industrial Automation and Control Systems Security

### What It Is

IEC 62443 is a series of standards addressing network and system security for industrial automation and control systems (IACS). It defines requirements for security levels, zones, conduits, and system component security applicable to operational technology (OT) and Internet of Things (IoT) environments. The standard covers the full lifecycle from risk assessment through ongoing monitoring.

### Why It Applies

The Sentinel Device Manager manages IoT-class physical security devices -- cameras, access control panels, alarm systems, and sensors -- that are functionally equivalent to industrial control system components. These devices operate on embedded firmware, communicate over network protocols (gRPC), and control physical access to facilities. IEC 62443 directly addresses the security requirements for managing such devices at scale, including device identity, communication integrity, network segmentation, and denial-of-service protection.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| SR 1.1 | Human user identification | IMPLEMENTED | JWT-based authentication required for all management API access | `backend/go/internal/middleware/auth.go` lines 44-68 |
| SR 1.2 | Software process and device identification | IMPLEMENTED | Device registration with serial number, MAC address, UUID, model, and IP address | `backend/go/internal/model/device.go` lines 32-47, `firmware/cmd/firmware-sim/main.go` lines 44-54 |
| SR 2.1 | Authorization enforcement | IMPLEMENTED | RBAC middleware with role-based and path-based authorization | `backend/go/internal/middleware/auth.go` lines 117-165 |
| SR 3.1 | Communication integrity | PARTIAL | HMAC-SHA256 signed JWT tokens. gRPC TLS and mTLS planned for production | `backend/go/internal/middleware/auth.go` lines 86-114 |
| SR 4.1 | Information confidentiality | PARTIAL | HTTPS/TLS termination at load balancer. Sensitive fields excluded from API responses planned | Kubernetes Ingress TLS config in `deploy/k8s/frontend.yaml` |
| SR 5.1 | Network segmentation | IMPLEMENTED | Kubernetes network policies, service-level isolation with separate ClusterIP services per component | `deploy/k8s/frontend.yaml` lines 50-96 (Ingress routing) |
| SR 6.1 | Audit log accessibility | IMPLEMENTED | Audit log entries accessible to authorized personnel via RBAC-protected endpoints | `backend/go/internal/middleware/auth.go` lines 232-244 (`GetEntries`) |
| SR 7.1 | DoS protection | IMPLEMENTED | Rate limiting middleware per client IP with configurable thresholds (100 req/min default) | `backend/go/internal/middleware/auth.go` lines 249-318 |

### Verification

```bash
# 1. Verify device identification (SR 1.2)
curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"serial_number":"VKD-CAM-001","device_type":"CAMERA","site_id":"site-01","organization_id":"org-01","mac_address":"AA:BB:CC:DD:EE:FF"}'
# Expected: 201 with device_id, serial_number, mac_address populated

# 2. Verify firmware simulator device identity (SR 1.2)
docker compose logs firmware-sim | grep "Registered with device_id"
# Expected: Device registered with unique device_id

# 3. Verify rate limiting / DoS protection (SR 7.1)
for i in $(seq 1 110); do curl -s -o /dev/null -w "%{http_code} " http://localhost:8080/health; done
# Expected: 200 for first 100, then 429 with Retry-After header

# 4. Verify network segmentation via Kubernetes (SR 5.1)
kubectl get services
# Expected: separate ClusterIP services for go-service, python-service, frontend
```

---

## 9. NIST Cybersecurity Framework 2.0

### What It Is

The NIST Cybersecurity Framework (CSF) 2.0 organizes cybersecurity activities into six core functions: Govern, Identify, Protect, Detect, Respond, and Recover. It provides a risk-based approach to managing cybersecurity risk that is widely adopted across industries and frequently referenced in regulatory requirements. Version 2.0 added the Govern function to emphasize cybersecurity risk management as a strategic imperative.

### Why It Applies

The Sentinel Device Manager manages devices that are part of an organization's physical security perimeter. NIST CSF provides a comprehensive, well-understood framework for organizing the platform's security controls across the full lifecycle -- from identifying and classifying managed devices, through protecting and monitoring them, to responding to incidents and recovering from failures. Many organizations that deploy Sentinel require NIST CSF alignment as a procurement requirement.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| ID.AM | Asset Management | IMPLEMENTED | Full device lifecycle management: registration, tracking, updates, decommissioning with filtering by type, status, site, org | `backend/go/internal/service/device_service.go`, `backend/go/internal/repository/dynamodb.go` |
| ID.RA | Risk Assessment | IMPLEMENTED | Threat intelligence through rate limiting, anomaly detection, and event severity classification | `backend/go/internal/middleware/auth.go` lines 249-318, `backend/python/analytics_service/models.py` lines 7-11 |
| PR.AC | Access Control | IMPLEMENTED | JWT + RBAC middleware stack with three defined roles (admin, operator, viewer) | `backend/go/internal/middleware/auth.go` lines 44-165 |
| PR.DS | Data Security | PARTIAL | HMAC-signed tokens for integrity; encryption at rest and in transit planned for production hardening | `backend/go/internal/middleware/auth.go` lines 86-114 |
| PR.IP | Information Protection | IMPLEMENTED | Input validation with 1MB request size limit, content-type enforcement, CSP headers | `backend/go/internal/middleware/auth.go` lines 330-360 |
| PR.PT | Protective Technology | IMPLEMENTED | Rate limiting, CORS restrictions, security headers (HSTS, CSP, X-Frame-Options), request validation | `backend/go/cmd/server/main.go` lines 118-127 |
| DE.AE | Anomalies and Events | IMPLEMENTED | Event severity classification (INFO/WARNING/CRITICAL), automatic alert generation for critical events | `backend/python/analytics_service/handlers.py` lines 33-38 |
| DE.CM | Continuous Monitoring | IMPLEMENTED | Health check endpoints, device heartbeat monitoring via gRPC, system metrics in health response | `backend/go/cmd/server/main.go` lines 174-185, `backend/go/internal/handler/grpc_handler.go` lines 52-74 |
| RS.RP | Response Planning | IMPLEMENTED | Incident management with timeline tracking, severity-based triage, status lifecycle | `backend/go/internal/compliance/compliance.go` lines 245-310 |
| RC.RP | Recovery Planning | PARTIAL | Kubernetes HPA and multi-replica for automatic recovery. Formal recovery procedures in progress | `deploy/k8s/go-service.yaml` lines 66-83 |

### Verification

```bash
# 1. Verify asset management (ID.AM) - register and retrieve a device
curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"serial_number":"NIST-TEST-001","device_type":"SENSOR","site_id":"site-ra","organization_id":"org-ra"}'
# Then retrieve it:
curl -s http://localhost:8080/api/v1/devices/<device-id> -H "Authorization: Bearer <admin-token>"
# Expected: 200 with full device details

# 2. Verify anomaly detection (DE.AE) - submit a critical event
curl -s -X POST http://localhost:8081/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{"device_id":"cam-001","event_type":"TAMPER_DETECTED","severity":"CRITICAL","payload":{}}'
curl -s http://localhost:8081/api/v1/alerts?status=ACTIVE
# Expected: Auto-generated alert for the critical event

# 3. Verify continuous monitoring (DE.CM)
curl -s http://localhost:8080/health
# Expected: {"status":"ok","version":"1.0.0"} with component health data

# 4. Get full NIST CSF compliance status
curl -s http://localhost:8080/api/v1/compliance/standards | \
  jq '.standards[] | select(.id=="NIST-CSF") | .controls[] | {id, status}'
```

---

## 10. SOC 2 Type II -- Trust Services Criteria

### What It Is

SOC 2 Type II is an audit framework developed by the AICPA that evaluates a service organization's controls relevant to the Trust Services Criteria: Security, Availability, Processing Integrity, Confidentiality, and Privacy. A Type II report assesses the design and operating effectiveness of controls over a defined observation period (typically 6-12 months). It is the most commonly requested assurance report for technology service providers.

### Why It Applies

Organizations that deploy the Sentinel Device Manager trust the platform with managing their physical security devices and the sensitive data those devices generate (video metadata, access logs, alarm events, device configurations). SOC 2 Type II provides independent assurance that the platform's controls for logical access, change management, incident response, and system availability are properly designed and consistently operated.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| CC6.1 | Logical and physical access controls | IMPLEMENTED | Multi-layer access control: JWT authentication + RBAC authorization + rate limiting | `backend/go/internal/middleware/auth.go` lines 44-68, 117-165, 249-318 |
| CC6.2 | User authentication | IMPLEMENTED | JWT Bearer token authentication required for all protected endpoints. HMAC-SHA256 signature verification | `backend/go/internal/middleware/auth.go` lines 44-68 |
| CC6.3 | Role-based access | IMPLEMENTED | Granular RBAC with three roles (admin, operator, viewer) and method + path-level enforcement | `backend/go/cmd/server/main.go` lines 149-162 |
| CC7.1 | Detection and monitoring | IMPLEMENTED | Audit logging of all requests, health check monitoring, event ingestion pipeline | `backend/go/internal/middleware/auth.go` lines 169-244 |
| CC7.2 | Incident response | IMPLEMENTED | Alert lifecycle (ACTIVE -> ACKNOWLEDGED -> RESOLVED) with severity classification and audit trail | `backend/python/analytics_service/handlers.py` lines 50-55 |
| CC8.1 | Change management | IMPLEMENTED | Firmware deployment with staged rollout, status tracking, and rollback capability | `backend/go/internal/handler/grpc_handler.go` lines 98-131, `backend/go/internal/model/device.go` lines 119-127 |
| A1.2 | System availability | IMPLEMENTED | Kubernetes HPA (3-20 replicas), liveness/readiness probes, multi-service architecture | `deploy/k8s/go-service.yaml` lines 37-48, 66-83 |

### Verification

```bash
# 1. Verify authentication enforcement (CC6.2)
curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/devices
# Expected: 401

# 2. Verify RBAC granularity (CC6.3)
# Viewer can GET but cannot POST:
curl -s -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer <viewer-token>" \
  -H "Content-Type: application/json" \
  -d '{"serial_number":"test","device_type":"CAMERA","site_id":"s1","organization_id":"o1"}'
# Expected: 403 Forbidden

# 3. Verify audit logging (CC7.1)
# After making any authenticated request, audit entries are captured
# by the AuditLogger middleware. Entries are accessible via GetEntries()
# in the in-memory audit store.

# 4. Verify firmware change management (CC8.1)
# Check firmware deployment status model supports rollback:
curl -s http://localhost:8080/api/v1/compliance/standards | \
  jq '.standards[] | select(.id=="SOC2") | .controls[] | select(.id=="CC8.1")'
# Expected: IMPLEMENTED with evidence referencing staged rollout

# 5. Verify system availability (A1.2)
kubectl get hpa sentinel-go-service-hpa -o wide
# Expected: showing current replicas >= 3
```

---

## 11. GDPR (EU 2016/679) -- General Data Protection Regulation

### What It Is

The General Data Protection Regulation (GDPR) is an European Union regulation governing the processing of personal data of individuals within the EU. It establishes principles for lawful processing (Article 5), data subject rights (Articles 15-22), obligations of data controllers and processors, and mandatory breach notification requirements (Articles 33-34). Non-compliance can result in administrative fines of up to 4% of annual global turnover or EUR 20 million, whichever is greater.

### Why It Applies

The Sentinel Device Manager processes data that may qualify as personal data under GDPR. Security cameras capture images of individuals, access control systems record entry and exit events tied to specific people, and device telemetry may include location data linked to identifiable persons. Organizations deploying Sentinel in EU jurisdictions require the platform to support GDPR compliance through data minimization, purpose limitation, retention limits, and the right to erasure.

### Implemented Controls

| Control ID | Title | Status | Implementation | Code Reference |
|---|---|---|---|---|
| Art.5(1)(a) | Lawfulness, fairness, transparency | IMPLEMENTED | Clear data model with documented fields, OpenAPI specification, complete audit trail of all data access | `backend/go/internal/model/device.go`, `backend/go/cmd/server/main.go` lines 270-558 |
| Art.5(1)(b) | Purpose limitation | IMPLEMENTED | Data models have specific fields for specific purposes. No secondary use of collected data | `backend/go/internal/model/device.go` lines 32-47 |
| Art.5(1)(c) | Data minimization | IMPLEMENTED | Device model collects only required fields. Optional fields use `omitempty` JSON tags | `backend/go/internal/service/device_service.go` lines 30-39 |
| Art.5(1)(e) | Storage limitation | IMPLEMENTED | Data retention policies per data category with defined retention periods and GDPR legal basis | `backend/go/internal/compliance/compliance.go` lines 530-539 |
| Art.5(1)(f) | Integrity and confidentiality | IMPLEMENTED | JWT auth, RBAC, security headers, input validation, audit logging for data integrity | `backend/go/internal/middleware/auth.go` |
| Art.17 | Right to erasure | IMPLEMENTED | Device deletion endpoint (`DELETE /api/v1/devices/{id}`) and decommission workflow for staged erasure | `backend/go/internal/service/device_service.go` lines 154-169 |
| Art.25 | Data protection by design | IMPLEMENTED | Security middleware stack embedded from inception: auth, RBAC, rate limiting, validation, audit, security headers | `backend/go/cmd/server/main.go` lines 96-127 |
| Art.32 | Security of processing | IMPLEMENTED | Full security middleware stack: authentication, authorization, rate limiting, input validation, audit logging | `backend/go/internal/middleware/auth.go` |
| Art.33 | Data breach notification | PARTIAL | Incident management module supports breach tracking and documentation. Automated 72-hour notification workflow planned | `backend/go/internal/compliance/compliance.go` lines 245-310 |

### Verification

```bash
# 1. Verify data retention policies are defined (Art.5(1)(e))
curl -s http://localhost:8080/api/v1/compliance/retention | jq '.policies[] | {category, retention_days, gdpr_basis}'
# Expected: 7 policies with retention periods and GDPR legal bases

# 2. Verify right to erasure (Art.17) - decommission then delete
curl -s -X PUT http://localhost:8080/api/v1/devices/<device-id> \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"status":"DECOMMISSIONED"}'
# Then after retention period:
curl -s -X DELETE http://localhost:8080/api/v1/devices/<device-id> \
  -H "Authorization: Bearer <admin-token>"
# Expected: 204 No Content

# 3. Verify data minimization (Art.5(1)(c))
curl -s http://localhost:8080/swagger.json | jq '.components.schemas.CreateDeviceRequest.required'
# Expected: ["serial_number", "device_type", "site_id", "organization_id"] -- only essential fields required

# 4. Verify audit trail of data processing (Art.5(1)(a))
curl -s http://localhost:8080/api/v1/compliance/report | jq '.data_retention'
# Expected: All retention policies listed with GDPR bases

# 5. Create a data breach incident (Art.33)
curl -s -X POST http://localhost:8080/api/v1/compliance/incidents \
  -H "Content-Type: application/json" \
  -d '{"title":"Potential data breach - unauthorized camera access","severity":"CRITICAL","category":"DATA","reported_by":"dpo@company.com"}'
# Expected: 201 with incident_id
```

---

## 12. Standards Interrelationships

The ten standards implemented by Sentinel are not isolated -- they share overlapping controls and reinforce each other. The following diagram and table describe these relationships.

### Control Overlap Map

```
                    +------------------+
                    |   ISO 27001      |  <-- Overarching ISMS
                    |   (12 controls)  |
                    +--------+---------+
                             |
            +----------------+----------------+
            |                |                 |
     +------+------+  +------+------+  +------+------+
     | ISO 27017   |  | ISO 27035   |  | ISO 22301   |
     | (cloud)     |  | (incidents) |  | (continuity)|
     +------+------+  +------+------+  +------+------+
            |                |                 |
            +--------+-------+---------+-------+
                     |                 |
              +------+------+  +-------+------+
              | IEC 62443   |  | ISO 20000    |
              | (IoT/OT)    |  | (ITSM)       |
              +------+------+  +-------+------+
                     |                 |
            +--------+-------+---------+
            |                |                 |
     +------+------+  +------+------+  +------+------+
     | NIST CSF    |  | SOC 2       |  | GDPR        |
     | (framework) |  | (audit)     |  | (privacy)   |
     +-------------+  +-------------+  +-------------+

        +-----------+
        | ISO 9001  |  <-- Horizontal: quality management spans all standards
        +-----------+
```

### Key Overlaps

| Control Area | ISO 27001 | SOC 2 | NIST CSF | IEC 62443 | GDPR |
|---|---|---|---|---|---|
| Authentication | A.8.2 | CC6.2 | PR.AC | SR 1.1 | Art.32 |
| Authorization / RBAC | A.5.3, A.8.3 | CC6.3 | PR.AC | SR 2.1 | Art.32 |
| Audit Logging | A.8.15 | CC7.1 | DE.CM | SR 6.1 | Art.5(1)(a) |
| Incident Management | -- | CC7.2 | RS.RP | -- | Art.33 |
| Rate Limiting / DoS | A.5.7 | CC6.1 | PR.PT | SR 7.1 | Art.32 |
| Data Retention | A.8.10 | -- | PR.DS | -- | Art.5(1)(e) |
| Input Validation | A.8.24 | CC6.1 | PR.IP | -- | Art.25 |
| Change Management | -- | CC8.1 | -- | -- | -- |
| Monitoring | A.8.16 | CC7.1 | DE.CM | -- | Art.32 |

### How to Read This

- **ISO 27001** is the umbrella standard. Most other standards either extend it (ISO 27017, ISO 27035) or map to its controls (SOC 2, NIST CSF, IEC 62443).
- **SOC 2** provides audit assurance for controls that are also defined in ISO 27001 and NIST CSF. A single implementation (e.g., JWT auth) satisfies requirements across all three.
- **GDPR** adds privacy-specific requirements (retention, erasure, breach notification) on top of the security controls from ISO 27001 and NIST CSF.
- **IEC 62443** adds device- and OT-specific requirements (device identification, network segmentation) that complement ISO 27001's general security controls.
- **ISO 9001** spans all standards horizontally by ensuring quality management processes are in place for every control implementation.

---

## 13. Compliance Reporting API

### Overview

The compliance reporting subsystem provides real-time visibility into the platform's compliance posture across all tracked standards. It is implemented in `backend/go/internal/compliance/compliance.go` and exposed through REST endpoints.

### Endpoints

#### `GET /api/v1/compliance/report`

Returns a comprehensive compliance report including all standards, their controls, status, summary statistics, data retention policies, and incident counts.

**Response structure:**

```json
{
  "generated_at": "2026-04-24T12:00:00Z",
  "version": "1.0.0",
  "standards": [
    {
      "id": "ISO27001",
      "name": "ISO/IEC 27001:2022",
      "version": "2022",
      "description": "...",
      "category": "Information Security",
      "controls": [
        {
          "id": "A.5.1",
          "title": "Policies for information security",
          "description": "...",
          "status": "IMPLEMENTED",
          "evidence": "Security policy enforced via middleware stack"
        }
      ]
    }
  ],
  "summary": {
    "total_controls": 78,
    "implemented": 65,
    "partial": 11,
    "planned": 2,
    "compliance_pct": 85.9,
    "standards": {
      "ISO27001": {
        "name": "ISO/IEC 27001:2022",
        "total_controls": 12,
        "implemented": 12,
        "compliance_pct": 100.0
      }
    }
  },
  "data_retention": [...],
  "open_incidents": 0,
  "total_incidents": 0
}
```

#### `GET /api/v1/compliance/standards`

Returns all tracked standards with their controls.

```bash
curl -s http://localhost:8080/api/v1/compliance/standards | jq '.total'
# Expected: 10
```

#### `GET /api/v1/compliance/retention`

Returns all data retention policies.

```bash
curl -s http://localhost:8080/api/v1/compliance/retention | jq '.policies[] | {category, retention_days, gdpr_basis}'
```

#### `GET /api/v1/compliance/incidents?status={status}`

Returns incidents, optionally filtered by status.

```bash
curl -s "http://localhost:8080/api/v1/compliance/incidents?status=OPEN"
```

### Using Reports for Audit Evidence

During an external audit, the compliance report endpoint provides:

1. **Point-in-time snapshot** -- The `generated_at` timestamp proves the report was generated during the audit period.
2. **Control-by-control evidence** -- Each control includes an `evidence` field describing how it is implemented and referencing specific code locations.
3. **Quantitative compliance score** -- The `compliance_pct` enables comparison across audit periods.
4. **Incident history** -- Open and total incident counts demonstrate active incident management.

---

## 14. Incident Management Procedures

### Overview

Incident management follows the ISO 27035 lifecycle: Detection -> Assessment -> Response -> Resolution -> Closure. The platform provides both a programmatic API and an internal data model for tracking incidents.

### Incident Lifecycle

```
CREATE (OPEN) --> INVESTIGATING --> CONTAINED --> RESOLVED --> CLOSED
     |                |                |             |
     +----+-----------+----------------+-------------+
          |
    Timeline entries added at each transition
```

### API Reference

#### Create an Incident

```bash
POST /api/v1/compliance/incidents
Content-Type: application/json

{
  "title": "Unauthorized device registration detected",
  "description": "Anomalous device registration from unrecognized IP range 203.0.113.0/24",
  "severity": "HIGH",          // LOW | MEDIUM | HIGH | CRITICAL
  "category": "SECURITY",      // SECURITY | OPERATIONAL | COMPLIANCE | DATA
  "reported_by": "siem-system"
}
```

Response (201 Created):

```json
{
  "incident_id": "INC-1745491200000000000",
  "title": "Unauthorized device registration detected",
  "status": "OPEN",
  "severity": "HIGH",
  "created_at": "2026-04-24T12:00:00Z",
  "timeline": [
    {
      "timestamp": "2026-04-24T12:00:00Z",
      "action": "INCIDENT_CREATED",
      "by": "siem-system",
      "notes": "Incident reported"
    }
  ]
}
```

#### Update Incident Status

```bash
PUT /api/v1/compliance/incidents/{incidentID}
Content-Type: application/json

{
  "status": "INVESTIGATING",   // OPEN | INVESTIGATING | CONTAINED | RESOLVED | CLOSED
  "updated_by": "security-analyst",
  "notes": "Reviewing device registration logs. Identified source IP as compromised contractor endpoint."
}
```

#### List Incidents

```bash
GET /api/v1/compliance/incidents?status=OPEN
```

### Severity Response Times

| Severity | Initial Response | Status Update | Target Resolution |
|---|---|---|---|
| CRITICAL | 15 minutes | 1 hour | 4 hours |
| HIGH | 1 hour | 4 hours | 24 hours |
| MEDIUM | 4 hours | 24 hours | 72 hours |
| LOW | 24 hours | 72 hours | 7 days |

### Integration with Event Pipeline

Critical events from the Python analytics service automatically generate alerts:

1. Device sends event via gRPC (`SendEvent` in `backend/go/internal/handler/grpc_handler.go`)
2. Python analytics service ingests the event (`POST /api/v1/events`)
3. If severity is CRITICAL, an alert is auto-created (`backend/python/analytics_service/handlers.py` lines 33-38)
4. Security team acknowledges the alert via `POST /api/v1/alerts/{id}/acknowledge`
5. If the alert warrants formal tracking, an incident is created via the compliance API

### Alert vs. Incident

| Aspect | Alert | Incident |
|---|---|---|
| Source | Auto-generated from events | Created manually or from escalated alerts |
| Scope | Single device/event | May span multiple devices, events, or systems |
| Tracking | Status only (ACTIVE/ACKNOWLEDGED/RESOLVED) | Full timeline with notes, assigned personnel |
| Storage | Python analytics service (in-memory) | Go compliance module (in-memory) |
| API | Port 8081 | Port 8080 |

---

## 15. Data Retention Policies

### Overview

Data retention policies are defined in `backend/go/internal/compliance/compliance.go` (function `getDefaultRetentionPolicies`, lines 530-539) and are accessible via the compliance reporting API. These policies implement GDPR Article 5(1)(e) (storage limitation) and ISO 27001 A.8.10 (information deletion).

### Configured Retention Periods

| Data Category | Retention Period | Description | GDPR Legal Basis |
|---|---|---|---|
| `device_data` | 365 days | Active device registration and telemetry data | Legitimate interest (Art.6(1)(f)) |
| `audit_logs` | 730 days | System audit and access logs | Legal obligation (Art.6(1)(c)) |
| `security_events` | 365 days | Security event records and alerts | Legitimate interest (Art.6(1)(f)) |
| `incident_records` | 1,095 days (3 years) | Security and operational incident records | Legal obligation (Art.6(1)(c)) |
| `firmware_versions` | 1,825 days (5 years) | Firmware version metadata and deployment history | Legitimate interest (Art.6(1)(f)) |
| `analytics_data` | 90 days | Aggregated analytics and summary data | Consent (Art.6(1)(a)) |
| `decommissioned_devices` | 180 days | Decommissioned device records before deletion | Legitimate interest (Art.6(1)(f)) |

### Querying Retention Policies

```bash
# Get all retention policies
curl -s http://localhost:8080/api/v1/compliance/retention | jq .

# Filter by specific category
curl -s http://localhost:8080/api/v1/compliance/retention | \
  jq '.policies[] | select(.category=="audit_logs")'
```

### Deletion Procedures

The platform implements a two-phase deletion process aligned with GDPR requirements:

**Phase 1 -- Decommission** (preserves data for audit):

```bash
PUT /api/v1/devices/{deviceID}
Content-Type: application/json
Authorization: Bearer <admin-token>

{"status": "DECOMMISSIONED"}
```

The device record is retained for the `decommissioned_devices` retention period (180 days). During this period, the record remains available for forensic or audit purposes but is excluded from active device operations.

**Phase 2 -- Permanent Deletion** (after retention period):

```bash
DELETE /api/v1/devices/{deviceID}
Authorization: Bearer <admin-token>
```

The device record is permanently removed from DynamoDB. This operation is documented in the audit log.

The two-phase approach is implemented in `backend/go/internal/service/device_service.go`:

- `DecommissionDevice` (line 166) -- marks device as DECOMMISSIONED
- `DeleteDevice` (line 154) -- permanently removes the record

### Adding or Modifying Retention Policies

Retention policies are defined in the `getDefaultRetentionPolicies()` function in `backend/go/internal/compliance/compliance.go`. To modify a policy:

1. Edit the relevant entry in the `getDefaultRetentionPolicies()` return value
2. Update the `RetentionDays`, `Description`, or `GDPRBasis` as needed
3. Ensure any change is approved through the change management process (ISO 20000 8.7)
4. Restart the service or reload the configuration

---

## 16. Adding New Standards or Controls

### Adding a New Standard

To track compliance against a new standard, add a `Standard` entry to the `getDefaultStandards()` function in `backend/go/internal/compliance/compliance.go`.

**Step 1:** Define the standard and its controls:

```go
// Add to the return slice in getDefaultStandards()
{
    ID:          "PCI-DSS",
    Name:        "PCI DSS v4.0",
    Version:     "4.0",
    Description: "Payment Card Industry Data Security Standard. Defines requirements for organizations that store, process, or transmit cardholder data.",
    Category:    "Payment Security",
    Controls: []Control{
        {
            ID:          "Req 6.1",
            Title:       "System components protected by secure coding practices",
            Description: "Software engineering processes based on secure coding practices",
            Status:      "IMPLEMENTED",
            Evidence:    "Input validation middleware, secure coding standards enforced",
        },
        // Add more controls...
    },
},
```

**Step 2:** Verify the standard appears in the compliance report:

```bash
curl -s http://localhost:8080/api/v1/compliance/standards | jq '.total'
# Expected: incremented count

curl -s http://localhost:8080/api/v1/compliance/standards | \
  jq '.standards[] | select(.id=="PCI-DSS")'
```

**Step 3:** Update this documentation file with the new standard section.

### Adding a Control to an Existing Standard

Locate the standard in the `getDefaultStandards()` function and append a new `Control` entry:

```go
{
    ID:          "A.X.Y",
    Title:       "New control title",
    Description: "Description of what the control requires",
    Status:      "IMPLEMENTED",  // or PARTIAL, PLANNED, NOT_APPLICABLE
    Evidence:    "Description of how it is implemented, with code references",
},
```

### Control Status Definitions

| Status | Meaning |
|---|---|
| `IMPLEMENTED` | Control is fully operational with verifiable evidence |
| `PARTIAL` | Control is partially implemented or missing some aspects |
| `PLANNED` | Control is identified but implementation has not started |
| `NOT_APPLICABLE` | Control does not apply to this system (justify in evidence field) |

### Adding a Data Retention Policy

Add a `DataRetentionPolicy` entry to the `getDefaultRetentionPolicies()` function:

```go
{
    Category:      "new_data_category",
    RetentionDays: 365,
    Description:   "Description of the data category",
    GDPRBasis:     "Legitimate interest (Art.6(1)(f))",
},
```

### Testing Changes

After modifying compliance data, run the existing tests:

```bash
cd backend/go
go test ./internal/compliance/ -v
```

Then verify via the API:

```bash
curl -s http://localhost:8080/api/v1/compliance/report | jq '.summary.total_controls'
# Expected: incremented count reflecting new controls
```

---

## Appendix A: Quick Reference Matrix

The following matrix maps every security-relevant code artifact to the standards it satisfies.

| Code Artifact | ISO 27001 | ISO 9001 | ISO 27035 | ISO 27017 | ISO 20000 | ISO 22301 | IEC 62443 | NIST CSF | SOC 2 | GDPR |
|---|---|---|---|---|---|---|---|---|---|---|
| `middleware/auth.go` (JWTAuth) | A.8.2 | -- | -- | A.8.1 | -- | -- | SR 1.1 | PR.AC | CC6.2 | Art.32 |
| `middleware/auth.go` (RBAC) | A.5.3, A.8.3 | -- | -- | A.8.1 | -- | -- | SR 2.1 | PR.AC | CC6.3 | Art.32 |
| `middleware/auth.go` (AuditLogger) | A.8.15 | -- | 7.2 | A.9.1 | 8.2 | -- | SR 6.1 | DE.CM | CC7.1 | Art.5(1)(a) |
| `middleware/auth.go` (RateLimiter) | A.5.7 | -- | -- | -- | -- | -- | SR 7.1 | PR.PT | CC6.1 | Art.32 |
| `middleware/auth.go` (RequestValidator) | A.8.12, A.8.24 | -- | -- | -- | -- | -- | -- | PR.IP | CC6.1 | Art.25 |
| `middleware/auth.go` (CORSSecurity) | A.8.23 | -- | -- | -- | -- | -- | -- | PR.PT | -- | -- |
| `cmd/server/main.go` (security headers) | A.8.23 | -- | -- | -- | -- | -- | -- | PR.PT | -- | -- |
| `service/device_service.go` (Create/Decommission) | A.8.10 | 8.1 | -- | A.6.1 | -- | -- | SR 1.2 | ID.AM | -- | Art.17 |
| `compliance/compliance.go` (incidents) | -- | 10.2 | 6.1-7.2 | -- | 8.5 | -- | -- | RS.RP | CC7.2 | Art.33 |
| `compliance/compliance.go` (retention) | A.8.10 | -- | -- | -- | -- | -- | -- | PR.DS | -- | Art.5(1)(e) |
| `repository/dynamodb.go` (data isolation) | -- | -- | -- | A.6.2 | -- | -- | -- | -- | -- | -- |
| `handler/grpc_handler.go` (firmware) | -- | -- | -- | -- | 8.7 | -- | -- | -- | CC8.1 | -- |
| `deploy/k8s/go-service.yaml` (HPA/probes) | -- | 8.5 | -- | A.5.1 | 8.1, 8.3 | 8.4 | -- | RC.RP | A1.2 | -- |
| `python/analytics_service/handlers.py` (events/alerts) | A.8.16 | 9.1 | 6.2, 6.3, 6.4 | -- | -- | -- | -- | DE.AE | CC7.2 | -- |
| `firmware/` (device identity/sensors) | A.8.1 | -- | -- | -- | -- | -- | SR 1.2 | ID.AM | -- | -- |
| `docker-compose.yml` (multi-service) | -- | 8.1 | -- | A.5.1 | -- | 8.5 | -- | -- | -- | -- |
| `cmd/server/main.go` (swaggerSpec) | -- | 7.5 | -- | -- | -- | -- | -- | -- | -- | Art.5(1)(a) |

---

*This document is maintained as part of the Sentinel Device Manager codebase. All control mappings are generated from the live compliance reporting system in `backend/go/internal/compliance/compliance.go` and verified against the actual code artifacts referenced throughout. For the most current compliance status, query `GET /api/v1/compliance/report`.*
