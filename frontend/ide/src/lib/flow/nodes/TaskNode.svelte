<!--
  TaskNode - Svelte Flow 用の TaskNode コンポーネント
  
  TaskNodePreview.svelte と同じ Crystal HUD スタイルを使用。
  Handle コンポーネントを含み、SvelteFlow コンテキストで動作。
-->
<script lang="ts">
  import { Handle, Position, type NodeProps } from "@xyflow/svelte";
  import type { Task, TaskStatus, PhaseName } from "../../../types";
  import { selectedTaskId } from "../../../stores/taskStore";

  interface Props extends NodeProps {
    data: {
      task: Task;
    };
  }

  let { data, selected }: Props = $props();

  let task = $derived(data.task);

  // ステータスラベル
  const statusLabels: Record<TaskStatus, string> = {
    PENDING: "待機中",
    READY: "準備完了",
    RUNNING: "実行中",
    SUCCEEDED: "成功",
    COMPLETED: "完了",
    FAILED: "失敗",
    CANCELED: "キャンセル",
    BLOCKED: "ブロック",
    RETRY_WAIT: "リトライ待機",
  };

  const phaseLabels: Record<PhaseName, string> = {
    "": "",
    概念設計: "CONCEPT",
    実装設計: "DESIGN",
    実装: "IMPL",
    検証: "VERIFY",
  };

  // CSS クラス用の小文字変換
  let statusClass = $derived(task.status.toLowerCase());
  let hasDependencies = $derived(
    task.dependencies && task.dependencies.length > 0
  );

  // Zoom-based visibility (always show for now, can be enhanced later)
  let showTitle = true;
  let showDetails = true;

  // 3段階のノードサイズ（タイトル長に応じて）
  function getSizeClass(text: string): string {
    const len = text.length;
    if (len <= 15) return "size-small";
    if (len <= 30) return "size-medium";
    return "size-large";
  }

  let sizeClass = $derived(getSizeClass(task.title));

  function stripMarkdown(text: string): string {
    return text.replace(/[*_~`]/g, "");
  }

  function handleClick() {
    selectedTaskId.select(task.id);
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      handleClick();
    }
  }
</script>

<div
  class="node status-{statusClass} {sizeClass}"
  class:selected
  onclick={handleClick}
  onkeydown={handleKeydown}
  role="button"
  tabindex="0"
  aria-label="{task.title} - {statusLabels[task.status]}"
>
  <!-- Handle Positions (invisible) -->
  <Handle type="target" position={Position.Left} class="flow-handle" />
  <Handle type="source" position={Position.Right} class="flow-handle" />

  <!-- ステータスインジケーター -->
  <div class="status-indicator">
    <span class="status-dot"></span>
    <span class="status-text">{statusLabels[task.status]}</span>
  </div>

  <!-- タイトル -->
  {#if showTitle}
    <div class="title" title={task.title}>
      {stripMarkdown(task.title)}
    </div>
  {/if}

  <!-- 詳細情報 -->
  {#if showDetails}
    <div class="details">
      <span class="pool">{task.poolId}</span>
      {#if hasDependencies}
        <span class="deps">↳ {task.dependencies?.length || 0}</span>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* Handle Styles (Invisible) */
  :global(.flow-handle) {
    background: transparent !important;
    border: none !important;
    width: 1px !important;
    height: 1px !important;
    min-width: 0 !important;
    min-height: 0 !important;
  }

  /* TaskNodePreview.svelte と同一の Crystal HUD スタイル */
  .node {
    position: relative;
    width: var(--mv-grid-cell-width);
    height: auto;
    min-height: 80px;
    border-radius: var(--mv-radius-lg);
    cursor: pointer;
    transition:
      transform var(--mv-duration-fast) var(--mv-easing-out),
      box-shadow var(--mv-duration-fast) var(--mv-easing-out),
      border-color var(--mv-duration-fast),
      width var(--mv-duration-fast) var(--mv-easing-out);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
    overflow: hidden;
    user-select: none;
    box-sizing: border-box;

    /* Crystal HUD Glass Style */
    background: var(--mv-glass-bg-chat);

    /* Multi-layer border */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-top-color: var(--mv-glass-border-light);
    border-bottom-color: var(--mv-glass-border-bottom);

    /* Refined shadow */
    box-shadow: var(--mv-shadow-glass-panel-full);

    padding: var(--mv-spacing-sm);
  }

  /* 3段階ノードサイズ */
  .node.size-small {
    width: 180px;
  }

  .node.size-medium {
    width: 240px;
  }

  .node.size-large {
    width: 320px;
  }

  .node:hover {
    transform: translateY(-3px) scale(1.02);
    background: var(--mv-glass-hover);
    border-color: var(--mv-glass-border-hover);
    box-shadow: var(--mv-shadow-glass-panel-with-glow);
  }

  .node:focus {
    outline: none;
    border-color: var(--mv-shadow-glow-accent-border);
    box-shadow: var(--mv-shadow-floating-with-accent);
  }

  .node.selected {
    border-color: var(--mv-shadow-glow-accent-border);
    box-shadow: var(--mv-shadow-floating-with-accent-inset);
    background: var(--mv-glow-frost-2-lighter);
  }

  /* ステータス別スタイル */
  .node.status-pending {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-pending-text);
  }

  .node.status-ready {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-ready-text);
    box-shadow: var(--mv-shadow-glass-panel-with-frost);
  }

  .node.status-running {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-running-text);
    box-shadow: var(--mv-shadow-glass-panel-with-running);
    animation: mv-pulse var(--mv-duration-pulse) infinite;
  }

  .node.status-succeeded,
  .node.status-completed {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-succeeded-text);
    box-shadow: var(--mv-shadow-glass-panel-with-frost);
  }

  .node.status-failed {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-failed-text);
    box-shadow: var(--mv-shadow-glass-panel-with-failed);
  }

  .node.status-canceled {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-canceled-text);
  }

  .node.status-blocked {
    border-left: var(--mv-border-width-default) solid
      var(--mv-color-status-blocked-text);
    box-shadow: var(--mv-shadow-glass-panel-with-blocked);
  }

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
    transition: box-shadow var(--mv-duration-fast);
  }

  .status-pending .status-dot {
    background: var(--mv-color-status-pending-text);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-pending-text);
  }
  .status-ready .status-dot {
    background: var(--mv-color-status-ready-text);
    box-shadow: var(--mv-shadow-badge-glow-md) var(--mv-color-status-ready-text);
  }
  .status-running .status-dot {
    background: var(--mv-color-status-running-text);
    box-shadow: var(--mv-shadow-badge-glow-lg)
      var(--mv-color-status-running-text);
    animation: dot-pulse 1.5s infinite ease-in-out;
  }
  .status-succeeded .status-dot,
  .status-completed .status-dot {
    background: var(--mv-color-status-succeeded-text);
    box-shadow: var(--mv-shadow-badge-glow-md)
      var(--mv-color-status-succeeded-text);
  }
  .status-failed .status-dot {
    background: var(--mv-color-status-failed-text);
    box-shadow: var(--mv-shadow-badge-glow-md)
      var(--mv-color-status-failed-text);
  }
  .status-canceled .status-dot {
    background: var(--mv-color-status-canceled-text);
  }
  .status-blocked .status-dot {
    background: var(--mv-color-status-blocked-text);
    box-shadow: var(--mv-shadow-badge-glow-sm)
      var(--mv-color-status-blocked-text);
  }

  @keyframes dot-pulse {
    0%,
    100% {
      box-shadow: var(--mv-shadow-badge-glow-sm)
        var(--mv-color-status-running-text);
    }
    50% {
      box-shadow: var(--mv-shadow-glow-frost-2-md)
        var(--mv-color-status-running-text);
    }
  }

  .status-text {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
    transition: all var(--mv-duration-fast);
  }

  .status-pending .status-text {
    color: var(--mv-color-status-pending-text);
    text-shadow: var(--mv-text-shadow-orange);
  }
  .status-ready .status-text {
    color: var(--mv-color-status-ready-text);
    text-shadow: var(--mv-text-shadow-cyan-content);
  }
  .status-running .status-text {
    color: var(--mv-color-status-running-text);
    text-shadow: var(--mv-text-shadow-green);
  }
  .status-succeeded .status-text,
  .status-completed .status-text {
    color: var(--mv-color-status-succeeded-text);
    text-shadow: var(--mv-text-shadow-cyan-content);
  }
  .status-failed .status-text {
    color: var(--mv-color-status-failed-text);
    text-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-glow-failed);
  }
  .status-canceled .status-text {
    color: var(--mv-color-status-canceled-text);
  }
  .status-blocked .status-text {
    color: var(--mv-color-status-blocked-text);
    text-shadow: var(--mv-text-shadow-purple-content);
  }

  /* タイトル */
  .title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-normal);
    flex: 1;
    text-shadow: var(--mv-text-shadow-base-white);

    /* 3行まで表示してclamp */
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .node:hover .title {
    color: var(--mv-primitive-snow-storm-2);
    text-shadow: var(--mv-text-shadow-hover-white);
  }

  /* 詳細情報 */
  .details {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-top: auto;
    padding-top: var(--mv-spacing-xxs);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
  }

  .pool {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-secondary);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    transition: all var(--mv-duration-fast);
  }

  .node:hover .pool {
    color: var(--mv-primitive-frost-2);
    border-color: var(--mv-glow-frost-2);
  }

  .deps {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-muted);
  }
</style>
