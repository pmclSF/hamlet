import { test, expect } from '@playwright/test';

test.describe('selectors', () => {

  test('should find elements', async ({ page }) => {
    await page.goto('/form');
    const element = page.locator('#username');
    const elements = page.locator('.items');
  });
});
