"""FastAPI route handlers for the analytics service."""

from typing import Optional

from fastapi import APIRouter, HTTPException

from .models import (
    AlertAcknowledge,
    AlertStatus,
    EventIngest,
    Severity,
)
from .store import EventStore


def create_router(store: EventStore) -> APIRouter:
    router = APIRouter(prefix="/api/v1")

    @router.get("/events", response_model=list[dict])
    def list_events(
        device_id: Optional[str] = None,
        severity: Optional[Severity] = None,
        limit: int = 50,
    ):
        events = store.list_events(device_id=device_id, severity=severity, limit=limit)
        return [e.model_dump(mode="json") for e in events]

    @router.post("/events", status_code=201)
    def ingest_event(body: EventIngest):
        event = store.ingest_event(body)
        # Auto-generate alert for critical events
        if event.severity == Severity.CRITICAL:
            store.create_alert(
                event,
                alert_type=event.event_type.value,
                message=f"Critical event from device {event.device_id}: {event.event_type.value}",
            )
        return event.model_dump(mode="json")

    @router.get("/alerts", response_model=list[dict])
    def list_alerts(
        status: Optional[AlertStatus] = None,
        severity: Optional[Severity] = None,
        limit: int = 50,
    ):
        alerts = store.list_alerts(status=status, severity=severity, limit=limit)
        return [a.model_dump(mode="json") for a in alerts]

    @router.post("/alerts/{alert_id}/acknowledge")
    def acknowledge_alert(alert_id: str, body: AlertAcknowledge):
        alert = store.acknowledge_alert(alert_id, body.acknowledged_by)
        if alert is None:
            raise HTTPException(status_code=404, detail="Alert not found")
        return alert.model_dump(mode="json")

    @router.get("/analytics/summary")
    def get_analytics_summary():
        return store.get_analytics_summary()

    return router
