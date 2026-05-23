import { test as setup, expect } from '@playwright/test';

const authFile = 'playwright/.auth/user.json';

setup('authenticate', async ({ page }) => {
  await page.goto('/');
  // The app shows login page when unauthenticated
  await page.waitForSelector('button[type="submit"]');

  // Credentials are pre-filled (admin@sentinel.io / admin), just submit
  await page.click('button[type="submit"]');

  // Wait for redirect to dashboard after login
  await page.waitForURL('**/', { timeout: 10000 });
  await expect(page.locator('.sidebar')).toBeVisible({ timeout: 5000 });

  await page.context().storageState({ path: authFile });
});
