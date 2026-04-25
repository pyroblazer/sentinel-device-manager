import { test, expect } from '@playwright/test';

test.describe('Alerts Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/alerts');
  });

  test('navigates to /alerts and shows Alerts heading', async ({ page }) => {
    await expect(page.locator('.page-header h2')).toHaveText('Alerts');
    await expect(page.locator('.page-header p')).toContainText(
      'Security alerts'
    );
  });

  test('shows alert list table or empty state', async ({ page }) => {
    const tableContainer = page.locator('.table-container');
    await expect(tableContainer).toBeVisible();

    const hasHeaders = (await page.locator('thead th').count()) > 0;
    if (hasHeaders) {
      const expectedHeaders = [
        'Alert ID',
        'Device',
        'Type',
        'Severity',
        'Status',
        'Message',
        'Created',
        'Actions',
      ];
      for (const header of expectedHeaders) {
        await expect(
          page.locator('thead th', { hasText: header })
        ).toBeVisible();
      }
    }

    const emptyState = page.locator('.empty-state');
    if (await emptyState.isVisible()) {
      await expect(emptyState).toContainText('No alerts found');
    }
  });

  test('filter controls are present', async ({ page }) => {
    const filterBar = page.locator('.filter-bar');
    await expect(filterBar).toBeVisible();

    const statusFilter = filterBar.locator('select');
    await expect(statusFilter).toBeVisible();

    const options = statusFilter.locator('option');
    await expect(options).toHaveCount(4);
  });

  test('status filter dropdown has expected values', async ({ page }) => {
    const statusFilter = page.locator('.filter-bar select').first();
    await expect(statusFilter.locator('option')).toHaveCount(4);
    const optionTexts = await statusFilter
      .locator('option')
      .allTextContents();
    expect(optionTexts).toContain('All Statuses');
    expect(optionTexts).toContain('Active');
    expect(optionTexts).toContain('Acknowledged');
    expect(optionTexts).toContain('Resolved');
  });

  test('active alerts show Acknowledge button', async ({ page }) => {
    const activeRows = page.locator('tbody tr');
    const rowCount = await activeRows.count();

    for (let i = 0; i < Math.min(rowCount, 5); i++) {
      const statusBadge = activeRows.nth(i).locator('.badge-active');
      if ((await statusBadge.count()) > 0) {
        const ackButton = activeRows
          .nth(i)
          .locator('button', { hasText: 'Acknowledge' });
        await expect(ackButton).toBeVisible();
      }
    }
  });
});
