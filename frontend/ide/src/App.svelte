<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import WorkspaceSelector from "./lib/WorkspaceSelector.svelte";
  import TaskCreate from "./lib/TaskCreate.svelte";
  import { GridCanvas } from "./lib/grid";
  import { Toolbar } from "./lib/toolbar";
  import { DetailPanel } from "./lib/panel";
  import { WBSView } from "./lib/wbs";
  import { Button } from "./design-system";
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

  const log = Logger.withComponent("App");

  let workspaceId: string | null = null;
  let showCreateModal = false;
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

  // タスク作成モーダルを開く
  function handleCreateTask() {
    log.debug("opening create task modal");
    showCreateModal = true;
  }

  // タスク作成完了
  function onTaskCreated() {
    log.info("task created, refreshing task list");
    showCreateModal = false;
    loadTasks();
  }

  // タスク作成キャンセル
  function onCreateCancel() {
    log.debug("create task cancelled");
    showCreateModal = false;
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
    <Toolbar on:createTask={handleCreateTask} />

    <!-- メインコンテンツ -->
    <div class="main-content">
      <!-- Graph/WBS ビュー切り替え -->
      {#if $viewMode === 'graph'}
        <GridCanvas />
      {:else}
        <WBSView />
      {/if}

      <!-- 詳細パネル -->
      <DetailPanel />
    </div>

    <!-- タスク作成モーダル -->
    {#if showCreateModal}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <div class="modal-overlay" on:click={onCreateCancel} role="presentation">
        <!-- svelte-ignore a11y-click-events-have-key-events a11y-no-noninteractive-element-interactions -->
        <div
          class="modal-content"
          on:click|stopPropagation
          role="dialog"
          aria-modal="true"
          aria-labelledby="create-task-title"
        >
          <header class="modal-header">
            <h2 id="create-task-title">新規タスク作成</h2>
            <Button
              variant="ghost"
              size="small"
              on:click={onCreateCancel}
              label="×"
            />
          </header>
          <TaskCreate on:created={onTaskCreated} />
        </div>
      </div>
    {/if}

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
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  /* モーダルオーバーレイ */
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: var(--mv-color-surface-overlay);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal-content {
    background: var(--mv-color-surface-primary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-lg);
    width: 100%;
    max-width: var(--mv-container-max-width-sm);
    max-height: var(--mv-container-max-height-modal);
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--mv-spacing-md);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-color-border-subtle);
  }

  .modal-header h2 {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0;
  }
</style>
