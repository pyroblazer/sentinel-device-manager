"""Tests for the demo seed module."""

from datetime import datetime, timezone, timedelta

import pytest

from analytics_service.seed import seed_demo_data, DEVICE_IDS
from analytics_service.store import EventStore
from analytics_service.models import AlertStatus, Severity


@pytest.fixture
def store():
    return EventStore()


class TestSeedDemoData:
    def test_seeds_events(self, store):
        seed_demo_data(store)
        events = store.list_events()
        assert len(events) == 20

    def test_seeds_alerts(self, store):
        seed_demo_data(store)
        alerts = store.list_alerts()
        assert len(alerts) == 8

    def test_events_reference_valid_device_ids(self, store):
        seed_demo_data(store)
        events = store.list_events()
        for event in events:
            assert event.device_id in DEVICE_IDS

    def test_alerts_have_mixed_statuses(self, store):
        seed_demo_data(store)
        alerts = store.list_alerts()
        statuses = {a.status for a in alerts}
        assert AlertStatus.ACTIVE in statuses
        assert AlertStatus.ACKNOWLEDGED in statuses
        assert AlertStatus.RESOLVED in statuses

    def test_alerts_include_critical_severity(self, store):
        seed_demo_data(store)
        alerts = store.list_alerts()
        severities = {a.severity for a in alerts}
        assert Severity.CRITICAL in severities
        assert Severity.WARNING in severities

    def test_events_within_24_hours(self, store):
        now = datetime.now(timezone.utc)
        seed_demo_data(store)
        events = store.list_events()
        for event in events:
            assert event.timestamp >= now - timedelta(hours=25)
            assert event.timestamp <= now + timedelta(minutes=1)

    def test_analytics_summary_nonzero_after_seeding(self, store):
        seed_demo_data(store)
        summary = store.get_analytics_summary()
        assert summary["events_last_24h"] > 0
        assert summary["active_alerts"] > 0
        assert summary["total_devices"] > 0

    def test_active_alerts_count(self, store):
        seed_demo_data(store)
        active = store.list_alerts(status=AlertStatus.ACTIVE)
        assert len(active) == 4

    def test_acknowledged_alerts_have_user(self, store):
        seed_demo_data(store)
        ack_alerts = store.list_alerts(status=AlertStatus.ACKNOWLEDGED)
        for alert in ack_alerts:
            assert alert.acknowledged_by == "admin@sentinel.io"
