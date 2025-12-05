import { writable, type Writable } from 'svelte/store';

// Wails runtime path inferred from list_dir results
// src/stores/chat.ts -> ../../wailsjs/wailsjs/runtime/runtime
// @ts-ignore
import * as runtime from '../../wailsjs/wailsjs/runtime/runtime';

export interface ChatMessage {
    id: string;
    role: 'user' | 'assistant' | 'system';
    content: string;
    timestamp: string;
}

function createChatStore() {
    const { subscribe, update, set } = writable<ChatMessage[]>([]);

    return {
        subscribe,
        
        // Initialize listeners
        init: () => {
            // Check if runtime is available (not available in browser/storybook)
            if (typeof runtime === 'undefined' || !runtime.EventsOn) {
                console.warn('Wails runtime not found. Chat events will be simulated in browser console.');
                // Expose a helper to simulate events in devtools
                // @ts-ignore
                window.simulateChatMessage = (msg: Omit<ChatMessage, 'id'>) => {
                     update(messages => [
                        ...messages,
                        { ...msg, id: crypto.randomUUID() }
                    ]);
                };
                return;
            }

            // Listen for messages from backend
            runtime.EventsOn('chat:message', (payload: any) => {
                console.log('[Chat] Received:', payload);
                // Payload structure depends on backend, assuming { role, content, timestamp? }
                // If timestamp missing, add current time
                const timestamp = payload.timestamp || new Date().toISOString();
                
                update(messages => [
                    ...messages,
                    {
                        id: crypto.randomUUID(),
                        role: payload.role || 'assistant',
                        content: payload.content || '',
                        timestamp: timestamp
                    }
                ]);
            });
        },

        // Send message to backend
        sendMessage: (text: string) => {
            if (!text.trim()) return;

            const userMsg: ChatMessage = {
                id: crypto.randomUUID(),
                role: 'user',
                content: text,
                timestamp: new Date().toISOString()
            };

            // Optimistic update
            update(messages => [...messages, userMsg]);

            if (typeof runtime !== 'undefined' && runtime.EventsEmit) {
                runtime.EventsEmit('chat:send', text);
            } else {
                console.log('[Chat] Sending (simulated):', text);
                // Simulate echo if not in Wails
                setTimeout(() => {
                     update(messages => [
                        ...messages,
                        {
                            id: crypto.randomUUID(),
                            role: 'assistant',
                            content: `(Simulated Backend) Received: ${text}`,
                            timestamp: new Date().toISOString()
                        }
                     ]);
                }, 500);
            }
        }
    };
}

export const chatStore = createChatStore();
