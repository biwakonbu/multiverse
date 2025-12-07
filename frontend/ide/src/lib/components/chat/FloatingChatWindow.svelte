<script lang="ts">
  import { stopPropagation } from 'svelte/legacy';

  import { onMount } from "svelte";
  import ChatInput from "./ChatInput.svelte";
  import DraggableWindow from "../ui/window/DraggableWindow.svelte";
  import ChatView from "./tabs/ChatView.svelte";
  import LogView from "./tabs/LogView.svelte";
  import {
    chatStore,
    currentSessionId,
    isChatLoading,
    type ChatResponse,
  } from "../../../stores/chat";
  import { get } from "svelte/store";
  import { windowStore } from "../../../stores/windowStore";
  import type { Task } from "../../../types";

  interface Props {
    onclose?: () => void;
    ontasksGenerated?: (data: { tasks: Task[]; understanding: string }) => void;
  }

  let { onclose, ontasksGenerated }: Props = $props();

  let isOpen = $derived($windowStore.chat.isOpen);
  let position = $derived($windowStore.chat.position);
  let size = $derived($windowStore.chat.size);
  let zIndex = $derived($windowStore.chat.zIndex);

  let conflicts: NonNullable<ChatResponse["conflicts"]> = $state([]);

  // Tabs
  const tabs = ["General", "Log"];
  let activeTab = $state("General");

  // セッション初期化
  onMount(async () => {
    await chatStore.initSession();
  });

  function closeWindow() {
    windowStore.close('chat');
    onclose?.();
  }

  function handleMinimize(data: { minimized: boolean }) {
    windowStore.minimize('chat', data.minimized);
  }

  function handleDragEnd(data: { x: number; y: number }) {
    windowStore.updatePosition('chat', data.x, data.y);
  }

  function handleClick() {
    windowStore.bringToFront('chat');
  }

  // チャット送信処理
  async function handleSend(text: string) {
    if (!text.trim()) {
      console.warn("FloatingChatWindow: empty text");
      return;
    }

    const currentId = get(currentSessionId);

    // セッションIDがない場合は再作成を試みる
    if (!currentId) {
      console.warn("Session ID is missing, attempting to recreate...");
      await chatStore.createSession();
      if (!get(currentSessionId)) {
        console.error("Failed to create session.");
        alert(
          "Chat session initialization failed. Please try reloading the workspace."
        );
        return;
      }
    }

    const response = await chatStore.sendMessage(text);
    conflicts = response?.conflicts ?? [];

    // タスクが生成された場合はコールバックを呼び出す
    if (response?.generatedTasks && response.generatedTasks.length > 0) {
      ontasksGenerated?.({
        tasks: response.generatedTasks as Task[],
        understanding: response.understanding ?? "",
      });
    }
  }
</script>

{#if isOpen}
  <DraggableWindow
    id="chat"
    title=""
    initialPosition={position}
    initialSize={size}
    {zIndex}
    onclose={closeWindow}
    onminimize={handleMinimize}
    ondragend={handleDragEnd}
    onclick={handleClick}
  >
    {#snippet header()}
      <div class="tabs">
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        {#each tabs as tab}
          <button
            class="tab"
            class:active={activeTab === tab}
            onclick={stopPropagation(() => (activeTab = tab))}
            type="button"
          >
            {tab}
          </button>
        {/each}
      </div>
    {/snippet}

    {#snippet children()}
      {#if activeTab === "General"}
        <ChatView {conflicts} />
      {:else if activeTab === "Log"}
        <LogView />
      {/if}
    {/snippet}

    {#snippet footer()}
      <div>
        <ChatInput onsend={handleSend} disabled={$isChatLoading} />
      </div>
    {/snippet}
  </DraggableWindow>
{/if}

<style>
  .tabs {
    display: flex;
    gap: var(--mv-spacing-sm);
  }

  /* Glass Tabs */
  .tab {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-muted);
    background: transparent;
    border: var(--mv-border-width-thin) solid transparent;
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-tab);
    transition: all var(--mv-duration-fast);
  }

  .tab:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .tab.active {
    color: var(--mv-primitive-frost-1);
    background: var(--mv-glass-tab-bg);
    border-color: var(--mv-glass-border-tab);
    box-shadow: var(--mv-shadow-tab-glow);
  }
</style>
