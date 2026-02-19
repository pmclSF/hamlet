import { test, expect } from '@playwright/test';

test.describe('navigation', () => {
  test('should navigate back and forward', async ({ page }) => {
    await page.goto('/page1');
    await page.goto('/page2');
    await page.goBack();
    await page.goForward();
    await page.reload();
  });
});
