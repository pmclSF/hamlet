import { test, expect } from '@playwright/test';

test.describe('My Suite', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('first test', async ({ page }) => {
  await page.locator('#btn').click();
});

test('second test', async ({ page }) => {
  await page.locator('#other').click();
});
