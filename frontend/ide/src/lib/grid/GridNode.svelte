<script lang="ts">
  import type { Task } from '../../types';
  import { gridToCanvas } from '../../design-system';
  import { statusToCssClass, statusLabels } from '../../types';
  import { selectedTaskId } from '../../stores';

  // Props
  export let task: Task;
  export let col: number;
  export let row: number;
  export let zoomLevel: number = 1;

  // キャンバス座標を計算
  $: position = gridToCanvas(col, row);
  $: isSelected = $selectedTaskId === task.id;
  $: statusClass = statusToCssClass(task.status);

  // ズームレベルに応じた表示制御
  $: showTitle = zoomLevel >= 0.4;
  $: showDetails = zoomLevel >= 1.2;

  function handleClick() {
    selectedTaskId.select(task.id);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleClick();
    }
  }
</script>

<div
  class="node status-{statusClass}"
  class:selected={isSelected}
  style="left: {position.x}px; top: {position.y}px;"
  on:click={handleClick}
  on:keydown={handleKeydown}
  role="button"
  tabindex="0"
  aria-label="{task.title} - {statusLabels[task.status]}"
>
  <!-- ステータスインジケーター -->
  <div class="status-indicator">
    <span class="status-dot"></span>
    <span class="status-text">{statusLabels[task.status]}</span>
  </div>

  <!-- タイトル（ズームレベルに応じて表示） -->
  {#if showTitle}
    <div class="title" title={task.title}>
      {task.title}
    </div>
  {/if}

  <!-- 詳細情報（高ズームレベルで表示） -->
  {#if showDetails}
    <div class="details">
      <span class="pool">{task.poolId}</span>
    </div>
  {/if}
</div>

<style>
  .node {
    position: absolute;
    width: var(--mv-grid-cell-width);
    height: var(--mv-grid-cell-height);
    background: var(--mv-color-surface-node);
    border: 2px solid var(--mv-color-border-default);
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
  }

  .node:hover {
    border-color: var(--mv-color-border-strong);
    transform: translateY(-2px);
  }

  .node:focus {
    outline: none;
    border-color: var(--mv-color-border-focus);
    box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.3);
  }

  .node.selected {
    border-color: var(--mv-color-border-focus);
    box-shadow: 0 0 0 3px rgba(76, 175, 80, 0.4);
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

  /* ステータスインジケーター */
  .status-indicator {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
  }

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .status-pending .status-dot {
    background: var(--mv-color-status-pending-text);
  }

  .status-ready .status-dot {
    background: var(--mv-color-status-ready-text);
  }

  .status-running .status-dot {
    background: var(--mv-color-status-running-text);
  }

  .status-succeeded .status-dot {
    background: var(--mv-color-status-succeeded-text);
  }

  .status-failed .status-dot {
    background: var(--mv-color-status-failed-text);
  }

  .status-canceled .status-dot {
    background: var(--mv-color-status-canceled-text);
  }

  .status-blocked .status-dot {
    background: var(--mv-color-status-blocked-text);
  }

  .status-text {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .status-pending .status-text {
    color: var(--mv-color-status-pending-text);
  }

  .status-ready .status-text {
    color: var(--mv-color-status-ready-text);
  }

  .status-running .status-text {
    color: var(--mv-color-status-running-text);
  }

  .status-succeeded .status-text {
    color: var(--mv-color-status-succeeded-text);
  }

  .status-failed .status-text {
    color: var(--mv-color-status-failed-text);
  }

  .status-canceled .status-text {
    color: var(--mv-color-status-canceled-text);
  }

  .status-blocked .status-text {
    color: var(--mv-color-status-blocked-text);
  }

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
    padding: 2px var(--mv-spacing-xxs);
    border-radius: var(--mv-radius-sm);
  }
</style>
