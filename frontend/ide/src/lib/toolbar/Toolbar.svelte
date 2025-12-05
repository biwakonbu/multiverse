<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { viewport, zoomPercent, taskCountsByStatus } from '../../stores';
  import type { TaskStatus } from '../../types';

  const dispatch = createEventDispatcher<{
    createTask: void;
  }>();

  // ステータスサマリの表示設定
  const statusDisplay: { key: TaskStatus; label: string; showCount: boolean }[] = [
    { key: 'RUNNING', label: '実行中', showCount: true },
    { key: 'PENDING', label: '待機', showCount: true },
    { key: 'FAILED', label: '失敗', showCount: true },
  ];

  function handleCreateTask() {
    dispatch('createTask');
  }
</script>

<header class="toolbar">
  <!-- 左側：タイトルと操作 -->
  <div class="toolbar-left">
    <h1 class="app-title">multiverse IDE</h1>

    <button class="btn btn-primary" on:click={handleCreateTask}>
      <span class="btn-icon">+</span>
      新規タスク
    </button>
  </div>

  <!-- 中央：ステータスサマリ -->
  <div class="toolbar-center">
    <div class="status-summary">
      {#each statusDisplay as { key, label, showCount }}
        {#if showCount && $taskCountsByStatus[key] > 0}
          <div class="status-badge status-{key.toLowerCase()}">
            <span class="status-count">{$taskCountsByStatus[key]}</span>
            <span class="status-label">{label}</span>
          </div>
        {/if}
      {/each}
    </div>
  </div>

  <!-- 右側：ズームコントロール -->
  <div class="toolbar-right">
    <div class="zoom-controls">
      <button
        class="btn btn-icon-only"
        on:click={() => viewport.zoomOut()}
        aria-label="ズームアウト"
        title="ズームアウト (-)"
      >
        −
      </button>

      <button
        class="zoom-value"
        on:click={() => viewport.reset()}
        aria-label="ズームリセット"
        title="リセット (0)"
      >
        {$zoomPercent}%
      </button>

      <button
        class="btn btn-icon-only"
        on:click={() => viewport.zoomIn()}
        aria-label="ズームイン"
        title="ズームイン (+)"
      >
        +
      </button>
    </div>
  </div>
</header>

<style>
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: var(--mv-layout-toolbar-height);
    padding: 0 var(--mv-spacing-md);
    background: var(--mv-color-surface-primary);
    border-bottom: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
    flex-shrink: 0;
  }

  .toolbar-left,
  .toolbar-center,
  .toolbar-right {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-md);
  }

  .toolbar-left {
    flex: 1;
  }

  .toolbar-center {
    flex: 2;
    justify-content: center;
  }

  .toolbar-right {
    flex: 1;
    justify-content: flex-end;
  }

  .app-title {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0;
  }

  /* ボタン基本スタイル */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-primary);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    cursor: pointer;
    transition: background var(--mv-transition-hover),
                border-color var(--mv-transition-hover);
  }

  .btn:hover {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-strong);
  }

  .btn:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .btn-primary {
    background: var(--mv-color-status-ready-bg);
    border-color: var(--mv-color-status-ready-border);
    color: var(--mv-color-status-ready-text);
  }

  .btn-primary:hover {
    background: var(--mv-color-status-ready-border);
  }

  .btn-icon {
    font-size: var(--mv-font-size-lg);
    line-height: 1;
  }

  .btn-icon-only {
    width: var(--mv-icon-size-xl);
    height: var(--mv-icon-size-xl);
    padding: 0;
    font-size: var(--mv-font-size-lg);
  }

  /* ステータスサマリ */
  .status-summary {
    display: flex;
    gap: var(--mv-spacing-sm);
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
  }

  .status-badge.status-running {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
  }

  .status-badge.status-pending {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
  }

  .status-badge.status-failed {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .status-count {
    font-weight: var(--mv-font-weight-bold);
  }

  .status-label {
    font-weight: var(--mv-font-weight-normal);
  }

  /* ズームコントロール */
  .zoom-controls {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    padding: var(--mv-spacing-xxs);
  }

  .zoom-controls .btn-icon-only {
    border: none;
    background: transparent;
  }

  .zoom-controls .btn-icon-only:hover {
    background: var(--mv-color-surface-hover);
  }

  .zoom-value {
    min-width: var(--mv-spacing-xxl);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-secondary);
    background: transparent;
    border: none;
    cursor: pointer;
    text-align: center;
  }

  .zoom-value:hover {
    color: var(--mv-color-text-primary);
  }
</style>
