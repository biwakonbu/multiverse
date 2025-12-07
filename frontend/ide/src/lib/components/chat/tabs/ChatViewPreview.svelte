<script lang="ts">
  import ChatView from "./ChatView.svelte";
  import {
    chatMessages,
    isChatLoading,
    chatError,
  } from "../../../../stores/chat";

  export let messageCount = 5;
  export let isLoading = false;
  export let error: string | null = null;

  // Generate mock messages
  const now = new Date("2024-01-15T10:00:00Z");
  const mockMessages = Array.from({ length: messageCount }, (_, i) => ({
    id: `msg-${i}`,
    role: i % 2 === 0 ? "user" : "assistant",
    content:
      i % 2 === 0
        ? `ユーザーメッセージ ${i + 1}：これはテスト用のメッセージです。`
        : `アシスタントからの応答 ${i + 1}：ご質問にお答えします。\n\nこれは複数行のレスポンスです。\n詳細な説明を含んでいます。`,
    timestamp: new Date(now.getTime() + i * 60000).toISOString(),
  }));

  // Set mock data to stores
  chatMessages.setMessages(mockMessages as any);
  isChatLoading.set(isLoading);
  chatError.set(error);
</script>

<div class="preview-container">
  <ChatView conflicts={[]} />
</div>

<style>
  .preview-container {
    width: var(--mv-preview-chat-width);
    height: var(--mv-preview-chat-height);
    background: var(--mv-glass-bg-chat);
    border-radius: var(--mv-radius-md);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
</style>
