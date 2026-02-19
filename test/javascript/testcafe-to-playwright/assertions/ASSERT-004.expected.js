import { test, expect } from '@playwright/test';

test.describe('Value Assertions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost/form');
  });

test('should check value', async ({ page }) => {
  await page.locator('#input').fill('test');
  await expect(page.locator('#input')).toHaveValue('test');
});
