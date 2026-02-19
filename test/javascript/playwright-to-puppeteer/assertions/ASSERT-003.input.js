import { test, expect } from '@playwright/test';

test.describe('text assertions', () => {
  test('should check text and value', async ({ page }) => {
    await page.goto('/form');
    await expect(page.locator('#msg')).toHaveText('Hello');
    await expect(page.locator('#input')).toHaveValue('test');
  });
});
