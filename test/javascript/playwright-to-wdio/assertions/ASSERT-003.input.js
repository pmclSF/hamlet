import { test, expect } from '@playwright/test';

test.describe('text assertions', () => {
  test('should check text', async ({ page }) => {
    await page.goto('/page');
    await expect(page.locator('#msg')).toHaveText('Hello');
    await expect(page.locator('#msg')).toContainText('Hel');
  });
});
