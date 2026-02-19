import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {;

test('should navigate', async ({ page }) => {
  await page.goto('http://localhost/dashboard');
  await expect(page.locator('#dashboard')).toBeVisible();
});
