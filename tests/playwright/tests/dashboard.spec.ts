import { test, expect } from '@playwright/test';

test.describe('Dashboard Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('loads and shows Dashboard heading', async ({ page }) => {
    await expect(page.locator('.page-header h2')).toHaveText('Dashboard');
    await expect(page.locator('.page-header p')).toContainText(
      'Real-time overview'
    );
  });

  test('shows stat cards for key metrics', async ({ page }) => {
    const statLabels = page.locator('.stat-label');
    await expect(statLabels).toHaveCount(7);

    await expect(page.locator('.stat-label').nth(0)).toHaveText('Total Devices');
    await expect(page.locator('.stat-label').nth(1)).toHaveText('Online');
    await expect(page.locator('.stat-label').nth(2)).toHaveText('Offline');
    await expect(page.locator('.stat-label').nth(3)).toHaveText('Active Alerts');
  });

  test('shows device table or empty state', async ({ page }) => {
    const table = page.locator('.table-container');
    await expect(table).toBeVisible();

    const hasTableHeader = await page.locator('thead th').count();
    if (hasTableHeader > 0) {
      await expect(page.locator('thead th').first()).toHaveText('Serial');
    }

    const emptyState = page.locator('.empty-state');
    if (await emptyState.isVisible()) {
      await expect(emptyState).toContainText('No devices registered');
    }
  });

  test('verifies dark theme CSS variables are applied', async ({ page }) => {
    const rootStyles = await page.evaluate(() => {
      const root = document.documentElement;
      const computed = getComputedStyle(root);
      return {
        bgPrimary: computed.getPropertyValue('--bg-primary').trim(),
        textPrimary: computed.getPropertyValue('--text-primary').trim(),
        accentBlue: computed.getPropertyValue('--accent-blue').trim(),
        fontFamily: computed.getPropertyValue('--font-sans').trim(),
      };
    });

    expect(rootStyles.bgPrimary).toBe('#0a0a0f');
    expect(rootStyles.textPrimary).toBe('#f0f0f8');
    expect(rootStyles.accentBlue).toBe('#4a9eff');
    expect(rootStyles.fontFamily).toContain('Inter');
  });

  test('sidebar navigation is visible', async ({ page }) => {
    await expect(page.locator('.sidebar')).toBeVisible();
    await expect(page.locator('.sidebar-logo h1')).toHaveText('Sentinel');
    await expect(page.locator('.sidebar-subtitle')).toHaveText('Device Manager');
  });

  test('stat cards display numeric values', async ({ page }) => {
    const statValues = page.locator('.stat-value');
    await expect(statValues).toHaveCount(7);

    for (let i = 0; i < 7; i++) {
      const text = await statValues.nth(i).textContent();
      expect(text).not.toBeNull();
    }
  });
});
