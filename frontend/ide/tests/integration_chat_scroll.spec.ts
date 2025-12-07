import { test, expect } from '@playwright/test';

test.describe('Integration: Chat Scroll Behavior', () => {
  test.beforeEach(async ({ page }) => {
    // Mock Wails Runtime & LocalStorage
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
      
      // Mock chat messages storage
      const chatMessages: any[] = [];
      
      (window as any).go = {
        main: {
          App: {
            ListTasks: async () => [],
            GetPoolSummaries: async () => [],
            GetExecutionState: async () => "IDLE",
            StartExecution: async () => {},
            StopExecution: async () => {},
            PauseExecution: async () => {},
            ResumeExecution: async () => {},
            CreateChatSession: async () => ({ id: 'test-session-id' }),
            GetChatHistory: async () => chatMessages,
            SendChatMessage: async (_sessionId: string, content: string) => {
              // Add user message
              chatMessages.push({
                id: `msg-user-${Date.now()}`,
                role: 'user',
                content: content,
                timestamp: new Date().toISOString()
              });
              // Add assistant response (simulated)
              chatMessages.push({
                id: `msg-assistant-${Date.now()}`,
                role: 'assistant',
                content: `Response to: ${content}\n\nThis is a multi-line response.\nLine 2\nLine 3\nLine 4\nLine 5`,
                timestamp: new Date().toISOString()
              });
              return {
                message: chatMessages[chatMessages.length - 1],
                generatedTasks: [],
                understanding: ''
              };
            }
          }
        }
      };
    });

    await page.goto('/');
    await page.getByRole('button', { name: 'Workspaceを開く' }).click();
    await expect(page.getByRole('button', { name: 'Start' })).toBeVisible();
  });

  test('chat-view should have scrollable area with min-height: 0', async ({ page }) => {
    // Verify the chat-view element has the correct CSS properties for scrolling
    const chatView = page.locator('.chat-view');
    await expect(chatView).toBeVisible();
    
    // Check computed styles
    const styles = await chatView.evaluate((el) => {
      const computed = window.getComputedStyle(el);
      return {
        minHeight: computed.minHeight,
        overflowY: computed.overflowY,
        display: computed.display,
        flexDirection: computed.flexDirection,
      };
    });
    
    // min-height: 0 is required for flex items to shrink and enable scrolling
    expect(styles.minHeight).toBe('0px');
    expect(styles.overflowY).toBe('auto');
    expect(styles.display).toBe('flex');
    expect(styles.flexDirection).toBe('column');
  });

  test('chat should auto-scroll to bottom when messages are added', async ({ page }) => {
    const chatView = page.locator('.chat-view');
    const chatInput = page.locator('.chat-input textarea, .chat-input input');
    
    // Send multiple messages to create overflow
    for (let i = 0; i < 5; i++) {
      await chatInput.fill(`Test message ${i + 1} - This is a longer message to ensure overflow happens in the chat window area.`);
      await chatInput.press('Enter');
      // Wait for message to appear and scroll to complete
      await page.waitForTimeout(500);
    }
    
    // The chat view should be scrolled to the bottom
    const scrollInfo = await chatView.evaluate((el) => {
      return {
        scrollTop: el.scrollTop,
        scrollHeight: el.scrollHeight,
        clientHeight: el.clientHeight,
        isScrolledToBottom: Math.abs((el.scrollTop + el.clientHeight) - el.scrollHeight) < 5
      };
    });
    
    // If there's scrollable content, scrollTop should be near the bottom
    if (scrollInfo.scrollHeight > scrollInfo.clientHeight) {
      expect(scrollInfo.isScrolledToBottom).toBe(true);
    }
  });

  test('DraggableWindow .content should not have overflow-y: auto (to avoid nested scroll)', async ({ page }) => {
    // The .content element inside DraggableWindow should NOT have overflow-y: auto
    // to prevent nested scroll containers
    const content = page.locator('.floating-window .content');
    await expect(content).toBeVisible();
    
    const styles = await content.evaluate((el) => {
      const computed = window.getComputedStyle(el);
      return {
        overflow: computed.overflow,
        overflowY: computed.overflowY,
        minHeight: computed.minHeight,
      };
    });
    
    // .content should have overflow: hidden (not auto) to let child handle scrolling
    expect(styles.overflowY).toBe('hidden');
    expect(styles.minHeight).toBe('0px');
  });
});
