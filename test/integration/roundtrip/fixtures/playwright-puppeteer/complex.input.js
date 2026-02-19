import { test, expect } from '@playwright/test';

// E2E tests for the file upload and document management feature
test.describe('Document Manager', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/documents', (route) => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            { id: 'd-1', name: 'Report Q1.pdf', size: 204800, status: 'ready' },
            { id: 'd-2', name: 'Invoice.pdf', size: 102400, status: 'ready' },
          ]),
        });
      }
    });
    await page.goto('/documents');
  });

  test('should list existing documents', async ({ page }) => {
    const rows = page.locator('[data-testid="doc-row"]');
    await expect(rows).toHaveCount(2);
    await expect(rows.first()).toContainText('Report Q1.pdf');
  });

  test('should search documents by name', async ({ page }) => {
    await page.locator('[data-testid="doc-search"]').fill('Invoice');
    await page.waitForSelector('[data-testid="doc-row"]');
    const filtered = page.locator('[data-testid="doc-row"]');
    await expect(filtered).toHaveCount(1);
    await expect(filtered.first()).toContainText('Invoice.pdf');
  });

  test.describe('file upload', () => {
    test('should upload a new document', async ({ page }) => {
      await page.route('**/api/documents/upload', (route) => {
        route.fulfill({
          status: 201,
          body: JSON.stringify({ id: 'd-3', name: 'NewFile.pdf', status: 'processing' }),
        });
      });
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'NewFile.pdf',
        mimeType: 'application/pdf',
        buffer: Buffer.from('fake-pdf-content'),
      });
      await page.locator('[data-testid="upload-button"]').click();
      await expect(page.locator('[data-testid="upload-success"]')).toBeVisible();
    });

    test('should show a progress indicator during upload', async ({ page }) => {
      await page.route('**/api/documents/upload', async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 500));
        route.fulfill({ status: 201, body: JSON.stringify({ id: 'd-4' }) });
      });
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'LargeFile.pdf',
        mimeType: 'application/pdf',
        buffer: Buffer.from('large-fake-content'),
      });
      await page.locator('[data-testid="upload-button"]').click();
      await expect(page.locator('[data-testid="upload-progress"]')).toBeVisible();
    });

    test('should reject files that exceed the size limit', async ({ page }) => {
      const fileInput = page.locator('input[type="file"]');
      await fileInput.setInputFiles({
        name: 'Huge.pdf',
        mimeType: 'application/pdf',
        buffer: Buffer.alloc(50 * 1024 * 1024),
      });
      await page.locator('[data-testid="upload-button"]').click();
      await expect(page.locator('[data-testid="size-error"]')).toContainText('exceeds the maximum');
    });
  });

  test.describe('document actions', () => {
    test('should download a document', async ({ page }) => {
      const [download] = await Promise.all([
        page.waitForEvent('download'),
        page.locator('[data-testid="doc-row"]').first().locator('[data-testid="download-button"]').click(),
      ]);
      expect(download.suggestedFilename()).toBe('Report Q1.pdf');
    });

    test('should delete a document after confirmation', async ({ page }) => {
      await page.route('**/api/documents/d-1', (route) => {
        if (route.request().method() === 'DELETE') {
          route.fulfill({ status: 204 });
        }
      });
      await page.locator('[data-testid="doc-row"]').first().locator('[data-testid="delete-button"]').click();
      await page.locator('[data-testid="confirm-delete"]').click();
      await expect(page.locator('[data-testid="doc-row"]')).toHaveCount(1);
    });
  });
});
