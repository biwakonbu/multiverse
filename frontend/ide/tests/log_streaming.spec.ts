import { test, expect } from '@playwright/test';

test.describe('E2E: Log Streaming', () => {
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
            GetExecutionState: async () => "RUNNING", // Simulate running state
            StartExecution: async () => {}, 
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
    
    // Assume we have a floating chat window or a log panel where logs appear.
    // Based on functionality, likely there is a log view associated with active task or global log.
    // For this test, we assume the log is visible when execution starts.
  });

  test('should display streamed logs in the log panel', async ({ page }) => {
     // 1. Trigger log chunk event
     const logChunk = "Step 1: Installing dependencies...\n";
     
     await page.evaluate((chunk) => {
         // Assuming 'execution:log' is the event name based on typical pattern, 
         // or it might be specific to task. Let's assume global execution log for phase 4.
         (window as any).runtime.__trigger('execution:log', { 
             source: 'worker', 
             level: 'INFO', 
             message: chunk, 
             timestamp: new Date().toISOString() 
         });
     }, logChunk);

     // 2. Verify log content appears in the UI
     // We need to target the log container. Assuming .log-terminal or similar.
     // If strict selector is unknown, we search by text.
     await expect(page.getByText('Step 1: Installing dependencies...')).toBeVisible();

     // 3. Trigger another chunk
     const logChunk2 = "Step 2: Build started.\n";
     await page.evaluate((chunk) => {
         (window as any).runtime.__trigger('execution:log', { 
             source: 'worker', 
             level: 'INFO', 
             message: chunk, 
             timestamp: new Date().toISOString() 
         });
     }, logChunk2);

     await expect(page.getByText('Step 2: Build started.')).toBeVisible();
  });
});
