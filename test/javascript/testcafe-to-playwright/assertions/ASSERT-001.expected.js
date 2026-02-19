import { test, expect } from '@playwright/test';

test.describe('Assertions', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost');
  });

test('should check visibility', async ({ page }) => {
  await expect(page.locator('#visible')).toBeVisible();
  await expect(page.locator('#hidden')).toBeHidden();
  await expect(page.locator('#present')).toBeAttached();
  await expect(page.locator('#absent')).not.toBeAttached();
});
