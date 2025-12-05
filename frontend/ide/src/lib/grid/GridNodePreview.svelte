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
    FAILED: '失敗',
    CANCELED: 'キャンセル',
    BLOCKED: 'ブロック',
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
    width: var(--mv-grid-cell-width, 160px);
    height: var(--mv-grid-cell-height, 100px);
    background: var(--mv-color-surface-node, #2d2d2d);
    border: 2px solid var(--mv-color-border-default, #444444);
    border-radius: var(--mv-radius-md, 8px);
    padding: var(--mv-spacing-xs, 8px) var(--mv-spacing-sm, 12px);
    cursor: pointer;
    transition: border-color var(--mv-transition-hover, 150ms ease-out),
                box-shadow var(--mv-transition-hover, 150ms ease-out),
                transform var(--mv-transition-hover, 150ms ease-out);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs, 4px);
    overflow: hidden;
    user-select: none;
    box-sizing: border-box;
  }

  .node:hover {
    border-color: var(--mv-color-border-strong, #666666);
    transform: translateY(-2px);
  }

  .node:focus {
    outline: none;
    border-color: var(--mv-color-border-focus, #4caf50);
    box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.3);
  }

  .node.selected {
    border-color: var(--mv-color-border-focus, #4caf50);
    box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.4);
  }

  /* ステータス別スタイル */
  .node.status-pending {
    background: var(--mv-color-status-pending-bg, #3a3a3a);
    border-color: var(--mv-color-status-pending-border, #666666);
  }

  .node.status-ready {
    background: var(--mv-color-status-ready-bg, #2a2a4a);
    border-color: var(--mv-color-status-ready-border, #5588ff);
  }

  .node.status-running {
    background: var(--mv-color-status-running-bg, #2a3a2a);
    border-color: var(--mv-color-status-running-border, #44bb44);
    animation: mv-pulse 2s infinite;
  }

  .node.status-succeeded {
    background: var(--mv-color-status-succeeded-bg, #1a3a1a);
    border-color: var(--mv-color-status-succeeded-border, #228822);
  }

  .node.status-failed {
    background: var(--mv-color-status-failed-bg, #3a1a1a);
    border-color: var(--mv-color-status-failed-border, #cc2222);
  }

  .node.status-canceled {
    background: var(--mv-color-status-canceled-bg, #2a2a2a);
    border-color: var(--mv-color-status-canceled-border, #555555);
  }

  .node.status-blocked {
    background: var(--mv-color-status-blocked-bg, #3a3a1a);
    border-color: var(--mv-color-status-blocked-border, #ccaa22);
  }

  /* パルスアニメーション */
  @keyframes mv-pulse {
    0%, 100% {
      box-shadow: 0 0 0 0 rgba(68, 187, 68, 0.4);
    }
    50% {
      box-shadow: 0 0 8px 4px rgba(68, 187, 68, 0.4);
    }
  }

  /* ステータスインジケーター */
  .status-indicator {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs, 4px);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-pending .status-dot { background: var(--mv-color-status-pending-text, #ff9800); }
  .status-ready .status-dot { background: var(--mv-color-status-ready-text, #5588ff); }
  .status-running .status-dot { background: var(--mv-color-status-running-text, #4caf50); }
  .status-succeeded .status-dot { background: var(--mv-color-status-succeeded-text, #228822); }
  .status-failed .status-dot { background: var(--mv-color-status-failed-text, #f44336); }
  .status-canceled .status-dot { background: var(--mv-color-status-canceled-text, #888888); }
  .status-blocked .status-dot { background: var(--mv-color-status-blocked-text, #ccaa22); }

  .status-text {
    font-size: var(--mv-font-size-xs, 10px);
    font-weight: var(--mv-font-weight-bold, 700);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .status-pending .status-text { color: var(--mv-color-status-pending-text, #ff9800); }
  .status-ready .status-text { color: var(--mv-color-status-ready-text, #5588ff); }
  .status-running .status-text { color: var(--mv-color-status-running-text, #4caf50); }
  .status-succeeded .status-text { color: var(--mv-color-status-succeeded-text, #228822); }
  .status-failed .status-text { color: var(--mv-color-status-failed-text, #f44336); }
  .status-canceled .status-text { color: var(--mv-color-status-canceled-text, #888888); }
  .status-blocked .status-text { color: var(--mv-color-status-blocked-text, #ccaa22); }

  /* タイトル */
  .title {
    font-size: var(--mv-font-size-sm, 12px);
    font-weight: var(--mv-font-weight-semibold, 600);
    color: var(--mv-color-text-primary, #eeeeee);
    line-height: var(--mv-line-height-normal, 1.5);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  /* 詳細情報 */
  .details {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs, 8px);
    margin-top: auto;
  }

  .pool {
    font-size: var(--mv-font-size-xs, 10px);
    font-family: var(--mv-font-mono, monospace);
    color: var(--mv-color-text-secondary, #aaaaaa);
    background: var(--mv-color-surface-secondary, #252525);
    padding: 2px var(--mv-spacing-xxs, 4px);
    border-radius: var(--mv-radius-sm, 4px);
  }
</style>
