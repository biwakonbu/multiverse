<script lang="ts">
  import { stopPropagation } from "svelte/legacy";

  import { onMount } from "svelte";
  import ChatInput from "./ChatInput.svelte";
  import DraggableWindow from "../ui/window/DraggableWindow.svelte";
  import ChatView from "./tabs/ChatView.svelte";
  import LogView from "./tabs/LogView.svelte";
  import LiveLogStream from "../../hud/LiveLogStream.svelte";
  import {
    logs as globalLogs,
    getTaskLogs,
    type LogEntry,
  } from "../../../stores/logStore";
  import { selectedTaskId } from "../../../stores/taskStore";
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
  const tabs = ["General", "Execution", "System"];
  let activeTab = $state("General");

  // Reactive log selection
  // getTaskLogs returns a store, so we subscribe to it via auto-subscription in layout or new derived
  // For simplicity in Svelte 5, we can use a derived store that switches source
  import { derived } from "svelte/store";

  // Create a switching store
  const activeLogStore = derived(
    [selectedTaskId, globalLogs],
    ([$id, $global], set) => {
      if ($id) {
        // Subscribe to specific task logs
        const unsub = getTaskLogs($id).subscribe((val) => set(val));
        return () => unsub();
      } else {
        set($global);
        return () => {};
      }
    },
    [] as LogEntry[] // initial value
  );

  // セッション初期化
  onMount(async () => {
    await chatStore.initSession();
  });

  function closeWindow() {
    windowStore.close("chat");
    onclose?.();
  }

  function handleDragEnd(data: { x: number; y: number }) {
    windowStore.updatePosition("chat", data.x, data.y);
  }

  function handleResizeEnd(data: { width: number; height: number }) {
    windowStore.updateSize("chat", data.width, data.height);
  }

  function handleClick() {
    windowStore.bringToFront("chat");
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
            {#if activeTab === tab}
              <div class="tab-indicator"></div>
            {/if}
          </button>
        {/each}
      </div>
    {/snippet}

    {#snippet children()}
      {#if activeTab === "General"}
        <ChatView {conflicts} />
      {:else if activeTab === "Execution"}
        <div style:height="100%" style:width="100%">
          <LiveLogStream logs={$activeLogStore} height="100%" />
        </div>
      {:else if activeTab === "System"}
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
    gap: var(--mv-space-1);
    height: var(--mv-size-full);
    align-items: center;
  }

  /* Sophisticated Tab Style */
  .tab {
    position: relative;
    padding: var(--mv-space-1-5) var(--mv-space-3);
    font-size: var(--mv-font-size-md);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-muted);
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .tab:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .tab.active {
    color: var(--mv-color-text-primary);
    font-weight: var(--mv-font-weight-semibold);
  }

  /* Animated bottom indicator inside the tab */
  .tab-indicator {
    position: absolute;
    bottom: calc(-1 * var(--mv-spacing-xs));
    left: var(--mv-space-0);
    right: var(--mv-space-0);
    height: var(--mv-border-width-md);
    background: var(--mv-primitive-frost-2);
    box-shadow: var(--mv-shadow-tab-glow);
    border-top-left-radius: var(--mv-border-width-md);
    border-top-right-radius: var(--mv-border-width-md);
  }
</style>
