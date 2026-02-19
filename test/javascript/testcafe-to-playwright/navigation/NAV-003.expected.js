import { test, expect } from '@playwright/test';

test.describe('Waits', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should wait', async ({ page }) => {
  await page.waitForTimeout(2000);
});
