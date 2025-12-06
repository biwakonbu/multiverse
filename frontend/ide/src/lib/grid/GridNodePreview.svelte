<!--
  GridNodePreview - Storybook 用の GridNode プレビューコンポーネント

  ストア依存を排除し、props のみで動作するバージョン。
  Storybook での単独表示・テスト用途。
-->
<script lang="ts">
  import type { TaskStatus } from '../../types';
  import { gridToCanvas } from '../../design-system';

  // タスク情報（idはStorybookのargs用に保持）
  // svelte-ignore unused-export-let
  export let id = 'task-1';
  export let title = 'タスク名';
  export let status: TaskStatus = 'PENDING';
  export let poolId = 'codegen';

  // グリッド位置
  export let col = 0;
  export let row = 0;

  // ズームレベル
  export let zoomLevel = 1;

  // 選択状態
  export let selected = false;

  // ステータスラベル
  const statusLabels: Record<TaskStatus, string> = {
    PENDING: '待機中',
    READY: '準備完了',
    RUNNING: '実行中',
    SUCCEEDED: '成功',
    COMPLETED: '完了',
    FAILED: '失敗',
    CANCELED: 'キャンセル',
    BLOCKED: 'ブロック',
    RETRY_WAIT: 'リトライ待機',
  };

  // CSS クラス用の小文字変換
  $: statusClass = status.toLowerCase();

  // キャンバス座標を計算
  $: position = gridToCanvas(col, row);

  // ズームレベルに応じた表示制御
  $: showTitle = zoomLevel >= 0.4;
  $: showDetails = zoomLevel >= 1.2;
</script>

<div
  class="node status-{statusClass}"
  class:selected
  style="left: {position.x}px; top: {position.y}px;"
  role="button"
  tabindex="0"
  aria-label="{title} - {statusLabels[status]}"
>
  <!-- ステータスインジケーター -->
  <div class="status-indicator">
    <span class="status-dot"></span>
    <span class="status-text">{statusLabels[status]}</span>
  </div>

  <!-- タイトル（ズームレベルに応じて表示） -->
  {#if showTitle}
    <div class="title" title={title}>
      {title}
    </div>
  {/if}

  <!-- 詳細情報（高ズームレベルで表示） -->
  {#if showDetails}
    <div class="details">
      <span class="pool">{poolId}</span>
    </div>
  {/if}
</div>

<style>
  .node {
    position: absolute;
    width: var(--mv-grid-cell-width);
    height: var(--mv-grid-cell-height);
    background: var(--mv-color-surface-node);
    border: var(--mv-border-width-default) solid var(--mv-color-border-default);
    border-radius: var(--mv-radius-md);
    padding: var(--mv-spacing-xs) var(--mv-spacing-sm);
    cursor: pointer;
    transition: border-color var(--mv-transition-hover),
                box-shadow var(--mv-transition-hover),
                transform var(--mv-transition-hover);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
    overflow: hidden;
    user-select: none;
    box-sizing: border-box;
  }

  .node:hover {
    border-color: var(--mv-color-border-strong);
    transform: var(--mv-transform-hover-lift);
  }

  .node:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .node.selected {
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-selected);
  }

  /* ステータス別スタイル */
  .node.status-pending {
    background: var(--mv-color-status-pending-bg);
    border-color: var(--mv-color-status-pending-border);
  }

  .node.status-ready {
    background: var(--mv-color-status-ready-bg);
    border-color: var(--mv-color-status-ready-border);
  }

  .node.status-running {
    background: var(--mv-color-status-running-bg);
    border-color: var(--mv-color-status-running-border);
    animation: mv-pulse var(--mv-duration-pulse) infinite;
  }

  .node.status-succeeded {
    background: var(--mv-color-status-succeeded-bg);
    border-color: var(--mv-color-status-succeeded-border);
  }

  .node.status-failed {
    background: var(--mv-color-status-failed-bg);
    border-color: var(--mv-color-status-failed-border);
  }

  .node.status-canceled {
    background: var(--mv-color-status-canceled-bg);
    border-color: var(--mv-color-status-canceled-border);
  }

  .node.status-blocked {
    background: var(--mv-color-status-blocked-bg);
    border-color: var(--mv-color-status-blocked-border);
  }

  /* パルスアニメーション（グローバル定義を使用） */

  /* ステータスインジケーター */
  .status-indicator {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
  }

  .status-dot {
    width: var(--mv-indicator-size-sm);
    height: var(--mv-indicator-size-sm);
    border-radius: var(--mv-radius-full);
    flex-shrink: 0;
  }

  .status-pending .status-dot { background: var(--mv-color-status-pending-text); }
  .status-ready .status-dot { background: var(--mv-color-status-ready-text); }
  .status-running .status-dot { background: var(--mv-color-status-running-text); }
  .status-succeeded .status-dot { background: var(--mv-color-status-succeeded-text); }
  .status-failed .status-dot { background: var(--mv-color-status-failed-text); }
  .status-canceled .status-dot { background: var(--mv-color-status-canceled-text); }
  .status-blocked .status-dot { background: var(--mv-color-status-blocked-text); }

  .status-text {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .status-pending .status-text { color: var(--mv-color-status-pending-text); }
  .status-ready .status-text { color: var(--mv-color-status-ready-text); }
  .status-running .status-text { color: var(--mv-color-status-running-text); }
  .status-succeeded .status-text { color: var(--mv-color-status-succeeded-text); }
  .status-failed .status-text { color: var(--mv-color-status-failed-text); }
  .status-canceled .status-text { color: var(--mv-color-status-canceled-text); }
  .status-blocked .status-text { color: var(--mv-color-status-blocked-text); }

  /* タイトル */
  .title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-normal);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  /* 詳細情報 */
  .details {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-top: auto;
  }

  .pool {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-secondary);
    background: var(--mv-color-surface-secondary);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xxs);
    border-radius: var(--mv-radius-sm);
  }
</style>
