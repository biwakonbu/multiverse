<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import WorkspaceSelector from "./lib/WorkspaceSelector.svelte";
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
  import { initExecutionEvents } from "./stores/executionStore";

  const log = Logger.withComponent("App");

  let workspaceId: string | null = null;
  let interval: ReturnType<typeof setInterval> | null = null;

  // Chat State
  let isChatVisible = true;
  let chatPosition = { x: 0, y: 0 };

  onMount(() => {
    // Calculate initial position (Bottom-Right)
    // 600px width, 350px height, 20px padding
    const width = 600;
    const height = 350;
    const padding = 20;
    chatPosition = {
      x: window.innerWidth - width - padding,
      y: window.innerHeight - height - padding,
    };
    initExecutionEvents();
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
    // 2秒間隔でポーリング
    interval = setInterval(loadData, 2000);
    log.info("polling started", { interval: 2000 });
  }

  onDestroy(() => {
    if (interval) {
      log.info("polling stopped");
      clearInterval(interval);
    }
  });
</script>

<main class="app">
  {#if !workspaceId}
    <WorkspaceSelector on:selected={onWorkspaceSelected} />
  {:else}
    <!-- ツールバー -->
    <Toolbar />

    <!-- メインコンテンツ -->
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

      <!-- 詳細パネルはフローティングまたはオーバーレイとして扱う (一旦非表示/必要に応じて表示実装) -->
      <!-- <DetailPanel /> -->
    </div>

    <!-- チャットウィンドウ -->
    {#if isChatVisible}
      <FloatingChatWindow
        initialPosition={chatPosition}
        on:close={() => (isChatVisible = false)}
      />
    {/if}

    <!-- チャット再表示ボタン (簡易FAB) -->
    {#if !isChatVisible}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <div
        class="chat-fab"
        on:click={() => (isChatVisible = true)}
        on:keydown={(e) => e.key === "Enter" && (isChatVisible = true)}
        role="button"
        tabindex="0"
        aria-label="Open Chat"
      >
        💬
      </div>
    {/if}
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
</style>
