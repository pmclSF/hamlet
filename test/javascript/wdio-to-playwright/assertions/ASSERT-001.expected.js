import { test, expect } from '@playwright/test';

test.describe('url assertions', () => {
  test('should check url', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL('http://localhost/dashboard');
  });
});
