import { test, expect } from '@playwright/test';

test.describe('eval selectors', () => {

  test('should evaluate', async ({ page }) => {
    await page.goto('/page');
    const text = await page.locator('#msg').evaluate(el => el.textContent);
    const texts = await page.locator('.items').evaluateAll(els => els.map(el => el.textContent));
  });
});
