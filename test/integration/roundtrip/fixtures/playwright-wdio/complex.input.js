import { test, expect } from '@playwright/test';

// E2E tests for the settings page with route mocking and multiple locator strategies
test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/user/profile', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          name: 'Alice Johnson',
          email: 'alice@example.com',
          plan: 'pro',
          notifications: { email: true, sms: false },
        }),
      });
    });
    await page.goto('/settings');
  });

  test('should display the user profile information', async ({ page }) => {
    await expect(page.locator('[data-testid="profile-name"]')).toHaveText('Alice Johnson');
    await expect(page.locator('[data-testid="profile-email"]')).toHaveText('alice@example.com');
  });

  test('should show the current subscription plan', async ({ page }) => {
    const planBadge = page.locator('[data-testid="plan-badge"]');
    await expect(planBadge).toBeVisible();
    await expect(planBadge).toHaveText('Pro');
  });

  test.describe('profile editing', () => {
    test('should enable editing when the edit button is clicked', async ({ page }) => {
      await page.locator('[data-testid="edit-profile-button"]').click();
      await expect(page.locator('[data-testid="name-input"]')).toBeEditable();
    });

    test('should save updated profile information', async ({ page }) => {
      await page.route('**/api/user/profile', (route) => {
        if (route.request().method() === 'PUT') {
          route.fulfill({ status: 200, body: JSON.stringify({ success: true }) });
        }
      });
      await page.locator('[data-testid="edit-profile-button"]').click();
      await page.locator('[data-testid="name-input"]').fill('Alice Smith');
      await page.locator('[data-testid="save-button"]').click();
      await expect(page.locator('[data-testid="success-toast"]')).toBeVisible();
    });

    test('should cancel editing and revert changes', async ({ page }) => {
      await page.locator('[data-testid="edit-profile-button"]').click();
      await page.locator('[data-testid="name-input"]').fill('Changed Name');
      await page.locator('[data-testid="cancel-button"]').click();
      await expect(page.locator('[data-testid="profile-name"]')).toHaveText('Alice Johnson');
    });
  });

  test.describe('notification preferences', () => {
    test('should reflect current notification settings', async ({ page }) => {
      const emailToggle = page.locator('[data-testid="notify-email"] input[type="checkbox"]');
      const smsToggle = page.locator('[data-testid="notify-sms"] input[type="checkbox"]');
      await expect(emailToggle).toBeChecked();
      await expect(smsToggle).not.toBeChecked();
    });

    test('should update notification preferences', async ({ page }) => {
      await page.route('**/api/user/notifications', (route) => {
        route.fulfill({ status: 200, body: JSON.stringify({ success: true }) });
      });
      await page.locator('[data-testid="notify-sms"] input[type="checkbox"]').check();
      await page.locator('[data-testid="save-notifications"]').click();
      await expect(page.locator('[data-testid="success-toast"]')).toBeVisible();
    });

    test('should warn before disabling all notifications', async ({ page }) => {
      await page.locator('[data-testid="notify-email"] input[type="checkbox"]').uncheck();
      await expect(page.locator('[data-testid="warning-dialog"]')).toBeVisible();
      await expect(page.locator('[data-testid="warning-dialog"]')).toContainText(
        'You will not receive any notifications'
      );
    });
  });

  test.afterAll(async () => {
    // Clean up any test artifacts created during the run
  });
});
