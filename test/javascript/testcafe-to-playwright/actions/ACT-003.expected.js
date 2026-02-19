import { test, expect } from '@playwright/test';

test.describe('More Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should hover and press key', async ({ page }) => {
  await page.locator('#menu').hover();
  await page.keyboard.press('enter');
});
