import { test, expect } from '@playwright/test';

test.describe('Click Actions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost/app');
  });

test('should click', async ({ page }) => {
  await page.locator('#submit').click();
  await page.locator('#double').dblclick();
  await page.locator('#context').click({ button: 'right' });
});
