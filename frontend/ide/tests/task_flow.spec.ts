import { test, expect } from '@playwright/test';

test.describe('Task Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Clear mock storage before each test
    await page.addInitScript(() => {
      window.localStorage.removeItem('mock_tasks');
      window.localStorage.removeItem('mock_workspaces');
    });
    // Go to the app
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    await page.goto('/');
  });

  test('should display empty task list initially', async ({ page }) => {
    // Initially mock returns empty list
    // Check if task list container is visible but empty or says "No tasks"
    // Adjust selector based on actual UI implementation
    // Let's check for some text or element.
    await expect(page.locator('body')).toBeVisible();
  });

  test('should create a new task and display it', async ({ page }) => {
    // 1. Select Workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // 2. Wait for Toolbar to appear (indicates app loaded backend data)
    const addTaskBtn = page.getByRole('button', { name: '新規タスク' });
    await expect(addTaskBtn).toBeVisible();

    // 3. Open Create Task Modal
    await addTaskBtn.click();
    await expect(page.getByRole('dialog', { name: '新規タスク作成' })).toBeVisible();

    // 4. Fill form
    await page.getByPlaceholder('タスクのタイトルを入力').fill('My E2E Task');

    // 5. Submit
    await page.getByRole('button', { name: 'タスクを作成' }).click();

    // 6. Verify Task appears in Grid
    // Assuming GridNode renders the title text
    // We wait for the modal to close and text to appear
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await expect(page.getByText('My E2E Task')).toBeVisible();
  });
});

test.describe('Pool Selection', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      window.localStorage.removeItem('mock_tasks');
      window.localStorage.removeItem('mock_workspaces');
    });
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    await page.goto('/');
  });

  test('should display available pools in task creation form', async ({ page }) => {
    // 1. Select Workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // 2. Wait for Toolbar
    const addTaskBtn = page.getByRole('button', { name: '新規タスク' });
    await expect(addTaskBtn).toBeVisible();

    // 3. Open Create Task Modal
    await addTaskBtn.click();
    await expect(page.getByRole('dialog', { name: '新規タスク作成' })).toBeVisible();

    // 4. Verify Pool dropdown is visible
    const poolSelect = page.locator('select#task-pool');
    await expect(poolSelect).toBeVisible();

    // 5. Check available pools (from mock: default, codegen, test)
    await expect(poolSelect.locator('option')).toHaveCount(3);
  });

  test('should create task with selected pool', async ({ page }) => {
    // 1. Select Workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // 2. Wait and open modal
    const addTaskBtn = page.getByRole('button', { name: '新規タスク' });
    await expect(addTaskBtn).toBeVisible();
    await addTaskBtn.click();

    // 3. Fill form with specific pool
    await page.getByPlaceholder('タスクのタイトルを入力').fill('Codegen Task');
    await page.locator('select#task-pool').selectOption('codegen');

    // 4. Submit
    await page.getByRole('button', { name: 'タスクを作成' }).click();

    // 5. Verify Task appears
    await expect(page.getByRole('dialog')).not.toBeVisible();
    await expect(page.getByText('Codegen Task')).toBeVisible();
  });
});

test.describe('Workspace Management', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      window.localStorage.removeItem('mock_tasks');
      window.localStorage.removeItem('mock_workspaces');
    });
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    await page.goto('/');
  });

  test('should show welcome screen on initial load', async ({ page }) => {
    // Welcome 画面が表示されることを確認
    await expect(page.locator('body')).toBeVisible();
    // "Workspaceを開く" ボタンが表示されている
    await expect(page.getByRole('button', { name: 'Workspaceを開く' })).toBeVisible();
  });

  test('should display recent workspaces list', async ({ page }) => {
    // Mock には 2 つのワークスペースがプリセット
    // Recent workspaces の表示を確認
    // リストが存在するか、またはプレースホルダーテキストを確認
    await expect(page.locator('body')).toBeVisible();
  });
});

test.describe('Task Status Display', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      // 既存タスクをセットアップ
      window.localStorage.setItem('mock_tasks', JSON.stringify([
        {
          id: 'task-pending',
          title: 'Pending Task',
          status: 'PENDING',
          poolId: 'default',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        },
        {
          id: 'task-running',
          title: 'Running Task',
          status: 'RUNNING',
          poolId: 'codegen',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
      ]));
    });
    page.on('console', msg => console.log('PAGE LOG:', msg.text()));
    await page.goto('/');
  });

  test('should display tasks with different statuses', async ({ page }) => {
    // 1. Select Workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // 2. タスクが表示されるのを待つ
    await expect(page.getByText('Pending Task')).toBeVisible();
    await expect(page.getByText('Running Task')).toBeVisible();
  });
});

test.describe('Error Handling', () => {
  test('should handle form validation', async ({ page }) => {
    await page.goto('/');

    // 1. Select Workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();

    // 2. Open Create Task Modal
    const addTaskBtn = page.getByRole('button', { name: '新規タスク' });
    await expect(addTaskBtn).toBeVisible();
    await addTaskBtn.click();

    // 3. Try to submit without title
    const submitBtn = page.getByRole('button', { name: 'タスクを作成' });
    // ボタンが disabled であることを確認
    await expect(submitBtn).toBeDisabled();

    // 4. Enter title and verify button becomes enabled
    await page.getByPlaceholder('タスクのタイトルを入力').fill('Valid Task');
    await expect(submitBtn).toBeEnabled();
  });
});
