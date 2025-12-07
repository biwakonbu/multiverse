<script lang="ts">
  import { createBubbler, stopPropagation } from "svelte/legacy";

  const bubble = createBubbler();
  import { onMount, onDestroy } from "svelte";
  import WorkspaceSelector from "./lib/WorkspaceSelector.svelte";
  import TitleBar from "./lib/TitleBar.svelte";
  import { Toolbar } from "./lib/toolbar";
  import { WBSListView, WBSGraphView } from "./lib/wbs";
  import {
    tasks,
    selectedTask,
    selectedTaskId,
    poolSummaries,
    viewMode,
  } from "./stores";
  import { Logger } from "./services/logger";
  import type { Task, PoolSummary } from "./types";
  // @ts-ignore - Wails自動生成ファイル
  import { ListTasks, GetPoolSummaries } from "../wailsjs/go/main/App";
  import FloatingChatWindow from "./lib/components/chat/FloatingChatWindow.svelte";
  // import ProcessHUD from "./lib/hud/ProcessHUD.svelte"; // Removed
  import { initLogEvents, logs } from "./stores/logStore";
  import { executionState, initExecutionEvents } from "./stores/executionStore";
  import { initProcessEvents, processResources } from "./stores/processStore";
  import { initTaskEvents } from "./stores/taskStore";
  import { initChatEvents } from "./stores/chat";
  import { initBacklogEvents, unresolvedCount } from "./stores/backlogStore";
  import BacklogPanel from "./lib/backlog/BacklogPanel.svelte";
  import LLMSettings from "./lib/settings/LLMSettings.svelte";
  import TaskBar from "./lib/hud/TaskBar.svelte";
  import ProcessWindow from "./lib/hud/ProcessWindow.svelte";
  import { windowStore } from "./stores/windowStore";

  const log = Logger.withComponent("App");

  let workspaceId: string | null = $state(null);
  let interval: ReturnType<typeof setInterval> | null = null;

  // 実行中のタスクを取得するリアクティブ変数
  let runningTask = $derived($tasks.find((t) => t.status === "RUNNING"));

  // Chat State (Managed by windowStore now)
  // let isChatVisible = $state(true);
  // let chatPosition = $state({ x: 0, y: 0 });

  // Backlog State (Managed by windowStore now)
  // let isBacklogVisible = $state(false);

  // Settings State
  let isSettingsVisible = $state(false);

  onMount(() => {
    // Window positioning is now handled by windowStore defaults
    /*
    const width = 600;
    const height = 350;
    const padding = 20;
    chatPosition = {
      x: window.innerWidth - width - padding,
      y: window.innerHeight - height - padding,
    };
    */
    // Wails Events 初期化
    initExecutionEvents();
    initTaskEvents();
    initChatEvents();
    initBacklogEvents();
    initLogEvents();
    initProcessEvents();
  });

  // タスク一覧を読み込み
  async function loadTasks() {
    if (!workspaceId) return;
    try {
      const result = await ListTasks();
      // Wails生成型からローカル型へ変換
      const taskList: Task[] = (result || []).map(
        (t): Task => ({
          id: t.id,
          title: t.title,
          status: t.status as Task["status"],
          poolId: t.poolId,
          createdAt: t.createdAt,
          updatedAt: t.updatedAt,
          startedAt: t.startedAt,
          doneAt: t.doneAt,
          description: t.description,
          dependencies: t.dependencies,
          parentId: t.parentId,
          wbsLevel: t.wbsLevel,
          phaseName: t.phaseName as Task["phaseName"],
          milestone: t.milestone,
          sourceChatId: t.sourceChatId,
          acceptanceCriteria: t.acceptanceCriteria,
        })
      );
      log.debug("tasks loaded", { count: taskList.length });
      tasks.setTasks(taskList);
    } catch (e) {
      log.error("failed to load tasks", { error: e });
    }
  }

  // Pool別サマリを読み込み
  async function loadPoolSummaries() {
    if (!workspaceId) return;
    try {
      const summaries: PoolSummary[] = await GetPoolSummaries();
      log.debug("pool summaries loaded", { count: summaries?.length ?? 0 });
      poolSummaries.setSummaries(summaries || []);
    } catch (e) {
      log.error("failed to load pool summaries", { error: e });
    }
  }

  // データ読み込み（タスク + Poolサマリ）
  async function loadData() {
    await Promise.all([loadTasks(), loadPoolSummaries()]);
  }

  // Workspace選択時
  function onWorkspaceSelected(event: CustomEvent<string>) {
    workspaceId = event.detail;
    log.info("workspace selected", { workspaceId });
    loadData();
    // 10秒間隔でポーリング（Wails Events でリアルタイム更新されるためフォールバック）
    interval = setInterval(loadData, 10000);
    log.info("polling started", { interval: 10000 });
  }

  onDestroy(() => {
    if (interval) {
      log.info("polling stopped");
      clearInterval(interval);
    }
  });
