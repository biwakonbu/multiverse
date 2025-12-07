
import { test, expect } from '@playwright/test';

test.describe('Markdown Task Rendering', () => {
  test.beforeEach(async ({ page }) => {
    // Inject mock tasks with markdown content
    await page.addInitScript(() => {
      window.localStorage.setItem('mock_tasks', JSON.stringify([
        {
          id: 'task-markdown-1',
          title: 'Implement `auth` module',
          status: 'PENDING',
          poolId: 'default',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        },
        {
          id: 'task-markdown-2',
          title: '```Fix login bug```',
          status: 'PENDING',
          poolId: 'default',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        },
        {
            id: 'task-markdown-3',
            title: 'Normal Task',
            status: 'PENDING',
            poolId: 'default',
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
        }
      ]));
      window.localStorage.removeItem('mock_workspaces');
    });
    
    await page.goto('/');
  });

  test('should not display backticks in task titles in Grid View', async ({ page }) => {
    // Open workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // Verify task with inline code backticks
    // Expectation: Backticks should be removed or parsed. If removed: "Implement auth module"
    // For now, let's assume we want to strip them for cleaner display in the small node view.
    const task1 = page.getByTitle('Implement `auth` module'); // The title attribute might still have them
    const task1Text = page.locator('.title', { hasText: 'Implement' }).first();
    
    // We expect the visible text to NOT contain backticks
    await expect(task1Text).not.toHaveText(/`/);
    await expect(task1Text).toHaveText('Implement auth module');

    // Verify task with block code backticks
    const task2Text = page.locator('.title', { hasText: 'Fix login bug' }).first();
    await expect(task2Text).not.toHaveText(/```/);
    await expect(task2Text).toHaveText('Fix login bug');
  });

  // WBS View verification (assuming it might share the same issue)
  test('should not display backticks in task titles in WBS View', async ({ page }) => {
      await page.getByRole('button', { name: 'Workspaceを開く' }).click();
      await page.getByTitle('WBS View').click();

      // Similar check for WBS view
      const task1Text = page.getByText('Implement auth module').first();
      await expect(task1Text).toBeVisible();
      await expect(task1Text).not.toHaveText(/`/);

      const task2Text = page.getByText('Fix login bug').first();
      await expect(task2Text).toBeVisible();
      await expect(task2Text).not.toHaveText(/```/);
  });
});
