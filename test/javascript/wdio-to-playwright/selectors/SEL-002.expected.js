import { test, expect } from '@playwright/test';

test.describe('text selectors', () => {
  test('should find by text', async ({ page }) => {
    await page.goto('/home');
    await page.getByText('Sign In').click();
    await page.getByText('Sign').click();
  });
});
