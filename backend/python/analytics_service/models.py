from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime
from enum import Enum


class Severity(str, Enum):
    INFO = "INFO"
    WARNING = "WARNING"
    CRITICAL = "CRITICAL"


class AlertStatus(str, Enum):
    ACTIVE = "ACTIVE"
    ACKNOWLEDGED = "ACKNOWLEDGED"
    RESOLVED = "RESOLVED"


class EventType(str, Enum):
    MOTION_DETECTED = "MOTION_DETECTED"
    DOOR_OPENED = "DOOR_OPENED"
    DOOR_FORCED = "DOOR_FORCED"
    ALARM_TRIGGERED = "ALARM_TRIGGERED"
    TEMPERATURE_THRESHOLD = "TEMPERATURE_THRESHOLD"
    DEVICE_OFFLINE = "DEVICE_OFFLINE"
    DEVICE_ONLINE = "DEVICE_ONLINE"
    FIRMWARE_UPDATED = "FIRMWARE_UPDATED"
    TAMPER_DETECTED = "TAMPER_DETECTED"


class Event(BaseModel):
    event_id: str
    device_id: str
    event_type: EventType
    severity: Severity
    payload: dict = Field(default_factory=dict)
    timestamp: datetime


class EventIngest(BaseModel):
    device_id: str
    event_type: EventType
    severity: Severity
    payload: dict = Field(default_factory=dict)


class Alert(BaseModel):
    alert_id: str
    device_id: str
    event_id: str
    alert_type: str
    severity: Severity
    status: AlertStatus
    message: str
    acknowledged_by: Optional[str] = None
    created_at: datetime
    updated_at: datetime


class AlertAcknowledge(BaseModel):
    acknowledged_by: str


class AnalyticsSummary(BaseModel):
    total_devices: int
    online_devices: int
    offline_devices: int
    active_alerts: int
    events_last_24h: int
    critical_alerts: int
    firmware_compliance_pct: float
