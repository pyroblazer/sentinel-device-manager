import { test, expect } from '@playwright/test';

test.describe('Notification Center', () => {
  test.describe('via API (no dedicated UI page)', () => {
    const API_BASE = process.env.API_BASE || 'http://localhost:8080';

    test('notifications endpoint returns list', async ({ request }) => {
      const response = await request.get(`${API_BASE}/api/v1/notifications`);
      expect(response.ok()).toBeTruthy();

      const body = await response.json();
      expect(body).toHaveProperty('notifications');
      expect(body).toHaveProperty('total');
      expect(Array.isArray(body.notifications)).toBeTruthy();
    });

    test('creating a group generates a notification', async ({ request }) => {
      const createResp = await request.post(`${API_BASE}/api/v1/groups`, {
        data: {
          name: 'Notification Test Group',
          type: 'ZONE',
          site_id: 'site-test',
        },
      });
      expect(createResp.ok()).toBeTruthy();

      const notifResp = await request.get(`${API_BASE}/api/v1/notifications`);
      expect(notifResp.ok()).toBeTruthy();

      const body = await notifResp.json();
      expect(body.total).toBeGreaterThanOrEqual(1);

      const hasGroupNotification = body.notifications.some(
        (n: any) =>
          n.title === 'Group Created' &&
          n.message.includes('Notification Test Group')
      );
      expect(hasGroupNotification).toBeTruthy();
    });

    test('mark read endpoint works', async ({ request }) => {
      const listResp = await request.get(`${API_BASE}/api/v1/notifications`);
      const body = await listResp.json();

      if (body.notifications.length > 0) {
        const notificationId = body.notifications[0].notification_id;
        const readResp = await request.post(
          `${API_BASE}/api/v1/notifications/${notificationId}/read`
        );
        expect(readResp.ok()).toBeTruthy();

        const updated = await readResp.json();
        expect(updated.read).toBeTruthy();
      }
    });
  });
});
