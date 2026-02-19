import { test, expect } from '@playwright/test';

test.describe('waits', () => {
  test('should wait', async ({ page }) => {
    await page.goto('/slow');
    await page.waitForTimeout(2000);
  });
});
