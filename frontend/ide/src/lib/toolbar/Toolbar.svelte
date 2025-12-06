<script lang="ts">
  import {
    taskCountsByStatus,
    poolSummaries,
    viewMode,
    overallProgress,
  } from "../../stores";
  import type { TaskStatus } from "../../types";
  import BrandText from "../components/brand/BrandText.svelte";
  import ProgressBar from "../wbs/ProgressBar.svelte";
  import ExecutionControls from "./ExecutionControls.svelte";

  // ステータスサマリの表示設定（フォールバック用）
  const statusDisplay: {
    key: TaskStatus;
    label: string;
    showCount: boolean;
  }[] = [
    { key: "RUNNING", label: "実行中", showCount: true },
    { key: "PENDING", label: "待機", showCount: true },
    { key: "FAILED", label: "失敗", showCount: true },
  ];

  // Pool別サマリがある場合はそれを表示、なければステータス別サマリを表示
  $: hasPoolSummaries = $poolSummaries.length > 0;
  $: isGraphMode = $viewMode === "graph";
</script>

<header class="toolbar">
  <!-- 左側：ブランド -->
  <div class="toolbar-left">
    <BrandText size="sm" />
  </div>

  <!-- 中央：Pool別サマリ or ステータスサマリ -->
  <div class="toolbar-center">
    {#if hasPoolSummaries}
      <!-- Pool別サマリ -->
      <div class="pool-summary">
        {#each $poolSummaries as pool (pool.poolId)}
          <div class="pool-badge">
            <span class="pool-name">{pool.poolId}</span>
            <span class="pool-separator">:</span>
            {#if pool.running > 0}
              <span class="pool-stat running">{pool.running} 実行中</span>
            {/if}
            {#if pool.queued > 0}
              <span class="pool-stat queued">{pool.queued} 待機</span>
            {/if}
            {#if pool.failed > 0}
              <span class="pool-stat failed">{pool.failed} 失敗</span>
            {/if}
            {#if pool.running === 0 && pool.queued === 0 && pool.failed === 0}
              <span class="pool-stat idle">{pool.total} タスク</span>
            {/if}
          </div>
        {/each}
      </div>
    {:else}
      <!-- フォールバック: ステータス別サマリ -->
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
    {/if}
  </div>

  <!-- 右側：進捗・ビュー切替 -->
  <!-- 右側：進捗・ビュー切替 -->
  <div class="toolbar-right">
    <!-- 実行コントロール -->
    <ExecutionControls />

    <!-- 進捗率バー -->
    <div class="progress-section">
      <ProgressBar percentage={$overallProgress.percentage} size="mini" />
      <span class="progress-text">{$overallProgress.percentage}%</span>
    </div>

    <!-- ビュー切替 -->
    <div class="view-toggle">
      <button
        class="view-btn"
        class:active={isGraphMode}
        on:click={() => viewMode.setGraph()}
        aria-label="グラフビュー"
        title="グラフビュー"
      >
        <span class="view-icon">◇</span>
        Graph
      </button>
      <button
        class="view-btn"
        class:active={!isGraphMode}
        on:click={() => viewMode.setWBS()}
        aria-label="WBSビュー"
        title="WBSビュー"
      >
        <span class="view-icon">≡</span>
        WBS
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
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-color-border-subtle);
    flex-shrink: 0;
  }

  .toolbar-left,
  .toolbar-center,
  .toolbar-right {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
  }

  .toolbar-left {
    flex: 1;
    justify-content: flex-start;
  }

  .toolbar-center {
    flex: 2;
    justify-content: center;
  }

  .toolbar-right {
    flex: 1;
    justify-content: flex-end;
    gap: var(--mv-spacing-md);
  }

  .app-title {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    margin: 0;
  }

  .btn-icon {
    font-size: var(--mv-font-size-lg);
    line-height: 1;
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

  /* Pool別サマリ */
  .pool-summary {
    display: flex;
    gap: var(--mv-spacing-md);
  }

  .pool-badge {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-glow-ambient); /* ボーダーをグロー色に */
    border-radius: var(--mv-radius-sm);
    font-size: var(--mv-font-size-xs);
    box-shadow: var(--mv-shadow-node-glow); /* 常時微発光 */
  }

  .pool-name {
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    font-family: var(--mv-font-mono);
  }

  .pool-separator {
    color: var(--mv-color-text-muted);
  }

  .pool-stat {
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    font-weight: var(--mv-font-weight-medium);
  }

  .pool-stat.running {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
  }

  .pool-stat.queued {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
  }

  .pool-stat.failed {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .pool-stat.idle {
    color: var(--mv-color-text-muted);
  }

  /* 進捗バー（ミニ） - styles handled by ProgressBar component */
  .progress-section {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }

  .progress-text {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-muted);
    min-width: var(--mv-progress-text-width-sm);
    text-align: right;
  }

  /* ビュー切り替え */
  .view-toggle {
    display: flex;
    background: var(--mv-color-surface-secondary);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-sm);
    overflow: hidden;
  }

  .view-btn {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-sm);
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-muted);
    background: transparent;
    border: none;
    cursor: pointer;
    transition:
      background-color var(--mv-transition-hover),
      color var(--mv-transition-hover);
  }

  .view-btn:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-color-surface-hover);
  }

  .view-btn.active {
    color: var(--mv-color-text-primary);
    background: var(--mv-color-surface-primary);
  }

  .view-icon {
    font-size: var(--mv-font-size-sm);
  }
</style>
