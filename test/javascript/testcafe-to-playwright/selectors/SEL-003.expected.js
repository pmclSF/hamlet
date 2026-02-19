import { test, expect } from '@playwright/test';

test.describe('Selector Properties', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should check properties', async ({ page }) => {
  await expect(page.locator('#elem')).toBeAttached();
  await expect(page.locator('#elem')).toBeVisible();
  await expect(page.locator('.items')).toHaveCount(3);
});
