import { test, expect } from '@playwright/test';

test.describe('Contact Form', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/contact');
  });

  test('should display the contact form fields', async ({ page }) => {
    await expect(page.locator('[data-testid="contact-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="contact-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="contact-message"]')).toBeVisible();
  });

  test('should show validation errors on empty submit', async ({ page }) => {
    await page.locator('[data-testid="submit-button"]').click();
    await expect(page.locator('[data-testid="name-error"]')).toHaveText('Name is required');
    await expect(page.locator('[data-testid="email-error"]')).toHaveText('Email is required');
  });

  test('should fill and submit the form successfully', async ({ page }) => {
    await page.locator('[data-testid="contact-name"]').fill('John Doe');
    await page.locator('[data-testid="contact-email"]').fill('john@example.com');
    await page.locator('[data-testid="contact-message"]').fill('I have a question about pricing.');
    await page.locator('[data-testid="submit-button"]').click();
    await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
  });

  test('should clear the form after successful submission', async ({ page }) => {
    await page.locator('[data-testid="contact-name"]').fill('John Doe');
    await page.locator('[data-testid="contact-email"]').fill('john@example.com');
    await page.locator('[data-testid="contact-message"]').fill('Hello');
    await page.locator('[data-testid="submit-button"]').click();
    await expect(page.locator('[data-testid="contact-name"]')).toHaveValue('');
    await expect(page.locator('[data-testid="contact-email"]')).toHaveValue('');
  });
});
