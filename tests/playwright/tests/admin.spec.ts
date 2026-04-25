import { test, expect } from '@playwright/test';

test.describe('Admin Panel', () => {
  const API_BASE = process.env.API_BASE || 'http://localhost:8080';

  test.describe('API Keys', () => {
    test('list API keys endpoint returns data', async ({ request }) => {
      const response = await request.get(`${API_BASE}/api/v1/api-keys`);
      expect(response.ok()).toBeTruthy();

      const body = await response.json();
      expect(body).toHaveProperty('api_keys');
      expect(body).toHaveProperty('total');
      expect(Array.isArray(body.api_keys)).toBeTruthy();
    });

    test('create API key form data is accepted', async ({ request }) => {
      const response = await request.post(`${API_BASE}/api/v1/api-keys`, {
        data: {
          name: 'E2E Test Key',
          role: 'admin',
          org_id: 'org-test',
        },
      });
      expect(response.status()).toBe(201);

      const body = await response.json();
      expect(body).toHaveProperty('key_id');
      expect(body).toHaveProperty('key');
      expect(body.key).toContain('sk-sentinel-');
      expect(body.name).toBe('E2E Test Key');
      expect(body.role).toBe('admin');
      expect(body.active).toBeTruthy();
    });
  });

  test.describe('Webhooks', () => {
    test('list webhooks endpoint returns data', async ({ request }) => {
      const response = await request.get(`${API_BASE}/api/v1/webhooks`);
      expect(response.ok()).toBeTruthy();

      const body = await response.json();
      expect(body).toHaveProperty('webhooks');
      expect(body).toHaveProperty('total');
    });

    test('create webhook with valid data succeeds', async ({ request }) => {
      const response = await request.post(`${API_BASE}/api/v1/webhooks`, {
        data: {
          name: 'E2E Test Webhook',
          url: 'https://example.com/webhook',
          events: ['DEVICE_CREATED', 'ALERT_TRIGGERED'],
          headers: { 'X-Custom': 'test' },
        },
      });
      expect(response.status()).toBe(201);

      const body = await response.json();
      expect(body).toHaveProperty('webhook_id');
      expect(body.name).toBe('E2E Test Webhook');
      expect(body.active).toBeTruthy();
    });

    test('create webhook without URL fails', async ({ request }) => {
      const response = await request.post(`${API_BASE}/api/v1/webhooks`, {
        data: {
          name: 'Bad Webhook',
        },
      });
      expect(response.status()).toBe(400);
    });
  });

  test.describe('Geofences', () => {
    test('list geofences endpoint returns data', async ({ request }) => {
      const response = await request.get(`${API_BASE}/api/v1/geofences`);
      expect(response.ok()).toBeTruthy();

      const body = await response.json();
      expect(body).toHaveProperty('geofences');
      expect(body).toHaveProperty('total');
    });

    test('create geofence with valid data succeeds', async ({ request }) => {
      const response = await request.post(`${API_BASE}/api/v1/geofences`, {
        data: {
          name: 'HQ Perimeter',
          center_lat: 40.7128,
          center_lng: -74.006,
          radius_meters: 500,
          site_id: 'site-hq',
        },
      });
      expect(response.status()).toBe(201);

      const body = await response.json();
      expect(body).toHaveProperty('zone_id');
      expect(body.name).toBe('HQ Perimeter');
    });
  });

  test.describe('Full admin workflow', () => {
    test('create and delete API key lifecycle', async ({ request }) => {
      const createResp = await request.post(`${API_BASE}/api/v1/api-keys`, {
        data: {
          name: 'Lifecycle Test Key',
          role: 'viewer',
          org_id: 'org-lifecycle',
        },
      });
      expect(createResp.status()).toBe(201);
      const created = await createResp.json();

      const listResp = await request.get(`${API_BASE}/api/v1/api-keys`);
      const listBody = await listResp.json();
      expect(listBody.total).toBeGreaterThanOrEqual(1);

      const deleteResp = await request.delete(
        `${API_BASE}/api/v1/api-keys/${created.key_id}`
      );
      expect(deleteResp.status()).toBe(204);
    });
  });
});
