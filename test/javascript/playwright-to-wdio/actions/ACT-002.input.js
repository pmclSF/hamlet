import { test, expect } from '@playwright/test';

test.describe('click actions', () => {
  test('should interact', async ({ page }) => {
    await page.goto('/app');
    await page.locator('#double').dblclick();
    await page.locator('#target').hover();
  });
});
