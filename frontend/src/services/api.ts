const API_BASE = process.env.REACT_APP_API_BASE || 'http://localhost:8080';
const ANALYTICS_BASE = process.env.REACT_APP_ANALYTICS_BASE || 'http://localhost:8081';

export interface Device {
  device_id: string;
  serial_number: string;
  device_type: 'CAMERA' | 'ACCESS_CONTROL' | 'ALARM' | 'SENSOR';
  model: string;
  firmware_version: string;
  status: 'ONLINE' | 'OFFLINE' | 'MAINTENANCE' | 'DECOMMISSIONED';
  site_id: string;
  organization_id: string;
  ip_address: string;
  mac_address: string;
  last_heartbeat: string;
  config: Record<string, string>;
  created_at: string;
  updated_at: string;
}

export interface DeviceListResponse {
  devices: Device[];
  total: number;
  page: number;
  limit: number;
}

export interface Alert {
  alert_id: string;
  device_id: string;
  event_id: string;
  alert_type: string;
  severity: 'INFO' | 'WARNING' | 'CRITICAL';
  status: 'ACTIVE' | 'ACKNOWLEDGED' | 'RESOLVED';
  message: string;
  acknowledged_by: string | null;
  created_at: string;
  updated_at: string;
}

export interface AnalyticsSummary {
  total_devices: number;
  online_devices: number;
  offline_devices: number;
  active_alerts: number;
  events_last_24h: number;
  critical_alerts: number;
  firmware_compliance_pct: number;
}

export interface Event {
  event_id: string;
  device_id: string;
  event_type: string;
  severity: 'INFO' | 'WARNING' | 'CRITICAL';
  payload: Record<string, string>;
  timestamp: string;
}

export interface FirmwareVersion {
  version: string;
  device_type: string;
  binary_url: string;
  checksum_sha256: string;
  release_notes: string;
  created_at: string;
}

export interface DeviceHealth {
  device_id: string;
  cpu_usage: number;
  memory_usage: number;
  temperature_c: number;
  uptime_seconds: number;
  network_latency_ms: number;
  last_reported: string;
}

async function request<T>(base: string, path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${base}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

// Device API
export const deviceApi = {
  list: (params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return request<DeviceListResponse>(API_BASE, `/api/v1/devices${qs}`);
  },
  get: (id: string) => request<Device>(API_BASE, `/api/v1/devices/${id}`),
  create: (data: Partial<Device>) =>
    request<Device>(API_BASE, '/api/v1/devices', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: string, data: Partial<Device>) =>
    request<Device>(API_BASE, `/api/v1/devices/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: string) =>
    fetch(`${API_BASE}/api/v1/devices/${id}`, { method: 'DELETE' }).then(r => { if (!r.ok && r.status !== 204) throw new Error(`HTTP ${r.status}`); }),
  health: (id: string) => request<DeviceHealth>(API_BASE, `/api/v1/devices/${id}/health`),
};

// Analytics API
export const analyticsApi = {
  summary: () => request<AnalyticsSummary>(ANALYTICS_BASE, '/api/v1/analytics/summary'),
  events: (params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return request<Event[]>(ANALYTICS_BASE, `/api/v1/events${qs}`);
  },
  alerts: (params?: Record<string, string>) => {
    const qs = params ? '?' + new URLSearchParams(params).toString() : '';
    return request<Alert[]>(ANALYTICS_BASE, `/api/v1/alerts${qs}`);
  },
  acknowledgeAlert: (id: string, userId: string) =>
    request<Alert>(ANALYTICS_BASE, `/api/v1/alerts/${id}/acknowledge`, {
      method: 'POST',
      body: JSON.stringify({ acknowledged_by: userId }),
    }),
};

// Superapp API
export const superappApi = {
  // Groups
  listGroups: () => request<any>(API_BASE, '/api/v1/groups'),
  createGroup: (data: any) => request<any>(API_BASE, '/api/v1/groups', { method: 'POST', body: JSON.stringify(data) }),
  deleteGroup: (id: string) => fetch(`${API_BASE}/api/v1/groups/${id}`, { method: 'DELETE' }),
  addDevicesToGroup: (id: string, deviceIds: string[]) => request<any>(API_BASE, `/api/v1/groups/${id}/devices`, { method: 'POST', body: JSON.stringify({ device_ids: deviceIds }) }),

  // Templates
  listTemplates: () => request<any>(API_BASE, '/api/v1/templates'),
  createTemplate: (data: any) => request<any>(API_BASE, '/api/v1/templates', { method: 'POST', body: JSON.stringify(data) }),
  deleteTemplate: (id: string) => fetch(`${API_BASE}/api/v1/templates/${id}`, { method: 'DELETE' }),

  // Bulk
  bulkDelete: (deviceIds: string[]) => request<any>(API_BASE, '/api/v1/devices/bulk-delete', { method: 'POST', body: JSON.stringify({ device_ids: deviceIds }) }),
  bulkUpdate: (deviceIds: string[], updates: Record<string, string>) => request<any>(API_BASE, '/api/v1/devices/bulk-update', { method: 'POST', body: JSON.stringify({ device_ids: deviceIds, updates }) }),
  bulkTag: (deviceIds: string[], tags: string[]) => request<any>(API_BASE, '/api/v1/devices/bulk-tag', { method: 'POST', body: JSON.stringify({ device_ids: deviceIds, tags }) }),
  exportDevices: (format: string) => `${API_BASE}/api/v1/devices/export?format=${format}`,

  // Webhooks
  listWebhooks: () => request<any>(API_BASE, '/api/v1/webhooks'),
  createWebhook: (data: any) => request<any>(API_BASE, '/api/v1/webhooks', { method: 'POST', body: JSON.stringify(data) }),
  deleteWebhook: (id: string) => fetch(`${API_BASE}/api/v1/webhooks/${id}`, { method: 'DELETE' }),
  testWebhook: (id: string) => request<any>(API_BASE, `/api/v1/webhooks/${id}/test`, { method: 'POST' }),

  // API Keys
  listAPIKeys: () => request<any>(API_BASE, '/api/v1/api-keys'),
  createAPIKey: (data: any) => request<any>(API_BASE, '/api/v1/api-keys', { method: 'POST', body: JSON.stringify(data) }),
  deleteAPIKey: (id: string) => fetch(`${API_BASE}/api/v1/api-keys/${id}`, { method: 'DELETE' }),

  // Notifications
  listNotifications: () => request<any>(API_BASE, '/api/v1/notifications'),
  markRead: (id: string) => request<any>(API_BASE, `/api/v1/notifications/${id}/read`, { method: 'POST' }),
  deleteNotification: (id: string) => fetch(`${API_BASE}/api/v1/notifications/${id}`, { method: 'DELETE' }),

  // Geofences
  listGeofences: () => request<any>(API_BASE, '/api/v1/geofences'),
  createGeofence: (data: any) => request<any>(API_BASE, '/api/v1/geofences', { method: 'POST', body: JSON.stringify(data) }),
  deleteGeofence: (id: string) => fetch(`${API_BASE}/api/v1/geofences/${id}`, { method: 'DELETE' }),
};
