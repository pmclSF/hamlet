import { test, expect } from '@playwright/test';

test.describe('click actions', () => {
  test('should click elements', async ({ page }) => {
    await page.goto('/app');
    await page.locator('#single').click();
    await page.locator('#double').dblclick();
  });
});
