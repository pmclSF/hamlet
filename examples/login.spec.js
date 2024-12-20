import { test, expect } from '@playwright/test';

test.describe('Login Test', () => {
    test('should login successfully', async () => {
      await page.goto('/login');
      await page.locator('[data-test=username]').fill('testuser');
      await page.locator('[data-test=password]').fill('password123');
      await page.locator('[data-test=submit]').click();
      await page.locator('.welcome-message').toBeVisible();
      await page.locator('.welcome-message').toHaveText('Welcome, testuser!');
    });
  });
