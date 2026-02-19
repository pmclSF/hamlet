// Playwright test for an analytics dashboard
// Inspired by real-world Playwright tests for admin panels and dashboards

import { test, expect } from '@playwright/test';

test.describe('Analytics Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Mock the analytics API to return consistent data
    await page.route('**/api/analytics/summary', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          totalUsers: 12450,
          activeUsers: 3200,
          revenue: 89750.50,
          conversionRate: 3.7,
        }),
      });
    });

    await page.route('**/api/analytics/chart*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          labels: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri'],
          datasets: [{ label: 'Visitors', data: [120, 200, 150, 300, 280] }],
        }),
      });
    });

    await page.goto('/dashboard');
  });

  test('should display the summary statistics cards', async ({ page }) => {
    await expect(page.getByText('Total Users')).toBeVisible();
    await expect(page.locator('[data-testid="stat-total-users"]')).toHaveText('12,450');
    await expect(page.locator('[data-testid="stat-active-users"]')).toHaveText('3,200');
    await expect(page.locator('[data-testid="stat-revenue"]')).toContainText('$89,750');
  });

  test('should render the visitors chart', async ({ page }) => {
    const chart = page.locator('[data-testid="visitors-chart"]');
    await expect(chart).toBeVisible();
    await expect(chart.locator('canvas')).toBeAttached();
  });

  test('should filter analytics data by date range', async ({ page }) => {
    await page.getByRole('button', { name: 'Date Range' }).click();
    await page.getByRole('option', { name: 'Last 7 days' }).click();

    // Verify the API was called with the correct query parameter
    const requestPromise = page.waitForRequest((req) =>
      req.url().includes('/api/analytics/summary') && req.url().includes('range=7d')
    );
    await requestPromise;
  });

  test('should navigate to the detailed report page', async ({ page }) => {
    await page.getByRole('link', { name: 'View Full Report' }).click();
    await expect(page).toHaveURL(/\/dashboard\/reports/);
    await expect(page.getByRole('heading', { name: 'Detailed Report' })).toBeVisible();
  });

  test('should toggle between chart and table views', async ({ page }) => {
    const chartView = page.locator('[data-testid="visitors-chart"]');
    const tableView = page.locator('[data-testid="visitors-table"]');

    await expect(chartView).toBeVisible();
    await expect(tableView).not.toBeVisible();

    await page.getByRole('tab', { name: 'Table' }).click();

    await expect(chartView).not.toBeVisible();
    await expect(tableView).toBeVisible();
    await expect(tableView.locator('tr')).toHaveCount(6); // header + 5 data rows
  });

  test('should export dashboard data as CSV', async ({ page }) => {
    const downloadPromise = page.waitForEvent('download');
    await page.getByRole('button', { name: 'Export CSV' }).click();
    const download = await downloadPromise;

    expect(download.suggestedFilename()).toContain('analytics');
    expect(download.suggestedFilename()).toMatch(/\.csv$/);
  });

  test('should show an error state when the API fails', async ({ page }) => {
    await page.route('**/api/analytics/summary', (route) => {
      route.fulfill({ status: 500, body: 'Internal Server Error' });
    });

    await page.reload();

    await expect(page.getByText('Failed to load dashboard data')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Retry' })).toBeVisible();
  });
});
