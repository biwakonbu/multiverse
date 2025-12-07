<script lang="ts">
  import { run } from 'svelte/legacy';

  /**
   * MainViewPreview - メインビュー全体のプレビューコンポーネント
   *
   * App.svelte のワークスペース選択後の状態を再現
   * Store/Wails依存を排除し、propsのみで動作
   */
  import { createEventDispatcher } from "svelte";
  import ToolbarPreview from "./toolbar/ToolbarPreview.svelte";
  import { WBSListView, WBSGraphView } from "./wbs";
  import GridCanvas from "./grid/GridCanvas.svelte";
  import FloatingChatWindow from "./components/chat/FloatingChatWindow.svelte";
  import { tasks, selectedTaskId } from "../stores/taskStore";
  import type { Task, TaskStatus, PoolSummary } from "../types";
  import { MessageSquare } from "lucide-svelte";

  const dispatch = createEventDispatcher();

  import BacklogPanelPreview from "./backlog/BacklogPanelPreview.svelte";
  import LLMSettingsPreview from "./settings/LLMSettingsPreview.svelte";

  // === Props ===

  

  

  

  

  

  
  interface Props {
    // ビュー設定
    viewMode?: "graph" | "wbs";
    // タスクデータ
    taskList?: Task[];
    poolSummaries?: PoolSummary[];
    // 進捗
    overallProgress?: any;
    // ステータス別カウント
    taskCountsByStatus?: Record<TaskStatus, number>;
    // 選択中タスク（ストア同期用。UI描画では未使用）
    selectedTask?: Task | null;
    // モーダル・チャット・バックログ
    showChat?: boolean;
    showBacklog?: boolean;
    unresolvedCount?: number;
    showSettings?: boolean;
  }

  let {
    viewMode = "wbs",
    taskList = [],
    poolSummaries = [],
    overallProgress = { percentage: 0, completed: 0, total: 0 },
    taskCountsByStatus = {
    PENDING: 0,
    READY: 0,
    RUNNING: 0,
    SUCCEEDED: 0,
    COMPLETED: 0,
    FAILED: 0,
    CANCELED: 0,
    BLOCKED: 0,
    RETRY_WAIT: 0,
  },
    selectedTask = null,
    showChat = true,
    showBacklog = $bindable(false),
    unresolvedCount = 0,
    showSettings = $bindable(false)
  }: Props = $props();

  // タスクストアを更新
  run(() => {
    tasks.setTasks(taskList);
    if (selectedTask) {
      selectedTaskId.select(selectedTask.id);
    } else {
      selectedTaskId.clear();
    }
  });

  function handleCloseChat() {
    dispatch("closeChat");
  }

  function handleOpenChat() {
    dispatch("openChat");
  }
</script>

