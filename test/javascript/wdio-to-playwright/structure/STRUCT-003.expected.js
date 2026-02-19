import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should login successfully', async ({ page }) => {
    await page.locator('#username').fill('admin');
    await page.locator('#password').fill('pass123');
    await page.locator('#login-btn').click();
    await expect(page).toHaveURL('http://localhost/dashboard');
    await expect(page.locator('#welcome')).toBeVisible();
    await expect(page.locator('#welcome')).toHaveText('Welcome, admin');
  });
});
