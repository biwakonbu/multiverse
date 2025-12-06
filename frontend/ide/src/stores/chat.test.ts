import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest';
import { get } from 'svelte/store';
import {
  chatStore,
  chatMessages,
  currentSessionId,
  chatError,
  isChatLoading,
} from './chat';

vi.mock('../../wailsjs/go/main/App', () => ({
  CreateChatSession: vi.fn(),
  GetChatHistory: vi.fn(),
  SendChatMessage: vi.fn(),
}));

import {
  CreateChatSession,
  GetChatHistory,
  SendChatMessage,
} from '../../wailsjs/go/main/App';

describe('chat store', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    chatMessages.clear();
    currentSessionId.set(null);
    chatError.set(null);
    isChatLoading.set(false);
  });

  it('createSession loads history and clears error/logs', async () => {
    (CreateChatSession as Mock).mockResolvedValue({ id: 'session-1' });
    (GetChatHistory as Mock).mockResolvedValue([
      {
        id: 'm1',
        role: 'system',
        content: 'hello',
        timestamp: new Date().toISOString(),
      },
    ]);

    await chatStore.createSession();

    expect(CreateChatSession).toHaveBeenCalledTimes(1);
    expect(GetChatHistory).toHaveBeenCalledWith('session-1');
    expect(get(currentSessionId)).toBe('session-1');
    expect(get(chatMessages).length).toBe(1);
    expect(get(chatError)).toBeNull();
  });

  it('sendMessage sets error when backend returns error', async () => {
    currentSessionId.set('session-err');
    (SendChatMessage as Mock).mockResolvedValue({ error: 'fail' });

    await chatStore.sendMessage('ping');

    expect(SendChatMessage).toHaveBeenCalledWith('session-err', 'ping');
    expect(get(chatError)).toBe('fail');
    expect(get(isChatLoading)).toBe(false);
  });

  it('sendMessage clears error and refreshes history on success', async () => {
    currentSessionId.set('session-ok');
    chatError.set('previous error');
    (SendChatMessage as Mock).mockResolvedValue({
      message: { id: 'm2', role: 'assistant', content: 'ok', timestamp: new Date().toISOString() },
      generatedTasks: [],
      understanding: '',
    });
    (GetChatHistory as Mock).mockResolvedValue([
      {
        id: 'm1',
        role: 'user',
        content: 'hello',
        timestamp: new Date().toISOString(),
      },
    ]);

    await chatStore.sendMessage('ping');

    expect(SendChatMessage).toHaveBeenCalledWith('session-ok', 'ping');
    expect(GetChatHistory).toHaveBeenCalledWith('session-ok');
    expect(get(chatMessages).length).toBe(1);
    expect(get(chatError)).toBeNull();
    expect(get(isChatLoading)).toBe(false);
  });
});
