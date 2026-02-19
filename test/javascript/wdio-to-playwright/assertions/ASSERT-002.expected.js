import { test, expect } from '@playwright/test';

test.describe('visibility assertions', () => {
  test('should check visibility', async ({ page }) => {
    await page.goto('/page');
    await expect(page.locator('#visible')).toBeVisible();
    await expect(page.locator('#hidden')).toBeHidden();
  });
});