<main class="app">
  <!-- ツールバー -->
  <div class="toolbar-overlay">
    <ToolbarPreview
      {viewMode}
      {overallProgress}
      {poolSummaries}
      {taskCountsByStatus}
    />
  </div>

  <!-- メインコンテンツ -->
  <div class="main-content">
    {#if viewMode === "graph"}
      <!-- Graph モード: GridCanvas で依存グラフ表示 -->
      <div class="canvas-layer">
        <GridCanvas />
      </div>
    {:else}
      <!-- WBS モード: WBSGraphView + WBSListView -->
      <div class="canvas-layer">
        <WBSGraphView />
      </div>
      <div class="list-overlay">
        <WBSListView />
      </div>
    {/if}
  </div>

  <!-- バックログ表示ボタン -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="backlog-fab"
    class:has-items={unresolvedCount > 0}
    onclick={() => (showBacklog = !showBacklog)}
    onkeydown={(e) => e.key === "Enter" && (showBacklog = !showBacklog)}
    role="button"
    tabindex="0"
    aria-label="Toggle Backlog"
  >
    {#if unresolvedCount > 0}
      <span class="backlog-count">{unresolvedCount}</span>
    {:else}
      &#9776;
    {/if}
  </div>

  <!-- バックログパネル -->
  {#if showBacklog}
    <div class="backlog-sidebar">
      <BacklogPanelPreview />
    </div>
  {/if}

  <!-- チャットウィンドウ -->
  {#if showChat}
    <FloatingChatWindow onclose={handleCloseChat} />
  {/if}

  <!-- チャット再表示ボタン -->
  {#if !showChat}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="chat-fab"
      onclick={handleOpenChat}
      onkeydown={(e) => e.key === "Enter" && handleOpenChat()}
      role="button"
      tabindex="0"
      aria-label="Open Chat"
    >
      <MessageSquare size="24" />
    </div>
  {/if}

  <!-- 設定モーダル -->
  {#if showSettings}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="settings-overlay"
      onclick={() => (showSettings = false)}
      role="dialog"
      aria-modal="true"
      aria-label="LLM Settings"
    >
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div class="settings-modal" onclick={(e) => e.stopPropagation()} role="document">
        <button
          class="close-btn"
          onclick={() => (showSettings = false)}
          aria-label="Close"
        >
          ×
        </button>
        <LLMSettingsPreview />
      </div>
    </div>
  {/if}
</main>

<style>
  .chat-fab {
    position: fixed;
    bottom: var(--mv-spacing-lg);
    right: var(--mv-spacing-lg);
    width: var(--mv-icon-size-xxxl);
    height: var(--mv-icon-size-xxxl);
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--mv-shadow-card);
    cursor: pointer;
    z-index: 1000;
    font-size: var(--mv-icon-size-md);
  }
  .chat-fab:hover {
    background: var(--mv-color-surface-hover);
  }

  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--mv-color-surface-app);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    overflow: hidden;
  }

  /* Toolbar is overlaid logic not strictly in App.svelte?
     Wait, App.svelte puts Toolbar *above* main-content in flex column.
     So keep Toolbar where it is.
     Correcting template structure to match App.svelte (Toolbar NOT overlay).
  */

  .main-content {
    display: block; /* Flex -> Block */
    position: relative;
    flex: 1;
    overflow: hidden;
    background: var(--mv-color-surface-base);
  }

  .canvas-layer {
    position: absolute;
    inset: 0;
    z-index: 1;
  }

  .list-overlay {
    position: absolute;
    inset: var(--mv-spacing-md);
    z-index: 10;
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-lg);
    box-shadow: var(--mv-shadow-modal);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  /* モーダルは削除済み */

  /* Backlog Styles */
  .backlog-fab {
    position: fixed;
    bottom: var(--mv-spacing-lg);
    left: var(--mv-spacing-lg);
    width: var(--mv-icon-size-xxxl);
    height: var(--mv-icon-size-xxxl);
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-full);
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: var(--mv-shadow-card);
    cursor: pointer;
    z-index: 1000;
    font-size: var(--mv-icon-size-md);
    transition: all var(--mv-transition-hover);
  }

  .backlog-fab:hover {
    background: var(--mv-color-surface-hover);
  }

  .backlog-fab.has-items {
    background: var(--mv-color-status-failed-bg);
    border-color: var(--mv-color-status-failed-text);
  }

  .backlog-count {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-status-failed-text);
  }

  .backlog-sidebar {
    position: fixed;
    top: var(--mv-backlog-sidebar-top);
    left: 0;
    bottom: 0;
    width: var(--mv-backlog-sidebar-width);
    z-index: 100;
    box-shadow: var(--mv-shadow-modal);
  }

  /* Settings Modal */
  .settings-overlay {
    position: fixed;
    inset: 0;
    background: var(--mv-glass-bg-overlay);
    backdrop-filter: var(--mv-glass-blur);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .settings-modal {
    position: relative;
    max-width: var(--mv-content-max-width-sm);
    max-height: var(--mv-container-max-height-modal);
    overflow-y: auto;
  }

  .close-btn {
    position: absolute;
    top: var(--mv-spacing-sm);
    right: var(--mv-spacing-sm);
    width: var(--mv-size-action-btn);
    height: var(--mv-size-action-btn);
    border: none;
    border-radius: var(--mv-radius-full);
    background: transparent;
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-xl);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: var(--mv-transition-base);
    z-index: 1;
  }

  .close-btn:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }
</style>
