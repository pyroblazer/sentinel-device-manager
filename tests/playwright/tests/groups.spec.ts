import { test, expect } from '@playwright/test';

test.describe('Device Groups', () => {
  const API_BASE = process.env.API_BASE || 'http://localhost:8080';

  test('list groups endpoint returns data', async ({ request }) => {
    const response = await request.get(`${API_BASE}/api/v1/groups`);
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body).toHaveProperty('groups');
    expect(body).toHaveProperty('total');
    expect(Array.isArray(body.groups)).toBeTruthy();
  });

  test('create group with valid data succeeds', async ({ request }) => {
    const response = await request.post(`${API_BASE}/api/v1/groups`, {
      data: {
        name: 'Building A - Floor 1',
        type: 'FLOOR',
        site_id: 'site-001',
      },
    });
    expect(response.status()).toBe(201);

    const body = await response.json();
    expect(body).toHaveProperty('group_id');
    expect(body).toHaveProperty('name');
    expect(body.name).toBe('Building A - Floor 1');
    expect(body.type).toBe('FLOOR');
    expect(body.site_id).toBe('site-001');
    expect(body.device_ids).toEqual([]);  });

  test('create group without name fails', async ({ request }) => {
    const response = await request.post(`${API_BASE}/api/v1/groups`, {
      data: {
        type: 'ZONE',
        site_id: 'site-001',
      },
    });
    expect(response.status()).toBe(400);

    const body = await response.json();
    expect(body.error).toContain('name is required');
  });

  test('create group, get, update, and delete lifecycle', async ({
    request,
  }) => {
    // Create
    const createResp = await request.post(`${API_BASE}/api/v1/groups`, {
      data: {
        name: 'Integration Test Group',
        type: 'CUSTOM',
        site_id: 'site-test',
      },
    });
    expect(createResp.status()).toBe(201);
    const created = await createResp.json();
    const groupId = created.group_id;

    // Get
    const getResp = await request.get(`${API_BASE}/api/v1/groups/${groupId}`);
    expect(getResp.ok()).toBeTruthy();
    const fetched = await getResp.json();
    expect(fetched.name).toBe('Integration Test Group');

    // Update
    const updateResp = await request.put(
      `${API_BASE}/api/v1/groups/${groupId}`,
      {
        data: { name: 'Updated Group Name' },
      }
    );
    expect(updateResp.ok()).toBeTruthy();
    const updated = await updateResp.json();
    expect(updated.name).toBe('Updated Group Name');

    // Delete
    const deleteResp = await request.delete(
      `${API_BASE}/api/v1/groups/${groupId}`
    );
    expect(deleteResp.status()).toBe(204);
  });

  test('add devices to a group', async ({ request }) => {
    const createResp = await request.post(`${API_BASE}/api/v1/groups`, {
      data: {
        name: 'Device Test Group',
        type: 'ZONE',
        site_id: 'site-001',
      },
    });
    const group = await createResp.json();

    const addResp = await request.post(
      `${API_BASE}/api/v1/groups/${group.group_id}/devices`,
      {
        data: {
          device_ids: ['device-001', 'device-002'],
        },
      }
    );
    expect(addResp.ok()).toBeTruthy();
    const updated = await addResp.json();
    expect(updated.device_ids).toContain('device-001');
    expect(updated.device_ids).toContain('device-002');
  });
});
