import { test, expect } from '@playwright/test';

test.describe('actions', () => {
  test('should type values', async ({ page }) => {
    await page.goto('/form');
    await page.locator('#email').fill('user@test.com');
    await page.locator('#field').clear();
  });
});
