import { test, expect } from '@playwright/test';

test.describe('existence assertions', () => {
  test('should check existence', async ({ page }) => {
    await page.goto('/page');
    await expect(page.locator('#present')).toBeAttached();
    await expect(page.locator('#absent')).not.toBeAttached();
  });
});
