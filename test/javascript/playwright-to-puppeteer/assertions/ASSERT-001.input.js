import { test, expect } from '@playwright/test';

test.describe('url assertions', () => {
  test('should check url', async ({ page }) => {
    await page.goto('http://localhost/dashboard');
    await expect(page).toHaveURL('http://localhost/dashboard');
  });
});
