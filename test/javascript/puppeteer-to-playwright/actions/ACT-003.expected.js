import { test, expect } from '@playwright/test';

test.describe('waits', () => {

  test('should wait for selector', async ({ page }) => {
    await page.goto('/slow');
    await page.locator('#loaded').waitFor();
    await page.locator('#loaded').click();
  });
});
