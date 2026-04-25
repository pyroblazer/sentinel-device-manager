"""In-memory store for events and alerts.

In production this would use DynamoDB tables. This module provides
a simple dict-based store for development and testing.
"""

import threading
from datetime import datetime, timezone, timedelta
from typing import Optional
from uuid import uuid4

from .models import Alert, AlertStatus, Event, EventIngest, Severity


class EventStore:
    def __init__(self) -> None:
        self._lock = threading.Lock()
        self._events: dict[str, Event] = {}
        self._alerts: dict[str, Alert] = {}

    def ingest_event(self, data: EventIngest) -> Event:
        event = Event(
            event_id=str(uuid4()),
            device_id=data.device_id,
            event_type=data.event_type,
            severity=data.severity,
            payload=data.payload,
            timestamp=datetime.now(timezone.utc),
        )
        with self._lock:
            self._events[event.event_id] = event
        return event

    def list_events(
        self,
        device_id: Optional[str] = None,
        severity: Optional[Severity] = None,
        limit: int = 50,
    ) -> list[Event]:
        with self._lock:
            events = list(self._events.values())
        if device_id:
            events = [e for e in events if e.device_id == device_id]
        if severity:
            events = [e for e in events if e.severity == severity]
        events.sort(key=lambda e: e.timestamp, reverse=True)
        return events[:limit]

    def create_alert(self, event: Event, alert_type: str, message: str) -> Alert:
        alert = Alert(
            alert_id=str(uuid4()),
            device_id=event.device_id,
            event_id=event.event_id,
            alert_type=alert_type,
            severity=event.severity,
            status=AlertStatus.ACTIVE,
            message=message,
            created_at=datetime.now(timezone.utc),
            updated_at=datetime.now(timezone.utc),
        )
        with self._lock:
            self._alerts[alert.alert_id] = alert
        return alert

    def list_alerts(
        self,
        status: Optional[AlertStatus] = None,
        severity: Optional[Severity] = None,
        limit: int = 50,
    ) -> list[Alert]:
        with self._lock:
            alerts = list(self._alerts.values())
        if status:
            alerts = [a for a in alerts if a.status == status]
        if severity:
            alerts = [a for a in alerts if a.severity == severity]
        alerts.sort(key=lambda a: a.created_at, reverse=True)
        return alerts[:limit]

    def acknowledge_alert(self, alert_id: str, user_id: str) -> Optional[Alert]:
        with self._lock:
            alert = self._alerts.get(alert_id)
            if alert is None:
                return None
            alert.status = AlertStatus.ACKNOWLEDGED
            alert.acknowledged_by = user_id
            alert.updated_at = datetime.now(timezone.utc)
            self._alerts[alert_id] = alert
        return alert

    def get_analytics_summary(self) -> dict:
        now = datetime.now(timezone.utc)
        last_24h = now - timedelta(hours=24)

        with self._lock:
            events = list(self._events.values())
            alerts = list(self._alerts.values())

        events_last_24h = [e for e in events if e.timestamp >= last_24h]
        active_alerts = [a for a in alerts if a.status == AlertStatus.ACTIVE]
        critical_alerts = [a for a in active_alerts if a.severity == Severity.CRITICAL]

        device_ids = {e.device_id for e in events}

        return {
            "total_devices": len(device_ids),
            "online_devices": len(device_ids),
            "offline_devices": 0,
            "active_alerts": len(active_alerts),
            "events_last_24h": len(events_last_24h),
            "critical_alerts": len(critical_alerts),
            "firmware_compliance_pct": 95.0,
        }
