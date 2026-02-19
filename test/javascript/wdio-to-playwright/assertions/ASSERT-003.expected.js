import { test, expect } from '@playwright/test';

test.describe('text and value assertions', () => {
  test('should check text and value', async ({ page }) => {
    await page.goto('/form');
    await expect(page.locator('#label')).toHaveText('Hello');
    await expect(page.locator('#label')).toContainText('Hel');
    await expect(page.locator('#input')).toHaveValue('test');
  });
});
