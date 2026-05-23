#!/bin/sh

echo "Seeding demo devices into sentinel-devices table..."

ENDPOINT="--endpoint-url ${DYNAMODB_ENDPOINT:-http://dynamodb:8000}"
REGION="--region us-east-1"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000001"},
    "serial_number": {"S": "VKD-CAM-001"},
    "device_type": {"S": "CAMERA"},
    "model": {"S": "D30"},
    "firmware_version": {"S": "2.4.1"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.101"},
    "mac_address": {"S": "AA:BB:CC:00:01:01"},
    "last_heartbeat": {"S": "2026-05-10T08:30:00Z"},
    "config": {"M": {"resolution": {"S": "4K"}, "retention_days": {"S": "30"}, "motion_sensitivity": {"S": "high"}}},
    "created_at": {"S": "2026-03-01T10:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:30:00Z"}
  }' || echo "  (VKD-CAM-001 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000002"},
    "serial_number": {"S": "VKD-CAM-002"},
    "device_type": {"S": "CAMERA"},
    "model": {"S": "D30"},
    "firmware_version": {"S": "2.4.1"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.102"},
    "mac_address": {"S": "AA:BB:CC:00:01:02"},
    "last_heartbeat": {"S": "2026-05-10T08:28:00Z"},
    "config": {"M": {"resolution": {"S": "4K"}, "retention_days": {"S": "30"}, "night_vision": {"S": "enabled"}}},
    "created_at": {"S": "2026-03-01T10:05:00Z"},
    "updated_at": {"S": "2026-05-10T08:28:00Z"}
  }' || echo "  (VKD-CAM-002 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000003"},
    "serial_number": {"S": "VKD-CAM-003"},
    "device_type": {"S": "CAMERA"},
    "model": {"S": "D30X"},
    "firmware_version": {"S": "2.3.0"},
    "status": {"S": "OFFLINE"},
    "site_id": {"S": "site-beta-002"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.2.101"},
    "mac_address": {"S": "AA:BB:CC:00:02:01"},
    "last_heartbeat": {"S": "2026-05-09T14:00:00Z"},
    "config": {"M": {"resolution": {"S": "1080p"}, "retention_days": {"S": "14"}}},
    "created_at": {"S": "2026-03-10T09:00:00Z"},
    "updated_at": {"S": "2026-05-09T14:00:00Z"}
  }' || echo "  (VKD-CAM-003 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000004"},
    "serial_number": {"S": "VKD-CAM-004"},
    "device_type": {"S": "CAMERA"},
    "model": {"S": "D30"},
    "firmware_version": {"S": "2.4.0"},
    "status": {"S": "MAINTENANCE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.103"},
    "mac_address": {"S": "AA:BB:CC:00:01:03"},
    "last_heartbeat": {"S": "2026-05-10T06:00:00Z"},
    "config": {"M": {"resolution": {"S": "4K"}, "maintenance_reason": {"S": "lens cleaning"}}},
    "created_at": {"S": "2026-03-15T14:00:00Z"},
    "updated_at": {"S": "2026-05-10T06:00:00Z"}
  }' || echo "  (VKD-CAM-004 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000005"},
    "serial_number": {"S": "VKD-AC-001"},
    "device_type": {"S": "ACCESS_CONTROL"},
    "model": {"S": "AC200"},
    "firmware_version": {"S": "1.8.3"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.201"},
    "mac_address": {"S": "AA:BB:CC:00:01:A1"},
    "last_heartbeat": {"S": "2026-05-10T08:32:00Z"},
    "config": {"M": {"auth_mode": {"S": "card_and_pin"}, "schedule": {"S": "24x7"}}},
    "created_at": {"S": "2026-02-15T08:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:32:00Z"}
  }' || echo "  (VKD-AC-001 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000006"},
    "serial_number": {"S": "VKD-AC-002"},
    "device_type": {"S": "ACCESS_CONTROL"},
    "model": {"S": "AC200"},
    "firmware_version": {"S": "1.8.3"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-beta-002"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.2.201"},
    "mac_address": {"S": "AA:BB:CC:00:02:A1"},
    "last_heartbeat": {"S": "2026-05-10T08:31:00Z"},
    "config": {"M": {"auth_mode": {"S": "biometric"}, "schedule": {"S": "business_hours"}}},
    "created_at": {"S": "2026-02-20T11:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:31:00Z"}
  }' || echo "  (VKD-AC-002 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000007"},
    "serial_number": {"S": "VKD-AC-003"},
    "device_type": {"S": "ACCESS_CONTROL"},
    "model": {"S": "AC100"},
    "firmware_version": {"S": "1.7.0"},
    "status": {"S": "OFFLINE"},
    "site_id": {"S": "site-gamma-003"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.3.201"},
    "mac_address": {"S": "AA:BB:CC:00:03:A1"},
    "last_heartbeat": {"S": "2026-05-08T22:00:00Z"},
    "config": {"M": {"auth_mode": {"S": "card_only"}}},
    "created_at": {"S": "2026-04-01T09:00:00Z"},
    "updated_at": {"S": "2026-05-08T22:00:00Z"}
  }' || echo "  (VKD-AC-003 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000008"},
    "serial_number": {"S": "VKD-ALM-001"},
    "device_type": {"S": "ALARM"},
    "model": {"S": "S500"},
    "firmware_version": {"S": "3.1.0"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.301"},
    "mac_address": {"S": "AA:BB:CC:00:01:B1"},
    "last_heartbeat": {"S": "2026-05-10T08:29:00Z"},
    "config": {"M": {"alarm_type": {"S": "intrusion"}, "zone": {"S": "perimeter"}}},
    "created_at": {"S": "2026-02-10T07:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:29:00Z"}
  }' || echo "  (VKD-ALM-001 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000009"},
    "serial_number": {"S": "VKD-ALM-002"},
    "device_type": {"S": "ALARM"},
    "model": {"S": "S500"},
    "firmware_version": {"S": "3.1.0"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-gamma-003"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.3.301"},
    "mac_address": {"S": "AA:BB:CC:00:03:B1"},
    "last_heartbeat": {"S": "2026-05-10T08:25:00Z"},
    "config": {"M": {"alarm_type": {"S": "fire"}, "zone": {"S": "building_b"}}},
    "created_at": {"S": "2026-03-05T13:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:25:00Z"}
  }' || echo "  (VKD-ALM-002 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000010"},
    "serial_number": {"S": "VKD-SEN-001"},
    "device_type": {"S": "SENSOR"},
    "model": {"S": "S10"},
    "firmware_version": {"S": "2.0.5"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-beta-002"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.2.401"},
    "mac_address": {"S": "AA:BB:CC:00:02:C1"},
    "last_heartbeat": {"S": "2026-05-10T08:33:00Z"},
    "config": {"M": {"sensor_type": {"S": "temperature_humidity"}, "poll_interval_sec": {"S": "60"}}},
    "created_at": {"S": "2026-03-20T10:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:33:00Z"}
  }' || echo "  (VKD-SEN-001 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000011"},
    "serial_number": {"S": "VKD-SEN-002"},
    "device_type": {"S": "SENSOR"},
    "model": {"S": "S10"},
    "firmware_version": {"S": "2.0.5"},
    "status": {"S": "ONLINE"},
    "site_id": {"S": "site-alpha-001"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.1.401"},
    "mac_address": {"S": "AA:BB:CC:00:01:C1"},
    "last_heartbeat": {"S": "2026-05-10T08:30:00Z"},
    "config": {"M": {"sensor_type": {"S": "motion"}, "sensitivity": {"S": "medium"}}},
    "created_at": {"S": "2026-04-05T15:00:00Z"},
    "updated_at": {"S": "2026-05-10T08:30:00Z"}
  }' || echo "  (VKD-SEN-002 already exists, skipping)"

aws dynamodb put-item \
  --table-name sentinel-devices \
  $ENDPOINT $REGION \
  --item '{
    "device_id": {"S": "a1b2c3d4-0001-4000-8000-000000000012"},
    "serial_number": {"S": "VKD-SEN-003"},
    "device_type": {"S": "SENSOR"},
    "model": {"S": "S10"},
    "firmware_version": {"S": "1.9.0"},
    "status": {"S": "DECOMMISSIONED"},
    "site_id": {"S": "site-gamma-003"},
    "organization_id": {"S": "org-demo-001"},
    "ip_address": {"S": "10.0.3.401"},
    "mac_address": {"S": "AA:BB:CC:00:03:C1"},
    "last_heartbeat": {"S": "2026-04-28T16:00:00Z"},
    "config": {"M": {"sensor_type": {"S": "door"}, "decommission_reason": {"S": "end of life"}}},
    "created_at": {"S": "2026-01-10T08:00:00Z"},
    "updated_at": {"S": "2026-04-28T16:00:00Z"}
  }' || echo "  (VKD-SEN-003 already exists, skipping)"

echo "Demo device seeding complete. 12 devices seeded."
