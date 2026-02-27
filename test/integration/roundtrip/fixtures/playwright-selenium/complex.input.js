import { test, expect } from '@playwright/test';

test.describe('User Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.locator('[data-testid="email-input"]').fill('admin@example.com');
    await page.locator('[data-testid="password-input"]').fill('adminPass');
    await page.locator('[data-testid="login-button"]').click();
  });

  test('should display the dashboard after login', async ({ page }) => {
    await expect(page.locator('[data-testid="dashboard"]')).toBeVisible();
    await expect(page.locator('[data-testid="welcome-message"]')).toContainText('admin');
  });

  test('should show user statistics', async ({ page }) => {
    await expect(page.locator('[data-testid="stats-panel"]')).toBeVisible();
    await expect(page.locator('[data-testid="total-users"]')).toBeVisible();
  });

  test.describe('user management', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('[data-testid="nav-users"]').click();
    });

    test('should list all users', async ({ page }) => {
      await expect(page.locator('[data-testid="user-row"]')).toHaveCount(5);
    });

    test('should search users by name', async ({ page }) => {
      await page.locator('[data-testid="user-search"]').fill('Alice');
      await expect(page.locator('[data-testid="user-row"]')).toHaveCount(1);
      await expect(page.locator('[data-testid="user-row"]').first()).toContainText('Alice');
    });

    test('should open user detail view', async ({ page }) => {
      await page.locator('[data-testid="user-row"]').first().click();
      await expect(page.locator('[data-testid="user-detail"]')).toBeVisible();
      await expect(page.locator('[data-testid="user-email"]')).toBeVisible();
    });

    test('should delete a user after confirmation', async ({ page }) => {
      await page.locator('[data-testid="user-row"]').first().locator('[data-testid="delete-btn"]').click();
      await page.locator('[data-testid="confirm-delete"]').click();
      await expect(page.locator('[data-testid="user-row"]')).toHaveCount(4);
    });
  });

  test.describe('settings', () => {
    test.beforeEach(async ({ page }) => {
      await page.locator('[data-testid="nav-settings"]').click();
    });

    test('should update the display name', async ({ page }) => {
      await page.locator('[data-testid="display-name"]').clear();
      await page.locator('[data-testid="display-name"]').fill('New Admin Name');
      await page.locator('[data-testid="save-settings"]').click();
      await expect(page.locator('[data-testid="save-success"]')).toBeVisible();
    });

    test('should toggle dark mode', async ({ page }) => {
      await page.locator('[data-testid="dark-mode-toggle"]').click();
      await expect(page.locator('body')).toHaveClass(/dark/);
    });
  });
});
