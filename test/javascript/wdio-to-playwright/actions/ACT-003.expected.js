import { test, expect } from '@playwright/test';

test.describe('misc actions', () => {
  test('should perform various actions', async ({ page }) => {
    await page.goto('/form');
    await page.locator('#field').clear();
    await page.locator('#hover-target').hover();
    await page.locator('#dropdown').selectOption({ label: 'Option A' });
  });
});
