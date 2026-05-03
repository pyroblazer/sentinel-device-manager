# Sentinel Device Manager -- User Guide

**Version 1.0.0**

Enterprise physical security device management platform for registering, monitoring, updating, and decommissioning cameras, access control panels, alarm systems, and sensors at scale.

---

## Table of Contents

1. [Why We Built This App](#1-why-we-built-this-app)
2. [How to Profit / Business Model](#2-how-to-profit--business-model)
3. [Features (ELI5)](#3-features-eli5)
4. [User Tutorial (Step by Step)](#4-user-tutorial-step-by-step)
5. [User Manual](#5-user-manual)
6. [API Quick Reference](#6-api-quick-reference)
7. [Troubleshooting](#7-troubleshooting)
8. [FAQ](#8-faq)

---

## 1. Why We Built This App

### The Problem

Managing thousands of physical security devices across multiple sites is chaotic. Security operations teams responsible for enterprise campuses, data centers, government buildings, retail chains, and hospitals face the same painful realities every day:

- **No real-time visibility.** Devices fail silently. A camera goes offline at 2 AM and nobody knows until a guard walks past it on patrol hours later. A broken access control panel at a side entrance sits undetected for days. Teams discover gaps only after an incident has already occurred.

- **Manual spreadsheets everywhere.** Most organizations track their security devices in Excel spreadsheets or shared Google Sheets. One person forgets to update the serial number after a hardware swap, and the entire inventory becomes unreliable. Mergers, site expansions, and staff turnover make this worse over time.

- **Slow incident response.** When a security event occurs -- a forced door, a tamper alert, a camera feed loss -- operators must manually cross-reference spreadsheets, log into multiple vendor-specific portals, and piece together what happened. Minutes turn into hours. In physical security, delayed response can mean property loss, safety risks, or regulatory violations.

- **Compliance requirements demand audit trails.** Standards like ISO 27001, SOC 2, GDPR, and NIST CSF require organizations to maintain detailed records of every device, every configuration change, every access event, and every incident response action. Producing these audit reports manually takes weeks of staff time and is error-prone.

- **Firmware updates are a nightmare.** Deploying firmware across 1,000 cameras from three different vendors means logging into three different management consoles, manually selecting devices, hoping nothing breaks, and having no way to roll back if something does.

### The Solution

Sentinel Device Manager is one platform to register, monitor, update, and decommission every physical security device in your organization. It consolidates device management, real-time alerting, firmware deployment, compliance reporting, and analytics into a single pane of glass.

**Core capabilities:**

- Register any device type (cameras, access control panels, alarms, sensors) with full metadata tracking
- Monitor device health in real time (CPU, memory, temperature, network latency, heartbeat status)
- Receive instant alerts when critical events occur, with severity classification and acknowledgment workflows
- Deploy firmware updates across hundreds or thousands of devices with staged rollouts and rollback support
- Generate compliance reports for 10 international standards automatically
- Organize devices into groups, zones, and geofences for logical management
- Integrate with other systems via webhooks, API keys, and a full REST API

### Return on Investment

| Metric | Before Sentinel | After Sentinel |
|--------|----------------|---------------|
| Device downtime detection | Hours to days | Seconds to minutes |
| Manual inventory work | 20+ hours/week | 4 hours/week (80% reduction) |
| Compliance audit preparation | 2-3 weeks | Automated, on-demand |
| Firmware deployment time | Days per vendor | Hours across all vendors |
| Incident response time | 30-60 minutes | Under 5 minutes |

---

## 2. How to Profit / Business Model

### SaaS Subscription Pricing

Sentinel Device Manager is offered as a Software-as-a-Service platform with tiered per-device monthly pricing:

| Tier | Price per Device/Month | Included Features |
|------|----------------------|-------------------|
| **Starter** | $5 | Device management, basic alerts, dashboard |
| **Professional** | $15 | All Starter features + firmware management, groups/zones, templates, bulk operations |
| **Enterprise** | $30 | All Professional features + compliance reporting, geofencing, webhooks, API keys |
| **Premium** | $50 | All Enterprise features + advanced analytics, priority support, custom integrations |

### Enterprise Licenses

For organizations managing 10,000+ devices across multiple sites, we offer annual enterprise licenses with volume discounts, dedicated infrastructure, custom SLAs, and white-labeling options. Contact sales for custom pricing.

### Professional Services

| Service | Rate |
|---------|------|
| Installation and deployment | $150/hour |
| Custom integration development | $200/hour |
| Training workshops (on-site) | $250/hour |
| Compliance audit preparation | $300/hour |
| Custom report development | $175/hour |

### Compliance-as-a-Service (Premium Add-On)

Automated audit report generation for ISO 27001, SOC 2, GDPR, NIST CSF, and 6 additional standards. Produces audit-ready documentation with evidence mapping, control status tracking, and incident timeline reports. Offered as a premium add-on at $10/device/month on top of any tier.

### Marketplace

A future marketplace for third-party integrations connecting Sentinel with:

- Access control vendors (HID, Lenel, Gallagher)
- Camera manufacturers (Axis, Hanwha, Bosch, Hikvision)
- SIEM platforms (Splunk, QRadar, SentinelOne)
- Building management systems
- Ticketing systems (ServiceNow, Jira)

Marketplace partners pay a 15-20% revenue share on integrations sold through the platform.

### Data Insights

Anonymized, aggregated industry benchmarks available to premium subscribers who opt in. Compare your device uptime, firmware compliance, alert response times, and incident rates against industry peers. Premium feature at $5/device/month.

### Target Customers

- **Enterprise campuses** -- corporate headquarters with 500-5,000 security devices
- **Data centers** -- colocation providers requiring strict physical security compliance
- **Government buildings** -- federal, state, and local facilities with compliance mandates
- **Retail chains** -- nationwide store networks with camera and alarm systems
- **Hospitals and healthcare** -- patient safety and HIPAA-related physical security
- **Critical infrastructure** -- power plants, water treatment, transportation hubs

### Revenue Projections

| Milestone | Devices | Avg Revenue/Device | Monthly Recurring Revenue |
|-----------|---------|-------------------|--------------------------|
| Year 1 Q2 | 1,000 | $20 | $20,000 MRR |
| Year 1 Q4 | 5,000 | $22 | $110,000 MRR |
| Year 2 Q2 | 15,000 | $25 | $375,000 MRR |
| Year 2 Q4 | 30,000 | $28 | $840,000 MRR |
| Year 3+ | 50,000+ | $30+ | $1,500,000+ MRR |

Professional services and compliance add-ons contribute an additional 20-30% on top of subscription revenue.

---

## 3. Features (ELI5)

Explain Like I'm 5 -- simple, jargon-free descriptions of every feature.

### Device Management

Like a contacts app, but for security cameras and alarms. You can see which ones are working and which need fixing. Every device gets a profile with its name, type, location, and health status. You can add new devices, update their information, or remove them when they are no longer needed.

### Alerts

When something bad happens, the app shouts at you immediately instead of finding out hours later. Critical events like a forced door, a camera going offline, or a tamper detection automatically create alerts. You can see all active alerts, acknowledge them to say "I am working on this," and mark them resolved when fixed.

### Firmware

Like updating your phone, but for 1,000 security cameras at once. We make sure every device runs the latest, safest software. Deploy firmware updates to selected devices in stages -- update 25% first, make sure nothing breaks, then continue with the rest. If something goes wrong, you can roll back. Every firmware version includes a SHA-256 checksum so you know the software has not been tampered with.

### Analytics

A report card showing how healthy your security system is. The analytics summary tells you how many devices you have, how many are online versus offline, how many alerts are active, how many events happened in the last 24 hours, and what your firmware compliance percentage is. It gives you the big picture at a glance.

### Compliance

A robot that checks you are following all the security rules, so auditors stay happy. Sentinel tracks your compliance across 10 international standards including ISO 27001, SOC 2, GDPR, NIST CSF, and more. Each standard has specific controls, and we show you which ones are implemented, partial, or planned. You can generate a full compliance report with one click.

### Groups and Zones

Put devices in folders, like organizing apps on your phone. Create groups called "Building A - Floor 3" or "Perimeter Cameras" and add devices to them. Groups can be zones, buildings, floors, or custom categories. This makes it easy to find and manage related devices together.

### Templates

Save device settings as a recipe, then apply the same recipe to 100 devices at once. If all your lobby cameras should have the same resolution, retention days, and motion sensitivity, create a template once and reuse it. No more configuring each device individually.

### Bulk Operations

Do the same thing to many devices at once, instead of one by one. Select 50 devices and update their status, delete them, or tag them with a single operation. You can also export your device list as JSON or CSV for use in other tools.

### Webhooks

Tell other apps what is happening in Sentinel, so they can react automatically. Create a webhook pointing to your Slack channel, PagerDuty account, or custom script, and Sentinel will send a message whenever a subscribed event occurs (device created, alert triggered, firmware updated, etc.). Test your webhooks with a single click.

### API Keys

A special password for other programs to talk to Sentinel safely. Generate API keys for scripts, automation tools, or third-party integrations. Each key has a role (admin, operator, viewer) and optional expiration date. Keys are shown only once at creation, so store them securely.

### Notifications

A mailbox that only shows important security messages. The notification center collects system messages about group creation, webhook activity, and other events. View all notifications, mark them as read, or clear them when you are done.

### Geofencing

Draw a circle on a map and say "these cameras belong to this building." Create geographic zones by specifying a center point (latitude and longitude) and a radius in meters, then assign devices to that zone. Useful for large campuses where physical location matters.

### OWASP Protection

An invisible shield that stops hackers from breaking in through the website. Sentinel implements protections against all 10 categories of the OWASP Top 10 (2021), including input sanitization to block injection attacks, brute-force protection to stop password guessing, security headers to prevent clickjacking and XSS, SSRF protection to block requests to internal services, rate limiting to prevent abuse, and tenant isolation to keep each organization's data separate.

---

## 4. User Tutorial (Step by Step)

Your first 15 minutes with Sentinel Device Manager.

### Prerequisites

Make sure the platform is running. Start it with a single command:

```bash
make dev
```

Or with Docker Compose directly:

```bash
docker compose up --build
```

This starts:
- Go service (device management) on port 8080
- Python service (analytics) on port 8081
- Frontend on port 3000
- DynamoDB local on port 8000
- Firmware simulator (auto-registers a test device)

Open your browser and navigate to `http://localhost:3000`.

### Step 1: Dashboard Overview

When you first log in, you see the Dashboard. This is your command center.

- **Total Devices** card shows how many devices are registered
- **Online / Offline** cards give you a quick health check
- **Active Alerts** card shows how many alerts need attention
- **Events (24h)** card shows recent activity volume
- **Firmware Compliance** card shows what percentage of devices run approved firmware versions
- **Recent Devices** table lists the latest registered devices with type, status, and site

If the dashboard is empty, that is expected -- you have not registered any devices yet.

### Step 2: Register Your First Device

1. Click **Devices** in the left navigation menu.
2. Click the **Register Device** button (top right).
3. Fill in the registration form:
   - **Serial Number** (required): Enter the device serial, e.g. `VKD-CAM-001`
   - **Device Type** (required): Select `CAMERA`, `ACCESS_CONTROL`, `ALARM`, or `SENSOR`
   - **Model**: Enter the hardware model, e.g. `D30`
   - **Site ID** (required): Enter or select the site, e.g. `site-hq-lobby`
   - **Organization ID** (required): Enter your organization ID, e.g. `org-acme`
   - **IP Address**: Enter the device IP, e.g. `192.168.1.100`
   - **MAC Address**: Enter the device MAC, e.g. `AA:BB:CC:DD:EE:FF`
   - **Config**: Add optional key-value configuration pairs (e.g. `resolution: 4K`)
4. Click **Register**.

The device appears in the devices list with status `ONLINE` and firmware version `0.0.0` (ready for its first update).

### Step 3: View Device Details

1. From the devices list, click on the device you just registered.
2. The device detail page shows:
   - **Identity**: device ID (UUID), serial number, model
   - **Type and Status**: device type, current status, firmware version
   - **Location**: site ID, organization ID
   - **Network**: IP address, MAC address
   - **Health**: CPU usage, memory usage, temperature, uptime, network latency
   - **Configuration**: key-value config pairs
   - **Activity Timeline**: recent events and alerts for this device
   - **Last Heartbeat**: when the device last checked in

### Step 4: Check Alerts

1. Click **Alerts** in the left navigation menu.
2. The alerts list shows all alerts with:
   - Severity badge (INFO, WARNING, CRITICAL)
   - Status badge (ACTIVE, ACKNOWLEDGED, RESOLVED)
   - Alert type and message
   - Associated device
   - Creation timestamp
3. Filter alerts by **Status** (ACTIVE, ACKNOWLEDGED, RESOLVED) or **Severity** (INFO, WARNING, CRITICAL).
4. To acknowledge an alert, click the **Acknowledge** button and enter your name. This tells your team you are handling it.
5. Critical events (like `DOOR_FORCED`, `TAMPER_DETECTED`, `DEVICE_OFFLINE`) automatically generate alerts.

### Step 5: Create a Device Group

1. Click **Groups** in the left navigation menu.
2. Click **Create Group**.
3. Fill in:
   - **Name**: e.g. `Lobby Cameras`
   - **Type**: Select `ZONE`, `BUILDING`, `FLOOR`, or `CUSTOM`
   - **Site ID**: e.g. `site-hq`
4. Click **Create**.
5. Open the newly created group and click **Add Devices**.
6. Select devices to add to the group by their device IDs.
7. Click **Add**.

You now have a logical grouping you can use for filtering, bulk operations, and reporting.

### Step 6: Set Up a Notification Webhook

1. Click **Admin Panel** in the left navigation menu, then **Webhooks**.
2. Click **Create Webhook**.
3. Fill in:
   - **Name**: e.g. `Slack Security Alerts`
   - **URL**: Your webhook endpoint URL, e.g. `https://hooks.slack.com/services/T00/B00/xxx`
   - **Events**: Select which events to subscribe to (e.g. `ALERT_TRIGGERED`, `DEVICE_CREATED`, `FIRMWARE_UPDATED`)
   - **Headers**: Add custom HTTP headers if needed (e.g. `Authorization: Bearer xxx`)
4. Click **Create**.
5. Click **Test** next to the webhook to send a test payload and verify the connection works.

### Step 7: View Compliance Report

1. Click **Compliance** in the left navigation menu.
2. The compliance dashboard shows:
   - **Overall Compliance Percentage**: aggregated across all standards
   - **Standards List**: each tracked standard with its control count and compliance percentage
   - **Control Details**: expand any standard to see each control, its status (IMPLEMENTED, PARTIAL, PLANNED), and evidence
3. Click **Generate Report** to download a full compliance report including all standards, controls, data retention policies, and incident counts.

---

## 5. User Manual

Complete reference for every page and feature.

### 5.1 Dashboard

The Dashboard is the landing page after login. It provides a high-level overview of your security infrastructure.

**Stat Cards (top row):**

| Card | Description |
|------|-------------|
| Total Devices | Count of all registered devices across all sites |
| Online Devices | Count of devices currently reporting heartbeats (status = ONLINE) |
| Offline Devices | Count of devices not reporting (status = OFFLINE) |
| Active Alerts | Count of alerts in ACTIVE status that have not been acknowledged |
| Events (24h) | Total number of events ingested in the last 24 hours |
| Critical Alerts | Count of alerts with CRITICAL severity that are still active |
| Firmware Compliance | Percentage of devices running approved firmware versions |

**Recent Devices Table:**

Displays the most recently registered or updated devices with columns:
- Device ID (truncated UUID, clickable to view details)
- Serial Number
- Device Type (CAMERA, ACCESS_CONTROL, ALARM, SENSOR)
- Status (ONLINE, OFFLINE, MAINTENANCE, DECOMMISSIONED)
- Site ID
- Last Updated timestamp

The analytics summary data is served by the Python analytics service at `GET /api/v1/analytics/summary`.

### 5.2 Devices

The Devices page is the primary interface for managing your device fleet.

**List View:**

- Displays all devices in a paginated table (default: 50 per page)
- Click column headers to sort
- Use the filter bar to narrow results:
  - **Type**: CAMERA, ACCESS_CONTROL, ALARM, SENSOR
  - **Status**: ONLINE, OFFLINE, MAINTENANCE, DECOMMISSIONED
  - **Site ID**: filter by site
  - **Organization ID**: filter by organization

**Register a Device (Create):**

Required fields:
- `serial_number` -- unique hardware identifier (e.g. `VKD-CAM-001`)
- `device_type` -- one of `CAMERA`, `ACCESS_CONTROL`, `ALARM`, `SENSOR`
- `site_id` -- the site where the device is installed
- `organization_id` -- the organization that owns the device

Optional fields:
- `model` -- hardware model name
- `ip_address` -- current IP address
- `mac_address` -- MAC address
- `config` -- key-value configuration pairs

When a device is created:
- A UUID is generated as the `device_id`
- Initial status is set to `ONLINE`
- Initial firmware version is `0.0.0`
- `created_at` and `updated_at` timestamps are set to the current UTC time
- `last_heartbeat` is set to the current UTC time

**View Device Details:**

Click any device in the list to open its detail page showing all fields including:
- Identity: `device_id`, `serial_number`, `model`
- Type and status: `device_type`, `status`, `firmware_version`
- Location: `site_id`, `organization_id`
- Network: `ip_address`, `mac_address`
- Health telemetry: `cpu_usage`, `memory_usage`, `temperature_c`, `uptime_seconds`, `network_latency_ms`
- Configuration: all key-value config pairs
- Timestamps: `last_heartbeat`, `created_at`, `updated_at`

**Edit a Device (Update):**

Supports partial updates -- only send the fields you want to change. Updatable fields:
- `device_type` -- change the device type
- `model` -- update the model name
- `firmware_version` -- update the firmware version (usually done via firmware deployment)
- `status` -- change status (ONLINE, OFFLINE, MAINTENANCE, DECOMMISSIONED)
- `site_id` -- move device to a different site
- `ip_address` -- update IP address
- `config` -- replace configuration key-value pairs

The `updated_at` timestamp is automatically refreshed on every update.

**Delete a Device:**

Permanently removes a device from the system. This action cannot be undone. For GDPR-compliant removal, decommission the device first, then delete it after the retention period (180 days for decommissioned devices).

**Decommission a Device:**

Sets the device status to `DECOMMISSIONED` while preserving the record for audit purposes during the data retention period. This is the recommended approach for GDPR compliance (Article 17, Right to Erasure). The device record is retained for 180 days before permanent deletion.

### 5.3 Alerts

The Alerts page shows all system alerts generated from device events.

**Alert Lifecycle:**

```
ACTIVE --> ACKNOWLEDGED --> RESOLVED
```

1. An alert is created automatically when a critical event is ingested
2. An operator acknowledges the alert to claim responsibility
3. The alert is resolved when the underlying issue is fixed

**List Alerts:**

- Filter by **Status**: ACTIVE, ACKNOWLEDGED, RESOLVED
- Filter by **Severity**: INFO, WARNING, CRITICAL
- Default sort: newest first
- Default limit: 50 alerts per page

**Alert Fields:**

| Field | Description |
|-------|-------------|
| alert_id | Unique identifier (UUID) |
| device_id | The device that triggered the alert |
| event_id | The event that caused the alert |
| alert_type | Category of the alert (e.g. DEVICE_OFFLINE, DOOR_FORCED) |
| severity | INFO, WARNING, or CRITICAL |
| status | ACTIVE, ACKNOWLEDGED, or RESOLVED |
| message | Human-readable description of the alert |
| acknowledged_by | Username of the person who acknowledged the alert |
| created_at | When the alert was generated |
| updated_at | When the alert was last modified |

**Acknowledge an Alert:**

Click the acknowledge button and provide your user identifier. This:
- Changes status from ACTIVE to ACKNOWLEDGED
- Records who acknowledged it
- Updates the `updated_at` timestamp

**Event Types That Generate Alerts:**

Critical-severity events automatically generate alerts. Event types include:
- `MOTION_DETECTED` -- motion sensor triggered
- `DOOR_OPENED` -- access control door opened
- `DOOR_FORCED` -- door forced open without authorization
- `ALARM_TRIGGERED` -- alarm system activated
- `TEMPERATURE_THRESHOLD` -- temperature exceeded safe range
- `DEVICE_OFFLINE` -- device stopped reporting
- `DEVICE_ONLINE` -- device came back online
- `FIRMWARE_UPDATED` -- firmware deployment completed
- `TAMPER_DETECTED` -- physical tampering detected

### 5.4 Firmware

The Firmware page manages firmware versions and deployments across your device fleet.

**Firmware Versions:**

Each firmware version record includes:
- `version` -- semantic version string (e.g. `1.2.3`)
- `device_type` -- which device type this firmware applies to
- `binary_url` -- URL to download the firmware binary
- `checksum_sha256` -- SHA-256 hash for integrity verification
- `release_notes` -- description of changes
- `created_at` -- when this version was registered

**Deploy Firmware:**

1. Select a firmware version
2. Choose target devices (by device IDs or by group)
3. Configure staged rollout (optional):
   - `percentage_per_stage` -- what percentage of devices to update per stage (e.g. 25%)
   - `delay_between_stages_minutes` -- wait time between stages (e.g. 30 minutes)
4. Start the deployment

**Deployment Status Lifecycle:**

```
PENDING --> DOWNLOADING --> VERIFYING --> APPLYING --> COMPLETED
                                                      --> FAILED --> ROLLED_BACK
```

Each deployment tracks:
- `deployment_id` -- unique identifier
- `version` -- firmware version being deployed
- `device_ids` -- list of target devices
- `status` -- current deployment status
- `created_at` and `updated_at` -- timestamps

**Integrity Verification:**

Every firmware binary includes a SHA-256 checksum. Before applying a firmware update, the system verifies that the downloaded binary matches the expected checksum. This implements ISO 27001 A.14.1 (secure development) and OWASP A08 (software and data integrity failures).

**gRPC Firmware Streaming:**

Devices can receive firmware updates via gRPC streaming. The firmware binary is sent in 64 KB chunks, each with its own checksum. This enables efficient updates even for devices on slow or unreliable network connections.

### 5.5 Groups

Groups let you organize devices into logical collections for easier management.

**Group Types:**

| Type | Description | Example |
|------|-------------|---------|
| ZONE | A geographic or functional zone | "Perimeter Zone" |
| BUILDING | A physical building | "Headquarters" |
| FLOOR | A floor within a building | "HQ - Floor 3" |
| CUSTOM | Any custom grouping | "Night Shift Cameras" |

**Create a Group:**

Provide:
- `name` (required) -- group name
- `type` -- group type (ZONE, BUILDING, FLOOR, CUSTOM)
- `site_id` -- associated site

**Manage Group Devices:**

- View all devices in a group
- Add devices to a group (by device ID)
- Remove devices from a group
- Rename a group
- Delete a group (devices are not deleted, only the grouping)

### 5.6 Templates

Templates store reusable device configurations that can be applied to multiple devices at once.

**Create a Template:**

Provide:
- `name` (required) -- template name (e.g. "Lobby Camera Standard Config")
- `device_type` -- which device type this template applies to
- `config` -- key-value configuration pairs (e.g. `{"resolution": "4K", "retention_days": "30", "motion_sensitivity": "high"}`)
- `description` -- human-readable description

**Use a Template:**

1. Browse templates and select one that matches your device type
2. Apply the template to one or more devices via bulk operations
3. The template's config key-value pairs overwrite the device's existing config

**Manage Templates:**

- List all templates
- View template details
- Delete templates you no longer need

### 5.7 Bulk Operations

Perform actions on multiple devices simultaneously.

**Available Bulk Operations:**

| Operation | Endpoint | Description |
|-----------|----------|-------------|
| Bulk Delete | `POST /api/v1/devices/bulk-delete` | Delete multiple devices by ID |
| Bulk Update | `POST /api/v1/devices/bulk-update` | Apply the same updates to multiple devices |
| Bulk Tag | `POST /api/v1/devices/bulk-tag` | Add tags to multiple devices |
| Export | `POST /api/v1/devices/export` | Export device list as JSON or CSV |

**How to Use Bulk Operations:**

1. Select devices from the device list (or provide device IDs)
2. Choose the operation
3. For bulk update: provide the fields to update
4. For bulk tag: provide the tags to apply
5. For export: choose JSON or CSV format
6. Execute -- operations are queued and processed asynchronously

**Export Formats:**

- **JSON**: Full device data with metadata and timestamp
- **CSV**: Flat file with columns `device_id,serial_number,device_type,status,site_id`

### 5.8 Admin Panel

The Admin Panel provides access to system configuration features.

#### 5.8.1 API Keys

API keys allow external programs and scripts to authenticate with Sentinel.

**Create an API Key:**

Provide:
- `name` -- descriptive name (e.g. "Monitoring Script")
- `role` -- one of `admin`, `operator`, `viewer`
  - `admin`: full read/write access to all endpoints
  - `operator`: read access + write access to device management
  - `viewer`: read-only access
- `org_id` -- organization this key belongs to
- `expires_in_seconds` (optional) -- how long until the key expires

The key is prefixed with `sk-sentinel-` and is only shown once at creation time. Store it securely.

**List API Keys:**

All keys are listed with their metadata. The actual key value is redacted (only the first 15 characters are shown).

**Delete an API Key:**

Immediately revokes access. Any scripts using this key will receive authentication errors.

#### 5.8.2 Webhooks

Webhooks send HTTP POST requests to external URLs when subscribed events occur in Sentinel.

**Create a Webhook:**

Provide:
- `name` -- descriptive name (e.g. "Slack Notifications")
- `url` (required) -- the endpoint URL to send payloads to
- `events` -- list of event types to subscribe to (e.g. `DEVICE_CREATED`, `ALERT_TRIGGERED`, `FIRMWARE_UPDATED`)
- `headers` -- custom HTTP headers to include in requests (useful for authentication)

**Manage Webhooks:**

- List all webhooks with their status (active/inactive)
- Test a webhook (sends a test payload to verify the connection)
- Delete a webhook

**Security Note:** Sentinel validates webhook URLs against an SSRF (Server-Side Request Forgery) protection blocklist. URLs targeting `localhost`, `127.0.0.1`, AWS metadata endpoints (`169.254.169.254`), and internal service names are rejected.

#### 5.8.3 Geofences

Geofences define geographic boundaries for device placement and management.

**Create a Geofence:**

Provide:
- `name` -- geofence name (e.g. "Building A Perimeter")
- `center_lat` -- latitude of the center point
- `center_lng` -- longitude of the center point
- `radius_meters` -- radius of the geofence circle in meters
- `site_id` -- associated site

**Manage Geofences:**

- List all geofences
- Delete geofences

### 5.9 Notifications

The notification center shows system messages about platform activity.

**Notification Types:**

| Type | Description |
|------|-------------|
| ALERT | Security alert notifications |
| SYSTEM | System-level messages (group created, webhook configured, etc.) |
| FIRMWARE | Firmware deployment status updates |
| MAINTENANCE | Scheduled maintenance notifications |

**Managing Notifications:**

- List all notifications (newest first)
- Mark individual notifications as read
- Delete notifications you no longer need

Notifications are automatically generated by system actions (e.g. creating a group generates a SYSTEM notification).

### 5.10 Compliance

The Compliance page tracks adherence to 10 international security and quality standards.

**Tracked Standards:**

| Standard | Full Name | Controls |
|----------|-----------|----------|
| ISO 27001 | ISO/IEC 27001:2022 | 12 controls |
| ISO 9001 | ISO 9001:2015 | 8 controls |
| ISO 27035 | ISO/IEC 27035:2023 | 6 controls |
| ISO 27017 | ISO/IEC 27017:2015 | 6 controls |
| ISO 20000 | ISO/IEC 20000-1:2018 | 7 controls |
| ISO 22301 | ISO 22301:2019 | 5 controls |
| IEC 62443 | IEC 62443 | 8 controls |
| NIST CSF | NIST Cybersecurity Framework 2.0 | 10 controls |
| SOC 2 | SOC 2 Type II (2017) | 7 controls |
| GDPR | EU Regulation 2016/679 | 9 controls |

**Control Statuses:**

| Status | Meaning |
|--------|---------|
| IMPLEMENTED | Fully in place with evidence |
| PARTIAL | Partially implemented, work remaining |
| PLANNED | Designed but not yet implemented |
| NOT_APPLICABLE | Not relevant to this deployment |

**Compliance Report:**

The full compliance report includes:
- All standards with every control, its status, and supporting evidence
- Summary statistics (total controls, implemented count, compliance percentage)
- Data retention policies with GDPR legal basis citations
- Open incident count and total incident count
- Generation timestamp and platform version

**Incident Management (ISO 27035):**

Create, track, and resolve security and operational incidents:
- Create an incident with title, description, severity (LOW, MEDIUM, HIGH, CRITICAL), and category (SECURITY, OPERATIONAL, COMPLIANCE, DATA)
- Incident lifecycle: OPEN -> INVESTIGATING -> CONTAINED -> RESOLVED -> CLOSED
- Each status change adds a timeline entry with who made the change and notes
- Filter incidents by status

**Data Retention Policies:**

| Data Category | Retention Period | GDPR Basis |
|---------------|-----------------|------------|
| Device data | 365 days | Legitimate interest (Art.6(1)(f)) |
| Audit logs | 730 days | Legal obligation (Art.6(1)(c)) |
| Security events | 365 days | Legitimate interest (Art.6(1)(f)) |
| Incident records | 1,095 days (3 years) | Legal obligation (Art.6(1)(c)) |
| Firmware versions | 1,825 days (5 years) | Legitimate interest (Art.6(1)(f)) |
| Analytics data | 90 days | Consent (Art.6(1)(a)) |
| Decommissioned devices | 180 days | Legitimate interest (Art.6(1)(f)) |

---

## 6. API Quick Reference

### Authentication

All device management and superapp endpoints require JWT Bearer token authentication (unless `AUTH_ENABLED=false` in development):

```
Authorization: Bearer <jwt-token>
```

JWT tokens are HMAC-SHA256 (HS256) signed and contain claims for `sub` (user ID), `role` (admin/operator/viewer), `org_id` (organization), and `scopes`.

Compliance endpoints (`/api/v1/compliance/*`) and the health check (`/health`) do not require authentication.

### Go Service -- Device Management (Port 8080)

#### Device CRUD

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/devices` | List devices with optional filters (type, status, site_id, organization_id, page, limit) |
| `POST` | `/api/v1/devices` | Register a new device |
| `GET` | `/api/v1/devices/{deviceID}` | Get device details by ID |
| `PUT` | `/api/v1/devices/{deviceID}` | Update device (partial update supported) |
| `DELETE` | `/api/v1/devices/{deviceID}` | Delete a device permanently |

#### Compliance

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/compliance/report` | Full compliance report across all standards |
| `GET` | `/api/v1/compliance/standards` | List all tracked compliance standards |
| `GET` | `/api/v1/compliance/retention` | Data retention policies (GDPR) |
| `GET` | `/api/v1/compliance/incidents` | List incidents (filterable by status) |
| `POST` | `/api/v1/compliance/incidents` | Create a new incident |
| `PUT` | `/api/v1/compliance/incidents/{incidentID}` | Update incident status with timeline entry |

#### Security

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | System health check (no auth required) |
| `GET` | `/api/v1/security/owasp` | OWASP Top 10 (2021) controls status |
| `GET` | `/swagger.json` | OpenAPI 3.0 specification |

#### Superapp Features (Groups, Templates, Webhooks, API Keys, Notifications, Geofences, Bulk Ops)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/groups` | List all device groups |
| `POST` | `/api/v1/groups` | Create a device group |
| `GET` | `/api/v1/groups/{groupID}` | Get group details |
| `PUT` | `/api/v1/groups/{groupID}` | Update group name |
| `DELETE` | `/api/v1/groups/{groupID}` | Delete a group |
| `POST` | `/api/v1/groups/{groupID}/devices` | Add devices to a group |
| `GET` | `/api/v1/templates` | List all configuration templates |
| `POST` | `/api/v1/templates` | Create a configuration template |
| `GET` | `/api/v1/templates/{templateID}` | Get template details |
| `DELETE` | `/api/v1/templates/{templateID}` | Delete a template |
| `POST` | `/api/v1/devices/bulk-delete` | Bulk delete devices |
| `POST` | `/api/v1/devices/bulk-update` | Bulk update devices |
| `POST` | `/api/v1/devices/bulk-tag` | Bulk tag devices |
| `POST` | `/api/v1/devices/export` | Export devices (JSON or CSV) |
| `GET` | `/api/v1/webhooks` | List all webhooks |
| `POST` | `/api/v1/webhooks` | Create a webhook |
| `DELETE` | `/api/v1/webhooks/{webhookID}` | Delete a webhook |
| `POST` | `/api/v1/webhooks/{webhookID}/test` | Send test payload to webhook |
| `GET` | `/api/v1/api-keys` | List API keys (redacted) |
| `POST` | `/api/v1/api-keys` | Create a new API key |
| `DELETE` | `/api/v1/api-keys/{keyID}` | Revoke an API key |
| `GET` | `/api/v1/notifications` | List all notifications |
| `POST` | `/api/v1/notifications/{notificationID}/read` | Mark notification as read |
| `DELETE` | `/api/v1/notifications/{notificationID}` | Delete a notification |
| `GET` | `/api/v1/geofences` | List all geofence zones |
| `POST` | `/api/v1/geofences` | Create a geofence zone |
| `DELETE` | `/api/v1/geofences/{zoneID}` | Delete a geofence zone |

### Python Service -- Analytics (Port 8081)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Analytics service health check |
| `GET` | `/api/v1/events` | List events (filterable by device_id, severity, limit) |
| `POST` | `/api/v1/events` | Ingest a new event (auto-generates alert for CRITICAL severity) |
| `GET` | `/api/v1/alerts` | List alerts (filterable by status, severity, limit) |
| `POST` | `/api/v1/alerts/{alert_id}/acknowledge` | Acknowledge an alert with user identifier |
| `GET` | `/api/v1/analytics/summary` | Dashboard summary statistics |

### gRPC Service -- Device Communication (Port 9090)

| RPC | Request | Response | Description |
|-----|---------|----------|-------------|
| `Register` | `RegisterRequest` | `RegisterResponse` | Register a new device via gRPC |
| `SendHeartbeat` | `HeartbeatRequest` | `HeartbeatResponse` | Device heartbeat with config sync |
| `SendEvent` | `EventRequest` | `EventResponse` | Send a device event |
| `StreamFirmware` | `FirmwareRequest` | `stream FirmwareChunk` | Stream firmware binary in 64KB chunks |
| `ReportFirmwareStatus` | `FirmwareStatusRequest` | `FirmwareStatusResponse` | Report firmware deployment status |
| `GetConfig` | `ConfigRequest` | `ConfigResponse` | Retrieve device configuration |

---

## 7. Troubleshooting

### Device Will Not Register

**Symptoms:** HTTP 400 error when calling `POST /api/v1/devices`.

**Check the following:**

1. **Required fields are present.** The endpoint requires `serial_number`, `device_type`, `site_id`, and `organization_id`. All four must be non-empty strings.
2. **Serial number format.** The serial number must be a non-empty string. Common formats are `VKD-CAM-001` or `ACM-DOOR-042`. Avoid special characters like `$`, `{`, `<script>` which are blocked by the input sanitization middleware.
3. **Device type is valid.** Must be exactly one of: `CAMERA`, `ACCESS_CONTROL`, `ALARM`, `SENSOR`. Case-sensitive.
4. **Request body is valid JSON.** Ensure the `Content-Type: application/json` header is set and the body is valid JSON.
5. **Request body size.** The maximum request body is 1 MB. If your config object is very large, reduce it.

**Example valid request:**

```bash
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "serial_number": "VKD-CAM-001",
    "device_type": "CAMERA",
    "model": "D30",
    "site_id": "site-hq",
    "organization_id": "org-acme",
    "ip_address": "192.168.1.100"
  }'
```

### Cannot See Alerts

**Symptoms:** The alerts list is empty or the analytics service returns errors.

**Check the following:**

1. **Python service is running.** Alerts are managed by the Python analytics service on port 8081. Verify it is healthy:
   ```bash
   curl http://localhost:8081/health
   ```
   Expected response: `{"status":"ok"}`

2. **Events have been ingested.** Alerts are only generated when critical-severity events are posted. Ingest a test event:
   ```bash
   curl -X POST http://localhost:8081/api/v1/events \
     -H "Content-Type: application/json" \
     -d '{
       "device_id": "test-device-001",
       "event_type": "DOOR_FORCED",
       "severity": "CRITICAL",
       "payload": {"location": "Building A"}
     }'
   ```

3. **Check filters.** If you are filtering by status or severity, verify the filter values match existing alerts. Use `GET /api/v1/alerts` without filters to see all alerts.

### Authentication Errors

**Symptoms:** HTTP 401 Unauthorized or 403 Forbidden responses.

**Check the following:**

1. **JWT_SECRET_KEY matches.** The JWT secret used to sign tokens must match the `JWT_SECRET_KEY` environment variable on the Go service. In `docker-compose.yml`, both the token issuer and the service must use the same secret.

2. **Token format.** The Authorization header must be: `Bearer <token>`. Do not include "JWT" or other prefixes.

3. **Token expiration.** Tokens have an `exp` claim. Check if your token has expired. Tokens with `exp` in the past are rejected.

4. **Role permissions.** The RBAC middleware enforces role-based access:
   - `admin`: all methods on all paths
   - `operator`: GET and POST on all paths, PUT only on `/api/v1/devices`
   - `viewer`: GET only on all paths
   If you get 403 errors, your role may not have permission for the requested method and path.

5. **AUTH_ENABLED setting.** In development, you can disable auth by setting `AUTH_ENABLED=false` in the environment. This skips JWT and RBAC middleware entirely.

### Slow Dashboard Loading

**Symptoms:** Dashboard takes more than a few seconds to load.

**Check the following:**

1. **DynamoDB connection.** The Go service reads device data from DynamoDB. If DynamoDB is unreachable or slow, device listing will be slow. Verify the connection:
   ```bash
   curl http://localhost:8000/shell/
   ```
   If using DynamoDB Local via Docker, ensure the container is healthy:
   ```bash
   docker-compose ps dynamodb
   ```

2. **DYNAMODB_ENDPOINT.** Ensure this environment variable points to the correct DynamoDB endpoint. In Docker Compose, it should be `http://dynamodb:8000`. For local development outside Docker, use `http://localhost:8000`.

3. **Large device counts.** The List devices endpoint uses DynamoDB Scan, which reads every item and applies filters in memory. For very large tables (100,000+ devices), consider paginating with smaller `limit` values.

4. **Rate limiting.** If you are making many rapid requests, you may hit the rate limiter (100 requests per minute per IP by default). The response will be HTTP 429 with a `Retry-After` header.

### CORS Errors

**Symptoms:** Browser console shows "CORS policy" errors; frontend cannot reach the API.

**Check the following:**

1. **CORS_ALLOWED_ORIGINS.** Update this environment variable to include your frontend's origin. In `docker-compose.yml`:
   ```yaml
   CORS_ALLOWED_ORIGINS=http://localhost:3000
   ```
   For multiple origins, separate with commas. Use `*` for development only (not production).

2. **Preflight requests.** The CORS middleware handles OPTIONS requests automatically. Ensure no other middleware or proxy is stripping the `Access-Control-Allow-*` headers.

3. **Credentials.** The CORS configuration sets `AllowCredentials: true`. When credentials are enabled, `Access-Control-Allow-Origin` cannot be `*`. You must specify explicit origins.

### Brute Force Lockout

**Symptoms:** HTTP 429 "account temporarily locked due to too many failed attempts."

**Resolution:**

This is the brute-force protection kicking in after 5 failed authentication attempts within 15 minutes from the same IP. The lockout lasts 30 minutes.

- Wait 30 minutes for the ban to expire, or
- Restart the Go service to clear in-memory attempt records, or
- If testing, use correct credentials or set `AUTH_ENABLED=false`

### Input Validation Errors

**Symptoms:** HTTP 400 "request contains disallowed pattern" or "request contains potentially unsafe content."

**Resolution:**

The input sanitization middleware blocks request bodies and URL paths containing patterns associated with:
- NoSQL injection (`$where`, `$gt`, `$lt`, `$ne`, `$regex`, `$expr`)
- XSS attacks (`<script`, `javascript:`, `on*=` event handlers, `eval()`)
- Path traversal (`../`, `..\\`, URL-encoded variants)

If you have legitimate data that matches these patterns (e.g. a config value containing `$ne`), use a different naming convention or encoding.

---

## 8. FAQ

**Q1: What types of devices can Sentinel manage?**

Sentinel manages four categories of physical security devices: cameras (CAMERA), access control panels (ACCESS_CONTROL), alarm systems (ALARM), and environmental sensors (SENSOR). Each device type has the same core data model (serial number, model, firmware, status, config) but can be configured differently using templates.

---

**Q2: How do devices communicate with Sentinel?**

Devices communicate through two channels:
- **REST API** (port 8080): Used by operators and integrations for device CRUD, group management, firmware deployment, and configuration.
- **gRPC** (port 9090): Used by devices themselves for registration, heartbeat reporting, event submission, firmware streaming, and config retrieval. gRPC is more efficient for high-frequency device communication.

---

**Q3: What compliance standards does Sentinel support?**

Sentinel tracks compliance across 10 international standards: ISO 27001 (information security), ISO 9001 (quality management), ISO 27035 (incident management), ISO 27017 (cloud security), ISO 20000 (IT service management), ISO 22301 (business continuity), IEC 62443 (industrial/IoT security), NIST CSF 2.0 (cybersecurity framework), SOC 2 Type II (audit assurance), and GDPR (data protection). A total of 78 controls are tracked across these standards.

---

**Q4: How is authentication handled?**

Sentinel uses JWT (JSON Web Token) authentication with HMAC-SHA256 signing. Each token contains the user ID, role, organization ID, and scopes. Three roles are supported: `admin` (full access), `operator` (read/write devices), and `viewer` (read-only). Tokens are validated on every request (except public endpoints like health checks and compliance reports). In development, authentication can be disabled with `AUTH_ENABLED=false`.

---

**Q5: How do I set up the development environment?**

Clone the repository and run:
```bash
make dev
```
This single command tears down any previous state, builds all Docker images, starts DynamoDB Local, auto-creates the database table, and launches all services (Go backend on :8080, Python analytics on :8081, React frontend on :3000, DynamoDB local on :8000). The firmware simulator also starts and automatically registers a test device. No environment configuration is needed. See `docs/LOCAL_DEVELOPMENT.md` for details.

---

**Q6: Can I export my device data?**

Yes. Use the export endpoint at `POST /api/v1/devices/export` with an optional `format` query parameter. Supported formats are `json` (default) and `csv`. The JSON export includes full device metadata with timestamps. The CSV export provides a flat file with the columns device_id, serial_number, device_type, status, and site_id. You can also use the bulk operations endpoints to filter and select specific devices before exporting.

---

**Q7: How does firmware deployment work?**

Firmware deployment is a staged process. You select a firmware version and target devices, then configure a rollout strategy (percentage per stage and delay between stages). The deployment progresses through statuses: PENDING, DOWNLOADING, VERIFYING, APPLYING, and COMPLETED. If verification or application fails, the status becomes FAILED and can be ROLLED_BACK. Each firmware version includes a SHA-256 checksum that is verified before the firmware is applied. Devices can also receive firmware via gRPC streaming, which sends the binary in 64 KB chunks.

---

**Q8: What security protections are in place?**

Sentinel implements defenses against all 10 OWASP Top 10 (2021) categories: input sanitization (injection, XSS, path traversal), JWT authentication with RBAC, brute-force protection (5 failures = 30-minute lockout), rate limiting (100 requests/minute/IP), security headers (HSTS, CSP, X-Frame-Options, X-XSS-Protection), SSRF protection for webhooks, tenant isolation by organization, audit logging of all requests, and request size limits (1 MB max). See `GET /api/v1/security/owasp` for a live status report.

---

**Q9: How do I create an API key for programmatic access?**

Use the Admin Panel's API Keys section or call `POST /api/v1/api-keys` with a JSON body containing `name` (descriptive label), `role` (admin, operator, or viewer), `org_id` (your organization ID), and optionally `expires_in_seconds` (for time-limited keys). The response includes the full key string prefixed with `sk-sentinel-`. This key is shown only once -- store it securely. Use it in the `Authorization: Bearer <key>` header for subsequent API calls.

---

**Q10: What happens to data when a device is decommissioned?**

When you decommission a device (set its status to DECOMMISSIONED), the device record is preserved in DynamoDB for audit purposes. According to the data retention policy, decommissioned device records are retained for 180 days (GDPR legitimate interest basis, Article 6(1)(f)). After the retention period, the record can be permanently deleted. This process satisfies GDPR Article 17 (Right to Erasure) while maintaining the audit trail required by ISO 27001 and SOC 2.
