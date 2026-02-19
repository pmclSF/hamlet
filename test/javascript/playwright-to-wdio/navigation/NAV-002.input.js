import { test, expect } from '@playwright/test';

test.describe('navigation', () => {
  test('should navigate around', async ({ page }) => {
    await page.goto('/page1');
    await page.reload();
    await page.goBack();
    await page.goForward();
  });
});
