import { test, expect } from '@playwright/test';

test.describe('navigation', () => {

  test('should navigate and wait', async ({ page }) => {
    await page.goto('/page1');
    ;
    await page.goto('/page2');
  });
});
