import { test, expect } from '@playwright/test';

test.describe('navigation', () => {

  test('should navigate', async ({ page }) => {
    await page.goto('/dashboard');
  });
});
