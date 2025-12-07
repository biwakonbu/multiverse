import { test, expect } from '@playwright/test';

test.describe('E2E: CLI Session Validation', () => {
  test.beforeEach(async ({ page }) => {
    // 1. Mock Wails Runtime & Backend
    await page.addInitScript(() => {
      window.localStorage.clear();

      const listeners = new Map<string, Function[]>();
      (window as any).runtime = {
        EventsOn: (eventName: string, callback: Function) => {
          if (!listeners.has(eventName)) listeners.set(eventName, []);
          listeners.get(eventName)?.push(callback);
        },
        EventsOff: () => {},
        __trigger: (eventName: string, data: any) => {
          const callbacks = listeners.get(eventName) || [];
          callbacks.forEach(cb => cb(data));
        }
      };
      
      (window as any).go = {
        main: {
          App: {
            ListTasks: async () => [],
            GetPoolSummaries: async () => [],
            GetExecutionState: async () => "IDLE",
            StartExecution: async () => {}, // Default success
            StopExecution: async () => {},
            PauseExecution: async () => {},
            ResumeExecution: async () => {},
          }
        }
      };
    });

    await page.goto('/');
    
    // Open workspace
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();
    await expect(page.getByRole('button', { name: 'Start' })).toBeVisible();
  });

  test('should show error notification when CLI session is missing', async ({ page }) => {
     // 1. Simulate CLI session missing error from backend
     const errorMessage = "failed to build agent tool plan: CLI session not found (please run 'codex login')";
     
     await page.evaluate((msg) => {
         (window as any).go.main.App.StartExecution = () => Promise.reject(msg);
     }, errorMessage);

     // 2. Click Start button
     await page.getByRole('button', { name: 'Start' }).click();

     // 3. Verify Error Toast appears with specific message
     // Check for toast container
     const toast = page.locator('.toast.error');
     await expect(toast).toBeVisible();
     
     // Check for error message text
     await expect(toast).toContainText('CLI session not found');
  });
});
