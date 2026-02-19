import { test, expect } from '@playwright/test';

test.describe('basic suite', () => {
  test('should work', async ({ page }) => {
    await page.goto('/test');
  });
});
