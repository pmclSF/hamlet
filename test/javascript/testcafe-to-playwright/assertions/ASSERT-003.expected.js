import { test, expect } from '@playwright/test';

test.describe('Count Assertions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should check count', async ({ page }) => {
  await expect(page.locator('.items')).toHaveCount(5);
});
