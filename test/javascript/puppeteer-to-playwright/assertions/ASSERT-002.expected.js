import { test, expect } from '@playwright/test';

test.describe('element assertions', () => {

  test('should check element exists', async ({ page }) => {
    await page.goto('/page');
    await expect(page.locator('#visible')).toBeVisible();
    await expect(page.locator('#hidden')).toBeHidden();
  });
});
