import { test, expect } from '@playwright/test';

test.describe('click actions', () => {
  test('should interact', async ({ page }) => {
    await page.goto('/app');
    await page.locator('#btn').click();
    await page.locator('#dbl').dblclick();
    await page.locator('#menu').hover();
  });
});
