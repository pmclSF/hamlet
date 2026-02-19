import { test, expect } from '@playwright/test';

test.describe('attribute and count assertions', () => {
  test('should check attributes and counts', async ({ page }) => {
    await page.goto('/page');
    await expect(page.locator('#link')).toHaveAttribute('href', '/home');
    await expect(page.locator('.item')).toHaveCount(3);
  });
});
