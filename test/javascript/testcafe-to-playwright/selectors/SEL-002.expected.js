import { test, expect } from '@playwright/test';

test.describe('Selectors', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should filter by text', async ({ page }) => {
  await page.locator(page.locator('.btn').click().withText('Submit'));
});
