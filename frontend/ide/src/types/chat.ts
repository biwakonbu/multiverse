/**
 * チャット関連の型定義
 */

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string; // ISO 8601 string
  sessionId?: string;
  generatedTasks?: string[]; // IDs of generated tasks
}
