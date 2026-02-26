import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test('should display the login form', async ({ page }) => {
    await page.goto('/login');
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  });

  test('should log in with valid credentials', async ({ page }) => {
    await page.goto('/login');
    await page.locator('[data-testid="email-input"]').fill('user@example.com');
    await page.locator('[data-testid="password-input"]').fill('password123');
    await page.locator('[data-testid="login-button"]').click();
    await expect(page).toHaveURL('/dashboard');
  });
});
