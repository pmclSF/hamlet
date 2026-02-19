import { test, expect } from '@playwright/test';

test.describe('mixed setup', () => {

  test.beforeAll(async ({ page }) => {

    await page.goto('/setup');
    await page.setViewportSize({ width: 1280, height: 720 });
  });

  test('should have setup applied', async ({ page }) => {
    await page.locator('#ready').click();
  });
});
