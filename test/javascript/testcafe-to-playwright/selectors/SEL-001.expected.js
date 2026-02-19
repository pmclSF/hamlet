import { test, expect } from '@playwright/test';

test.describe('Selectors', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost/form');
  });

test('should find elements', async ({ page }) => {
  const nameField = page.locator('#name');
  await page.locator(nameField).fill('John');
});
