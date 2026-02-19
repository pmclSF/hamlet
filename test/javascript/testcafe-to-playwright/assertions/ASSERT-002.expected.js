import { test, expect } from '@playwright/test';

test.describe('Text Assertions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should check text', async ({ page }) => {
  await expect(page.locator('#msg')).toHaveText('Hello');
  await expect(page.locator('#msg')).toContainText('Hel');
});
