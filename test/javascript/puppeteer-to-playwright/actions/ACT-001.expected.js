import { test, expect } from '@playwright/test';

test.describe('type action', () => {

  test('should type text', async ({ page }) => {
    await page.goto('/form');
    await page.locator('#email').fill('user@test.com');
    await page.locator('#password').fill('secret');
  });
});
