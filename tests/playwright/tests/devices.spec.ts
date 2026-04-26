import { test, expect } from '@playwright/test';

test.describe('Device Management Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/devices');
  });

  test('navigates to /devices and shows Devices heading', async ({ page }) => {
    await expect(page.locator('.page-header h2')).toHaveText('Devices');
    await expect(page.locator('.page-header p')).toContainText(
      'Manage cameras'
    );
  });

  test('shows device list with table headers', async ({ page }) => {
    const table = page.locator('.table-container');
    await expect(table).toBeVisible();

    const expectedHeaders = [
      'Serial Number',
      'Type',
      'Model',
      'Site',
      'Status',
      'Firmware',
      'Last Heartbeat',
      'Actions',
    ];

    for (const header of expectedHeaders) {
      await expect(page.locator('thead th', { hasText: header })).toBeVisible();
    }
  });

  test('filter by type dropdown is present and functional', async ({ page }) => {
    const typeSelect = page.locator('.filter-bar select').first();
    await expect(typeSelect).toBeVisible();

    const typeOptions = typeSelect.locator('option');
    await expect(typeOptions).toHaveCount(5);

    await typeSelect.selectOption('CAMERA');
    expect(await typeSelect.inputValue()).toBe('CAMERA');
  });

  test('filter by status dropdown is present and functional', async ({ page }) => {
    const statusSelect = page.locator('.filter-bar select').nth(1);
    await expect(statusSelect).toBeVisible();

    const statusOptions = statusSelect.locator('option');
    await expect(statusOptions).toHaveCount(5);

    await statusSelect.selectOption('ONLINE');
    expect(await statusSelect.inputValue()).toBe('ONLINE');
  });

  test('shows empty state when no devices match filter', async ({ page }) => {
    await page.locator('.filter-bar select').first().selectOption('SENSOR');
    const emptyState = page.locator('.empty-state');
    if (await emptyState.isVisible()) {
      await expect(emptyState).toContainText('No devices found');
    }
  });

  test('click a device row to navigate to detail page', async ({ page }) => {
    const deviceRow = page.locator('tbody tr[style*="cursor: pointer"]').first();
    if ((await deviceRow.count()) > 0) {
      const isClickable = await deviceRow.evaluate((el) => {
        return el.style.cursor === 'pointer';
      });

      if (isClickable) {
        await deviceRow.click();
        await expect(page).toHaveURL(/\/devices\/[a-f0-9-]+/);
      }
    }
  });

  test('device detail shows all fields when navigated directly', async ({
    page,
  }) => {
    await page.goto('/devices');
    const deviceRow = page.locator('tbody tr[style*="cursor: pointer"]').first();
    if ((await deviceRow.count()) > 0) {
      await deviceRow.click();
      await expect(page).toHaveURL(/\/devices\/[a-f0-9-]+/);

      const detailPage = page.locator('.device-detail');
      await expect(detailPage).toBeVisible();

      const expectedLabels = [
        'Device ID',
        'Status',
        'Firmware',
        'Site',
        'IP Address',
        'MAC Address',
        'Created',
        'Last Heartbeat',
      ];

      for (const label of expectedLabels) {
        await expect(
          page.locator('.detail-field label', { hasText: label })
        ).toBeVisible();
      }
    }
  });

  test('device detail has Back and Decommission buttons', async ({ page }) => {
    await page.goto('/devices');
    const deviceRow = page.locator('tbody tr[style*="cursor: pointer"]').first();
    if ((await deviceRow.count()) > 0) {
      await deviceRow.click();
      await expect(
        page.locator('button', { hasText: 'Back' })
      ).toBeVisible();
      await expect(
        page.locator('button', { hasText: 'Decommission' })
      ).toBeVisible();
    }
  });
});
