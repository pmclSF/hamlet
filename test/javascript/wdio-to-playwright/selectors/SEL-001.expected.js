import { test, expect } from '@playwright/test';

test.describe('selectors', () => {
  test('should find elements', async ({ page }) => {
    await page.goto('/form');
    await page.locator('#username').fill('test');
    await page.locator('#submit').click();
  });
});
