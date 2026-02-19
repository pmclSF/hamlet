import { test, expect } from '@playwright/test';

test.describe('Home Page', () => {
  test('should display the welcome heading', async ({ page }) => {
    await page.goto('/');
    const heading = page.locator('h1');
    await expect(heading).toBeVisible();
    await expect(heading).toHaveText('Welcome');
  });

  test('should navigate to the about page', async ({ page }) => {
    await page.goto('/');
    await page.locator('nav >> a:has-text("About")').click();
    await expect(page).toHaveURL(/\/about$/);
  });
});
