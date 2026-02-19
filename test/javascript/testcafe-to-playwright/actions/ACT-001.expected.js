import { test, expect } from '@playwright/test';

test.describe('Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost/form');
  });

test('should type text', async ({ page }) => {
  await page.locator('#email').fill('user@test.com');
  await page.locator('#password').fill('secret');
});
