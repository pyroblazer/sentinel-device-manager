"""Seed demo events and alerts into the EventStore for local demo mode."""

from datetime import datetime, timezone, timedelta

from .models import AlertStatus, EventIngest
from .store import EventStore

# Fixed device IDs matching scripts/seed-dynamodb.sh
DEVICE_IDS = [
    "a1b2c3d4-0001-4000-8000-000000000001",
    "a1b2c3d4-0001-4000-8000-000000000002",
    "a1b2c3d4-0001-4000-8000-000000000003",
    "a1b2c3d4-0001-4000-8000-000000000004",
    "a1b2c3d4-0001-4000-8000-000000000005",
    "a1b2c3d4-0001-4000-8000-000000000006",
    "a1b2c3d4-0001-4000-8000-000000000007",
    "a1b2c3d4-0001-4000-8000-000000000008",
    "a1b2c3d4-0001-4000-8000-000000000009",
    "a1b2c3d4-0001-4000-8000-000000000010",
    "a1b2c3d4-0001-4000-8000-000000000011",
    "a1b2c3d4-0001-4000-8000-000000000012",
]


def seed_demo_data(store: EventStore) -> None:
    now = datetime.now(timezone.utc)

    events_spec = [
        # (hours_ago, device_index, event_type, severity, payload)
        (1, 0, "MOTION_DETECTED", "INFO", {"zone": "lobby", "confidence": "0.92"}),
        (2, 4, "DOOR_FORCED", "CRITICAL", {"door_id": "north-entrance", "badge": "none"}),
        (3, 7, "TAMPER_DETECTED", "CRITICAL", {"sensor": "enclosure", "location": "site-gamma"}),
        (4, 2, "DEVICE_OFFLINE", "WARNING", {"last_seen": "30 min ago"}),
        (5, 9, "ALARM_TRIGGERED", "CRITICAL", {"zone": "perimeter", "alarm_type": "intrusion"}),
        (6, 1, "DOOR_OPENED", "INFO", {"door_id": "east-stairwell", "badge": "emp-042"}),
        (7, 10, "TEMPERATURE_THRESHOLD", "WARNING", {"temperature_c": "38.5", "threshold": "35.0"}),
        (8, 3, "DEVICE_ONLINE", "INFO", {"downtime_minutes": "180"}),
        (9, 5, "MOTION_DETECTED", "INFO", {"zone": "server-room", "confidence": "0.78"}),
        (10, 8, "DOOR_FORCED", "CRITICAL", {"door_id": "loading-dock", "badge": "unknown"}),
        (11, 6, "DEVICE_OFFLINE", "WARNING", {"last_seen": "2 hours ago"}),
        (12, 0, "MOTION_DETECTED", "INFO", {"zone": "parking-garage", "confidence": "0.85"}),
        (13, 11, "TEMPERATURE_THRESHOLD", "INFO", {"temperature_c": "28.2", "threshold": "30.0"}),
        (14, 1, "MOTION_DETECTED", "INFO", {"zone": "lobby", "confidence": "0.95"}),
        (15, 9, "ALARM_TRIGGERED", "WARNING", {"zone": "building_b", "alarm_type": "fire"}),
        (16, 3, "TAMPER_DETECTED", "CRITICAL", {"sensor": "lens", "location": "site-beta-002"}),
        (17, 7, "DOOR_OPENED", "INFO", {"door_id": "south-exit", "badge": "emp-117"}),
        (18, 2, "DEVICE_ONLINE", "INFO", {"downtime_minutes": "45"}),
        (19, 4, "MOTION_DETECTED", "INFO", {"zone": "reception", "confidence": "0.88"}),
        (20, 10, "TEMPERATURE_THRESHOLD", "WARNING", {"temperature_c": "36.1", "threshold": "35.0"}),
    ]

    created_events = []
    for hours_ago, dev_idx, event_type, severity, payload in events_spec:
        event = store.ingest_event(EventIngest(
            device_id=DEVICE_IDS[dev_idx],
            event_type=event_type,
            severity=severity,
            payload=payload,
        ))
        # Overwrite timestamp to backdate
        event.timestamp = now - timedelta(hours=hours_ago)
        created_events.append((event, event_type, severity))

    # Create alerts from critical and warning events
    alerts_spec = [
        # (event_index, alert_type, message, initial_status)
        (1, "DOOR_FORCED", "Forced entry detected at north entrance - no badge presented", AlertStatus.ACTIVE),
        (2, "TAMPER_DETECTED", "Tamper alert on sensor enclosure at site-gamma", AlertStatus.ACTIVE),
        (4, "DEVICE_OFFLINE",
         "Camera VKD-CAM-003 offline for 30 min at site-beta",
         AlertStatus.ACTIVE),
        (6, "TEMPERATURE_THRESHOLD",
         "Temp 38.5C exceeds threshold at site-beta server room",
         AlertStatus.ACTIVE),
        (9, "DOOR_FORCED",
         "Forced entry at loading dock - resolved by security",
         AlertStatus.RESOLVED),
        (14, "ALARM_TRIGGERED",
         "Fire alarm at site-alpha perimeter - under investigation",
         AlertStatus.ACKNOWLEDGED),
        (15, "TEMPERATURE_THRESHOLD", "Temperature warning at site-beta acknowledged by ops", AlertStatus.ACKNOWLEDGED),
        (16, "TAMPER_DETECTED", "Lens tamper detected on camera at site-beta", AlertStatus.ACKNOWLEDGED),
    ]

    for event_idx, alert_type, message, initial_status in alerts_spec:
        event = created_events[event_idx][0]
        alert = store.create_alert(event, alert_type, message)
        if initial_status == AlertStatus.ACKNOWLEDGED:
            store.acknowledge_alert(alert.alert_id, "admin@sentinel.io")
        elif initial_status == AlertStatus.RESOLVED:
            store.acknowledge_alert(alert.alert_id, "admin@sentinel.io")
            alert.status = AlertStatus.RESOLVED
