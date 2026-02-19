import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost/home');
  });

test('should load page', async ({ page }) => {
  await expect(page.locator('#content')).toBeVisible();
});
