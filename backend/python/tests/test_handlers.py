"""Tests for the analytics service."""

import pytest
from fastapi.testclient import TestClient

from analytics_service.main import app, store
from analytics_service.models import EventIngest, Severity, EventType


@pytest.fixture
def client():
    return TestClient(app)


@pytest.fixture(autouse=True)
def clear_store():
    store._events.clear()
    store._alerts.clear()


class TestHealthCheck:
    def test_health_returns_ok(self, client):
        resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json() == {"status": "ok"}


class TestEventIngestion:
    def test_ingest_event(self, client):
        resp = client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "MOTION_DETECTED",
            "severity": "INFO",
            "payload": {"zone": "entrance"},
        })
        assert resp.status_code == 201
        data = resp.json()
        assert data["device_id"] == "dev-001"
        assert data["event_type"] == "MOTION_DETECTED"
        assert data["severity"] == "INFO"
        assert "event_id" in data
        assert "timestamp" in data

    def test_ingest_critical_creates_alert(self, client):
        resp = client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "ALARM_TRIGGERED",
            "severity": "CRITICAL",
            "payload": {"zone": "server_room"},
        })
        assert resp.status_code == 201

        alerts = client.get("/api/v1/alerts").json()
        assert len(alerts) == 1
        assert alerts[0]["severity"] == "CRITICAL"
        assert alerts[0]["device_id"] == "dev-001"
        assert alerts[0]["status"] == "ACTIVE"

    def test_ingest_info_does_not_create_alert(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "DEVICE_ONLINE",
            "severity": "INFO",
        })
        alerts = client.get("/api/v1/alerts").json()
        assert len(alerts) == 0

    def test_list_events(self, client):
        for i in range(3):
            client.post("/api/v1/events", json={
                "device_id": f"dev-{i}",
                "event_type": "MOTION_DETECTED",
                "severity": "INFO",
            })
        resp = client.get("/api/v1/events")
        assert resp.status_code == 200
        assert len(resp.json()) == 3

    def test_list_events_filter_by_device(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "MOTION_DETECTED",
            "severity": "INFO",
        })
        client.post("/api/v1/events", json={
            "device_id": "dev-002",
            "event_type": "DOOR_OPENED",
            "severity": "WARNING",
        })
        resp = client.get("/api/v1/events?device_id=dev-001")
        data = resp.json()
        assert len(data) == 1
        assert data[0]["device_id"] == "dev-001"

    def test_list_events_filter_by_severity(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "MOTION_DETECTED",
            "severity": "INFO",
        })
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "ALARM_TRIGGERED",
            "severity": "CRITICAL",
        })
        resp = client.get("/api/v1/events?severity=CRITICAL")
        data = resp.json()
        assert len(data) == 1
        assert data[0]["severity"] == "CRITICAL"


class TestAlerts:
    def test_list_alerts_empty(self, client):
        resp = client.get("/api/v1/alerts")
        assert resp.status_code == 200
        assert resp.json() == []

    def test_acknowledge_alert(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "TAMPER_DETECTED",
            "severity": "CRITICAL",
        })
        alerts = client.get("/api/v1/alerts").json()
        alert_id = alerts[0]["alert_id"]

        resp = client.post(f"/api/v1/alerts/{alert_id}/acknowledge", json={
            "acknowledged_by": "admin@sentinel.io",
        })
        assert resp.status_code == 200
        data = resp.json()
        assert data["status"] == "ACKNOWLEDGED"
        assert data["acknowledged_by"] == "admin@sentinel.io"

    def test_acknowledge_nonexistent_alert(self, client):
        resp = client.post("/api/v1/alerts/nonexistent/acknowledge", json={
            "acknowledged_by": "admin@sentinel.io",
        })
        assert resp.status_code == 404

    def test_list_alerts_filter_by_status(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "TAMPER_DETECTED",
            "severity": "CRITICAL",
        })
        active = client.get("/api/v1/alerts?status=ACTIVE").json()
        assert len(active) == 1

        ack = client.get("/api/v1/alerts?status=ACKNOWLEDGED").json()
        assert len(ack) == 0


class TestAnalyticsSummary:
    def test_summary_returns_fields(self, client):
        client.post("/api/v1/events", json={
            "device_id": "dev-001",
            "event_type": "MOTION_DETECTED",
            "severity": "INFO",
        })
        resp = client.get("/api/v1/analytics/summary")
        assert resp.status_code == 200
        data = resp.json()
        assert "total_devices" in data
        assert "online_devices" in data
        assert "active_alerts" in data
        assert "events_last_24h" in data
        assert "critical_alerts" in data
        assert "firmware_compliance_pct" in data

    def test_summary_counts_events(self, client):
        for _ in range(5):
            client.post("/api/v1/events", json={
                "device_id": "dev-001",
                "event_type": "MOTION_DETECTED",
                "severity": "INFO",
            })
        data = client.get("/api/v1/analytics/summary").json()
        assert data["events_last_24h"] == 5
