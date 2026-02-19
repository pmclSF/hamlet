import { test, expect } from '@playwright/test';

test.describe('Task Manager', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/tasks', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([
          { id: 1, title: 'Write tests', done: false },
          { id: 2, title: 'Review PR', done: true },
        ]),
      });
    });
    await page.goto('/tasks');
  });

  test('should display all tasks', async ({ page }) => {
    const tasks = page.locator('[data-testid="task-item"]');
    await expect(tasks).toHaveCount(2);
  });

  test('should add a new task', async ({ page }) => {
    await page.locator('[data-testid="task-input"]').fill('Deploy to staging');
    await page.locator('[data-testid="add-task-button"]').click();
    await expect(page.locator('[data-testid="task-item"]')).toHaveCount(3);
  });

  test('should toggle a task as complete', async ({ page }) => {
    const firstTask = page.locator('[data-testid="task-item"]').first();
    await firstTask.locator('input[type="checkbox"]').check();
    await expect(firstTask).toHaveClass(/completed/);
  });

  test('should delete a task', async ({ page }) => {
    const firstTask = page.locator('[data-testid="task-item"]').first();
    await firstTask.locator('[data-testid="delete-button"]').click();
    await expect(page.locator('[data-testid="task-item"]')).toHaveCount(1);
  });
});
