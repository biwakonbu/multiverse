<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import WorkspaceSelector from './lib/WorkspaceSelector.svelte';
  import TaskCreate from './lib/TaskCreate.svelte';
  import { GridCanvas } from './lib/grid';
  import { Toolbar } from './lib/toolbar';
  import { DetailPanel } from './lib/panel';
  import { tasks, selectedTask, selectedTaskId } from './stores';
  import type { Task } from './types';
  // @ts-ignore - Wails自動生成ファイル
  import { ListTasks } from '../wailsjs/go/main/App';

  let workspaceId: string | null = null;
  let showCreateModal = false;
  let interval: ReturnType<typeof setInterval> | null = null;

  // タスク一覧を読み込み
  async function loadTasks() {
    if (!workspaceId) return;
    try {
      const taskList: Task[] = await ListTasks();
      tasks.setTasks(taskList);
    } catch (e) {
      console.error('タスク読み込みエラー:', e);
    }
  }

  // Workspace選択時
  function onWorkspaceSelected(event: CustomEvent<string>) {
    workspaceId = event.detail;
    loadTasks();
    // 2秒間隔でポーリング
    interval = setInterval(loadTasks, 2000);
  }

  // タスク作成モーダルを開く
  function handleCreateTask() {
    showCreateModal = true;
  }

  // タスク作成完了
  function onTaskCreated() {
    showCreateModal = false;
    loadTasks();
  }

  // タスク作成キャンセル
  function onCreateCancel() {
    showCreateModal = false;
  }

  onDestroy(() => {
    if (interval) {
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
      <!-- グリッドキャンバス -->
      <GridCanvas />

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
            <button class="btn-close" on:click={onCreateCancel} aria-label="閉じる">
              ×
            </button>
          </header>
          <TaskCreate on:created={onTaskCreated} />
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
    border-bottom: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
  }

  .modal-header h2 {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0;
  }

  .btn-close {
    width: var(--mv-icon-size-xl);
    height: var(--mv-icon-size-xl);
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    border-radius: var(--mv-radius-sm);
    color: var(--mv-color-text-secondary);
    font-size: var(--mv-font-size-xl);
    cursor: pointer;
    transition: background var(--mv-transition-hover),
                color var(--mv-transition-hover);
  }

  .btn-close:hover {
    background: var(--mv-color-surface-hover);
    color: var(--mv-color-text-primary);
  }
</style>
