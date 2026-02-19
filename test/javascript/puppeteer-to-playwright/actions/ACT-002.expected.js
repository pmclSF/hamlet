import { test, expect } from '@playwright/test';

test.describe('click action', () => {

  test('should click', async ({ page }) => {
    await page.goto('/app');
    await page.locator('#submit').click();
    await page.locator('#menu').hover();
  });
});
