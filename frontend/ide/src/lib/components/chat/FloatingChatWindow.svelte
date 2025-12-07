<script lang="ts">
  import { stopPropagation } from 'svelte/legacy';

  import { createEventDispatcher, onMount } from "svelte";
  import ChatInput from "./ChatInput.svelte";
  import DraggableWindow from "./window/DraggableWindow.svelte";
  import ChatView from "./tabs/ChatView.svelte";
  import LogView from "./tabs/LogView.svelte";
  import {
    chatStore,
    currentSessionId,
    isChatLoading,
    type ChatResponse,
  } from "../../../stores/chat";
  import { get } from "svelte/store";

  let { initialPosition = { x: 20, y: 20 } } = $props();

  let conflicts: NonNullable<ChatResponse["conflicts"]> = $state([]);

  // Tabs
  const tabs = ["General", "Log"];
  let activeTab = $state("General");

  const dispatch = createEventDispatcher();

  // セッション初期化
  onMount(async () => {
    await chatStore.initSession(); // Use initSession to restore or create
  });

  function closeWindow() {
    dispatch("close");
  }

  // チャット送信処理
  async function handleSend(e: CustomEvent<string>) {
    const text = e.detail;
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

    // タスクが生成された場合はイベントを発行
    if (response?.generatedTasks && response.generatedTasks.length > 0) {
      dispatch("tasksGenerated", {
        tasks: response.generatedTasks,
        understanding: response.understanding,
      });
    }
  }
</script>

<DraggableWindow {initialPosition} title="" on:close={closeWindow}>
  {#snippet header()}
    <div  class="tabs">
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

  {#if activeTab === "General"}
    <ChatView {conflicts} />
  {:else if activeTab === "Log"}
    <LogView />
  {/if}

  {#snippet footer()}
    <div >
      <ChatInput on:send={handleSend} disabled={$isChatLoading} />
    </div>
  {/snippet}
</DraggableWindow>

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
