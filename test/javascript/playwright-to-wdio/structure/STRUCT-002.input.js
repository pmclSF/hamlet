import { test, expect } from '@playwright/test';

test.describe('hooks', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/setup');
  });

  test.afterEach(async ({ page }) => {
    await page.goto('/teardown');
  });

  test('should use hooks', async ({ page }) => {
    await page.locator('#btn').click();
  });
});
