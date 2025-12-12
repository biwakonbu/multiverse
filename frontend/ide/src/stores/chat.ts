/**
 * チャットデータ管理ストア
 */

import { writable, get } from 'svelte/store';
import { GetChatHistory, SendChatMessage, CreateChatSession } from '../../wailsjs/go/main/App';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import type { ChatMessage } from '../types';

export interface ChatResponse {
    message: ChatMessage;
    generatedTasks: Array<{
        id: string;
        title: string;
        status: string;
        // ... 他のフィールド
    }>;
    understanding: string;
    conflicts?: Array<{
        file: string;
        tasks: string[];
        warning: string;
    }>;
    error?: string;
}

// Chat Log Interface
export interface ChatLogEntry {
    step: string;
    message: string;
    timestamp: string;
}

// チャットメッセージストア
function createMessagesStore() {
  const { subscribe, set, update } = writable<ChatMessage[]>([]);

  return {
    subscribe,
    setMessages: (messages: ChatMessage[]) => set(messages),
    addMessage: (message: ChatMessage) => update((msgs) => [...msgs, message]),
    clear: () => set([]),
  };
}

// チャットログストア
function createChatLogStore() {
  const { subscribe, update } = writable<ChatLogEntry[]>([]);

  return {
    subscribe,
    addLog: (entry: ChatLogEntry) => {
      update((logs) => [...logs, entry]);
    },
    clear: () => update(() => []),
  };
}

export const chatMessages = createMessagesStore();
export const chatLog = createChatLogStore();
export const currentSessionId = writable<string | null>(null);
export const isChatLoading = writable<boolean>(false);
export const chatError = writable<string | null>(null);

// ストア
const chatStore = {
    // セッション初期化（起動時に呼び出す）
    initSession: async () => {
        const lastSessionId = localStorage.getItem('lastSessionId');
        if (lastSessionId) {
            currentSessionId.set(lastSessionId);
            chatMessages.clear();
            chatLog.clear();
            chatError.set(null);

            try {
                const history = await GetChatHistory(lastSessionId) as unknown as ChatMessage[];
                chatMessages.setMessages(history);
                console.log('Restored session:', lastSessionId);
                return;
            } catch (e) {
                console.error('Failed to restore chat history, creating new session:', e);
                // 復元失敗時は新規作成へ
            }
        }
        await chatStore.createSession();
    },

    // セッション作成
    createSession: async () => {
        try {
            const session = await CreateChatSession();
            if (!session?.id) return;

            currentSessionId.set(session.id);
            localStorage.setItem('lastSessionId', session.id); // Save to localStorage

            // セッション切替時に既存ログとメッセージをクリア
            chatMessages.clear();
            chatLog.clear();
            chatError.set(null);

            try {
                const history = await GetChatHistory(session.id) as unknown as ChatMessage[];
                chatMessages.setMessages(history);
            } catch (e) {
                console.error('Failed to load chat history:', e);
            }
        } catch (e) {
            console.error('Failed to create session:', e);
        }
    },

    // メッセージ送信
    sendMessage: async (content: string): Promise<ChatResponse | null> => {
        const optimisticId = 'temp-' + Date.now();
        const optimisticMessage: ChatMessage = {
            id: optimisticId,
            role: 'user',
            content: content,
            timestamp: new Date().toISOString()
        };

        // Optimistic UI: 即座に追加
        chatMessages.addMessage(optimisticMessage);

        // 進捗ログはリクエスト単位でリセットして、古いログが残らないようにする
        chatLog.clear();
        
        isChatLoading.set(true);
        chatError.set(null);
        let sessionId: string | null = get(currentSessionId);

        if (!sessionId) {
            console.warn('No active session. Attempting to recreate...');
            await chatStore.createSession();
            sessionId = get(currentSessionId);
            if (!sessionId) {
                console.error('No active session inside sendMessage');
                isChatLoading.set(false);
                chatError.set('No active chat session');
                return null;
            }
        }

        try {
            const response = await SendChatMessage(sessionId, content);

            if (response.error) {
                console.error('Chat error:', response.error);
                chatError.set(response.error);
                // エラー時はロールバック推奨だが、現状はエラー表示のみ
            } else {
                 const history = await GetChatHistory(sessionId!) as unknown as ChatMessage[];
                 chatMessages.setMessages(history);
                 chatError.set(null);
            }

            return response as ChatResponse;

        } catch (e) {
            console.error('Failed to send message:', e);
            chatError.set(e instanceof Error ? e.message : 'Failed to send message');
            // 失敗した場合、Optimistic Message をどうするか？
            // ここでは再取得で整合性を取るのが安全
            try {
                const history = await GetChatHistory(sessionId!) as unknown as ChatMessage[];
                chatMessages.setMessages(history);
            } catch (ignore) {}
            return null;
        } finally {
            isChatLoading.set(false);
        }
    }
};

export { chatStore };

// Wailsイベントリスナーの初期化
export function initChatEvents() {
    EventsOn('chat:progress', (event: { step: string; message: string; timestamp: string }) => {
        console.log('Chat Progress:', event);
        chatLog.addLog({
            step: event.step,
            message: event.message,
            timestamp: event.timestamp
        });
    });
}