</script>

<main class="app">
  <TitleBar />
  {#if !workspaceId}
    <WorkspaceSelector on:selected={onWorkspaceSelected} />
  {:else}
    <!-- ツールバー -->
    <Toolbar on:showSettings={() => (isSettingsVisible = true)} />

    <!-- メインコンテンツ -->
    <div class="main-content">
      <!-- 常にGraphViewを描画し、canvasとして機能させる -->
      <div
        class="canvas-layer"
        style:visibility={$viewMode === "graph" ? "visible" : "hidden"}
      >
        <WBSGraphView />
      </div>

      <!-- WBSモード時はオーバーレイとして表示（あるいはcanvas上に配置） -->
      {#if $viewMode === "wbs"}
        <div class="list-overlay">
          <WBSListView />
        </div>
      {/if}
    </div>

    <!-- Window System -->
    <FloatingChatWindow />
    <ProcessWindow resources={$processResources} />

    <!-- TaskBar (Dock) -->
    <TaskBar />

    <!-- バックログ表示ボタン (TaskBarに統合) -->

    <!-- バックログパネル (TODO: Window化するかサイドバーのままにするか。一旦サイドバーのままTaskBarでtoggle) -->
    <!-- BacklogPanelはwindowStore.backlog.isOpenで制御する -->
    {#if $windowStore.backlog.isOpen}
      <div class="backlog-sidebar">
        <BacklogPanel />
      </div>
    {/if}

    <!-- 設定モーダル -->
    {#if isSettingsVisible}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <div
        class="settings-overlay"
        onclick={() => (isSettingsVisible = false)}
        role="dialog"
        aria-modal="true"
        aria-label="LLM Settings"
      >
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <div
          class="settings-modal"
          onclick={stopPropagation(bubble("click"))}
          role="document"
        >
          <button
            class="close-btn"
            onclick={() => (isSettingsVisible = false)}
            aria-label="Close"
          >
            ×
          </button>
          <LLMSettings />
        </div>
      </div>
    {/if}
  {/if}
</main>

<style>
  .app {
    height: 100vh;
    display: flex;
    flex-direction: column;
    background: var(--mv-color-surface-app);
    padding-top: var(--mv-titlebar-height);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-sans);
    overflow: hidden;
  }

  .main-content {
    display: block; /* フレックスからブロックへ変更 (絶対配置のコンテナにするため) */
    position: relative;
    flex: 1;
    overflow: hidden;
    background: var(--mv-color-surface-base); /* Canvasの背景色 */
  }

  .canvas-layer {
    position: absolute;
    inset: 0;
    z-index: 1;
  }

  .list-overlay {
    position: absolute;
    inset: var(--mv-spacing-md); /* 少し余白を持たせてフローティング感を出す */
    z-index: 10;
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-lg);
    box-shadow: var(--mv-shadow-modal);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  /* タスク作成モーダルは削除済み */

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
    max-height: var(--mv-settings-modal-max-height, 80vh);
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
  }

  .close-btn:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }
</style>
