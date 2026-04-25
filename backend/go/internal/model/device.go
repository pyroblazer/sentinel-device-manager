// Package model defines the domain entities for the Sentinel Device Manager.
//
// These models represent the core data structures used throughout the application:
// devices, health telemetry, events, alerts, firmware, and firmware deployments.
// All models include both JSON and DynamoDB attribute value tags for serialization.
package model

import "time"

// DeviceType enumerates the physical security device categories managed by the platform.
type DeviceType string

const (
	DeviceTypeCamera        DeviceType = "CAMERA"
	DeviceTypeAccessControl DeviceType = "ACCESS_CONTROL"
	DeviceTypeAlarm         DeviceType = "ALARM"
	DeviceTypeSensor        DeviceType = "SENSOR"
)

// DeviceStatus represents the operational status of a managed device.
type DeviceStatus string

const (
	StatusOnline         DeviceStatus = "ONLINE"
	StatusOffline        DeviceStatus = "OFFLINE"
	StatusMaintenance    DeviceStatus = "MAINTENANCE"
	StatusDecommissioned DeviceStatus = "DECOMMISSIONED"
)

// Device represents a physical security device in the system.
// This is the primary domain entity stored in DynamoDB with device_id as the partition key.
type Device struct {
	DeviceID        string            `json:"device_id" dynamodbav:"device_id"`
	SerialNumber    string            `json:"serial_number" dynamodbav:"serial_number"`
	DeviceType      DeviceType        `json:"device_type" dynamodbav:"device_type"`
	Model           string            `json:"model" dynamodbav:"model"`
	FirmwareVersion string            `json:"firmware_version" dynamodbav:"firmware_version"`
	Status          DeviceStatus      `json:"status" dynamodbav:"status"`
	SiteID          string            `json:"site_id" dynamodbav:"site_id"`
	OrganizationID  string            `json:"organization_id" dynamodbav:"organization_id"`
	IPAddress       string            `json:"ip_address" dynamodbav:"ip_address"`
	MACAddress      string            `json:"mac_address" dynamodbav:"mac_address"`
	LastHeartbeat   time.Time         `json:"last_heartbeat" dynamodbav:"last_heartbeat"`
	Config          map[string]string `json:"config" dynamodbav:"config"`
	CreatedAt       time.Time         `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at" dynamodbav:"updated_at"`
}

// DeviceHealth captures real-time telemetry from a device including CPU, memory,
// temperature, and network metrics. Used for monitoring and alerting.
type DeviceHealth struct {
	DeviceID         string    `json:"device_id" dynamodbav:"device_id"`
	CPUUsage         float64   `json:"cpu_usage" dynamodbav:"cpu_usage"`
	MemoryUsage      float64   `json:"memory_usage" dynamodbav:"memory_usage"`
	TemperatureC     float64   `json:"temperature_c" dynamodbav:"temperature_c"`
	UptimeSeconds    int64     `json:"uptime_seconds" dynamodbav:"uptime_seconds"`
	NetworkLatencyMs float64   `json:"network_latency_ms" dynamodbav:"network_latency_ms"`
	LastReported     time.Time `json:"last_reported" dynamodbav:"last_reported"`
}

// Severity classifies the impact level of events and alerts.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Event represents a telemetry or security event emitted by a device.
// Events are ingested through the Python analytics service and may trigger alerts.
type Event struct {
	EventID   string            `json:"event_id" dynamodbav:"event_id"`
	DeviceID  string            `json:"device_id" dynamodbav:"device_id"`
	EventType string            `json:"event_type" dynamodbav:"event_type"`
	Severity  Severity          `json:"severity" dynamodbav:"severity"`
	Payload   map[string]string `json:"payload" dynamodbav:"payload"`
	Timestamp time.Time         `json:"timestamp" dynamodbav:"timestamp"`
}

// AlertStatus tracks the lifecycle of an alert from creation to resolution.
type AlertStatus string

const (
	AlertStatusActive        AlertStatus = "ACTIVE"
	AlertStatusAcknowledged  AlertStatus = "ACKNOWLEDGED"
	AlertStatusResolved      AlertStatus = "RESOLVED"
)

// Alert represents an actionable notification generated from critical events.
// Alerts follow a lifecycle: ACTIVE -> ACKNOWLEDGED -> RESOLVED.
type Alert struct {
	AlertID         string      `json:"alert_id" dynamodbav:"alert_id"`
	DeviceID        string      `json:"device_id" dynamodbav:"device_id"`
	EventID         string      `json:"event_id" dynamodbav:"event_id"`
	AlertType       string      `json:"alert_type" dynamodbav:"alert_type"`
	Severity        Severity    `json:"severity" dynamodbav:"severity"`
	Status          AlertStatus `json:"status" dynamodbav:"status"`
	Message         string      `json:"message" dynamodbav:"message"`
	AcknowledgedBy  string      `json:"acknowledged_by,omitempty" dynamodbav:"acknowledged_by"`
	CreatedAt       time.Time   `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at" dynamodbav:"updated_at"`
}

// Firmware represents a firmware release artifact with integrity verification.
// SHA-256 checksums are used for firmware integrity validation (ISO 27001 A.14.1).
type Firmware struct {
	Version       string     `json:"version" dynamodbav:"version"`
	DeviceType    DeviceType `json:"device_type" dynamodbav:"device_type"`
	BinaryURL     string     `json:"binary_url" dynamodbav:"binary_url"`
	ChecksumSHA256 string   `json:"checksum_sha256" dynamodbav:"checksum_sha256"`
	ReleaseNotes  string     `json:"release_notes" dynamodbav:"release_notes"`
	CreatedAt     time.Time  `json:"created_at" dynamodbav:"created_at"`
}

// FirmwareDeploymentStatus tracks the lifecycle of a firmware deployment to devices.
type FirmwareDeploymentStatus string

const (
	FirmwareStatusPending    FirmwareDeploymentStatus = "PENDING"
	FirmwareStatusDownloading FirmwareDeploymentStatus = "DOWNLOADING"
	FirmwareStatusVerifying  FirmwareDeploymentStatus = "VERIFYING"
	FirmwareStatusApplying   FirmwareDeploymentStatus = "APPLYING"
	FirmwareStatusCompleted  FirmwareDeploymentStatus = "COMPLETED"
	FirmwareStatusFailed     FirmwareDeploymentStatus = "FAILED"
	FirmwareStatusRolledBack FirmwareDeploymentStatus = "ROLLED_BACK"
)

// FirmwareDeployment tracks a firmware rollout targeting specific devices.
// Supports staged deployments with status progression monitoring.
type FirmwareDeployment struct {
	DeploymentID string                   `json:"deployment_id" dynamodbav:"deployment_id"`
	Version      string                   `json:"version" dynamodbav:"version"`
	DeviceIDs    []string                 `json:"device_ids" dynamodbav:"device_ids"`
	Status       FirmwareDeploymentStatus `json:"status" dynamodbav:"status"`
	CreatedAt    time.Time                `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time                `json:"updated_at" dynamodbav:"updated_at"`
}
