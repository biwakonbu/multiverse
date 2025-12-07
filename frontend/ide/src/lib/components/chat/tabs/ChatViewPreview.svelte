<script lang="ts">
  import { onMount } from "svelte";
  import ChatView from "./ChatView.svelte";
  import {
    chatMessages,
    isChatLoading,
    chatError,
  } from "../../../../stores/chat";

  interface Props {
    messageCount?: number;
    isLoading?: boolean;
    error?: string | null;
  }

  let { messageCount = 5, isLoading = false, error = null }: Props = $props();

  // Generate mock messages
  function generateMockMessages(count: number) {
    const now = new Date("2024-01-15T10:00:00Z");
    return Array.from({ length: count }, (_, i) => ({
      id: `msg-${i}`,
      role: i % 2 === 0 ? "user" : "assistant",
      content:
        i % 2 === 0
          ? `ユーザーメッセージ ${i + 1}：これはテスト用のメッセージです。`
          : `アシスタントからの応答 ${i + 1}：ご質問にお答えします。\n\nこれは複数行のレスポンスです。\n詳細な説明を含んでいます。`,
      timestamp: new Date(now.getTime() + i * 60000).toISOString(),
    }));
  }

  // Set mock data to stores only on mount
  onMount(() => {
    const mockMessages = generateMockMessages(messageCount);
    chatMessages.setMessages(mockMessages as any);
    isChatLoading.set(isLoading);
    chatError.set(error);
  });
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
