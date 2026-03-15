import { test, expect } from '@playwright/test';

test.describe('Checkout Flow', () => {
  test('should complete purchase', async ({ page }) => {
    await page.goto('/checkout');
    await page.fill('#email', 'user@example.com');
    await page.click('button[type="submit"]');
    await expect(page.locator('.success')).toBeVisible();
  });

  test('should show validation errors', async ({ page }) => {
    await page.goto('/checkout');
    await page.click('button[type="submit"]');
    await expect(page.locator('.error')).toBeVisible();
  });
});
