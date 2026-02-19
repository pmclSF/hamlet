import { test, expect } from '@playwright/test';

test.describe('suite', () => {
  test('should work', async ({ page }) => {
    await page.goto('/test');
  });
});
