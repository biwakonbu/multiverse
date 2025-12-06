<script lang="ts">
  import type { WBSNode } from "../../stores/wbsStore";
  import { expandedNodes, viewMode } from "../../stores/wbsStore";
  import { selectedTaskId } from "../../stores";
  import { phaseToCssClass, type PhaseName } from "../../schemas";
  import StatusBadge from "./StatusBadge.svelte"; // Import new component
  import ProgressBar from "./ProgressBar.svelte"; // Import ProgressBar
  import { getProgressColor } from "./utils"; // Import color logic
  import { onDestroy } from "svelte";

  // Props
  export let node: WBSNode;
  export let expanded: boolean = true;

  $: isPhase = node.type === "phase";
  $: isTask = node.type === "task";
  $: hasChildren = node.children.length > 0;
  $: phaseClass = phaseToCssClass(node.phaseName);
  // statusClass removed as it's handled by StatusBadge
  $: isSelected = node.task && $selectedTaskId === node.task.id;
  $: progressPercent = node.progress.percentage;
  $: progressColor = getProgressColor(progressPercent); // Get dynamic color
  $: indentStyle = `padding-left: ${node.level * 24 + 12}px`;

  // Retry Logic
  $: isRetryWait = node.task?.status === "RETRY_WAIT";
  $: attemptCount = node.task?.attemptCount || 0;
  $: nextRetryAt = node.task?.nextRetryAt;

  let timeRemaining = "";
  let interval: any;

  function updateTimeRemaining() {
    if (!nextRetryAt) {
      timeRemaining = "";
      return;
    }
    const now = new Date();
    const target = new Date(nextRetryAt);
    const diff = target.getTime() - now.getTime();

    if (diff <= 0) {
      timeRemaining = ""; // 0秒以下は表示しない（すぐにリセットされるはず）
    } else {
      const seconds = Math.ceil(diff / 1000);
      timeRemaining = `${seconds}s`;
    }
  }

  $: {
    if (isRetryWait && nextRetryAt) {
      updateTimeRemaining();
      if (!interval) {
        interval = setInterval(updateTimeRemaining, 1000);
      }
    } else {
      if (interval) {
        clearInterval(interval);
        interval = undefined;
      }
      timeRemaining = "";
    }
  }

  onDestroy(() => {
    if (interval) clearInterval(interval);
  });

  function handleToggle(event: MouseEvent) {
    event.stopPropagation();
    if (hasChildren) {
      expandedNodes.toggle(node.id);
    }
  }

  function handleClick() {
    if (isTask && node.task) {
      selectedTaskId.select(node.task.id);
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      if (isPhase && hasChildren) {
        expandedNodes.toggle(node.id);
      } else {
        handleClick();
      }
    }
  }
</script>

<!-- svelte-ignore a11y-no-noninteractive-tabindex -->
<div
  class="wbs-node"
  class:is-phase={isPhase}
  class:is-task={isTask}
  class:selected={isSelected}
  class:phase-concept={phaseClass === "phase-concept"}
  class:phase-design={phaseClass === "phase-design"}
  class:phase-impl={phaseClass === "phase-impl"}
  class:phase-verify={phaseClass === "phase-verify"}
  style={indentStyle}
  role={isPhase ? "treeitem" : "button"}
  tabindex="0"
  aria-expanded={isPhase && hasChildren ? expanded : undefined}
  aria-label={node.label}
  on:click={handleClick}
  on:keydown={handleKeydown}
>
  <!-- 展開/折りたたみトグル -->
  {#if isPhase && hasChildren}
    <button
      class="toggle-btn"
      on:click={handleToggle}
      aria-label={expanded ? "折りたたむ" : "展開する"}
    >
      <span class="toggle-icon" class:expanded>{expanded ? "▼" : "▶"}</span>
    </button>
  {:else if isTask}
    <span class="task-bullet">•</span>
  {:else}
    <span class="spacer"></span>
  {/if}

  <!-- フェーズバー（カラーインジケーター） -->
  {#if phaseClass}
    <span class="phase-indicator" aria-hidden="true"></span>
  {/if}

  <!-- ラベル -->
  <span class="node-label">
    {node.label}
  </span>

  <!-- ステータスバッジ（タスクのみ） -->
  {#if isTask && node.task}
    <StatusBadge status={node.task.status} />
    {#if isRetryWait}
      <span class="retry-info">
        Try {attemptCount} • {timeRemaining}
      </span>
    {/if}
  {/if}

  <!-- 進捗バー -->
  {#if isPhase}
    <div class="progress-container">
      <ProgressBar percentage={progressPercent} size="sm" />
      <span
        class="progress-text"
        style:color={progressColor.fill}
        style:text-shadow={progressColor.textShadowXs}
      >
        {node.progress.completed}/{node.progress.total}
        ({progressPercent}%)
      </span>
    </div>
  {/if}
</div>

<style>
  .wbs-node {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    height: var(--mv-wbs-node-height);
    padding: 0 var(--mv-spacing-xs); /* Added padding */
    cursor: pointer;
    border-radius: var(--mv-radius-sm);
    border: var(--mv-border-width-thin) solid transparent; /* Prepare for border */
    transition:
      background-color var(--mv-transition-hover),
      border-color var(--mv-transition-hover),
      box-shadow var(--mv-transition-hover),
      transform var(--mv-transition-hover);
    user-select: none;
    position: relative; /* For absolute positioning if needed */
  }

  .wbs-node:hover {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-subtle);
    box-shadow: var(--mv-shadow-glow-sm);
  }

  .wbs-node:focus {
    outline: none;
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  .wbs-node.selected {
    background: var(--mv-color-surface-selected);
    border-color: var(--mv-color-border-focus);
    box-shadow: var(--mv-shadow-focus);
  }

  /* フェーズノード */
  .wbs-node.is-phase {
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    background: var(
      --mv-color-surface-secondary
    ); /* Slight background for phases */
    border-color: var(--mv-color-border-subtle);
  }

  .wbs-node.is-phase:hover {
    background: var(--mv-color-surface-hover);
    border-color: var(--mv-color-border-default);
    box-shadow: var(--mv-shadow-card);
  }

  .wbs-node.is-phase .node-label {
    font-size: var(--mv-font-size-base);
    letter-spacing: var(--mv-letter-spacing-widest);
  }

  /* タスクノード */
  .wbs-node.is-task {
    font-weight: var(--mv-font-weight-normal);
    color: var(--mv-color-text-secondary);
  }

  .wbs-node.is-task:hover {
    color: var(--mv-color-text-primary);
    transform: translateX(2px); /* Micro-interaction */
  }

  /* トグルボタン */
  .toggle-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-wbs-toggle-size);
    height: var(--mv-wbs-toggle-size);
    padding: 0;
    border: none;
    background: transparent;
    cursor: pointer;
    color: var(--mv-color-text-muted);
    border-radius: var(--mv-radius-sm);
    transition:
      background-color var(--mv-transition-hover),
      color var(--mv-transition-hover),
      transform var(--mv-transition-hover);
  }

  .toggle-btn:hover {
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
    transform: scale(1.1);
  }

  .toggle-icon {
    font-size: var(--mv-wbs-toggle-icon-size);
    transition: transform var(--mv-transition-hover);
  }

  .task-bullet {
    width: var(--mv-wbs-toggle-size);
    text-align: center;
    color: var(--mv-color-border-strong); /* Subtler bullet */
    font-size: var(--mv-font-size-xs);
    opacity: 0.5;
  }

  .spacer {
    width: var(--mv-wbs-toggle-size);
  }

  /* フェーズインジケーター */
  .phase-indicator {
    width: var(--mv-wbs-phase-bar-width);
    height: var(--mv-wbs-toggle-size);
    border-radius: var(--mv-radius-pill);
    flex-shrink: 0;
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--phase-color);
    opacity: 0.8;
  }

  .phase-concept .phase-indicator {
    background: var(--mv-primitive-frost-3);
    box-shadow: var(--mv-shadow-phase-glow-sm) var(--mv-primitive-frost-3);
  }

  .phase-design .phase-indicator {
    background: var(--mv-primitive-aurora-purple);
    box-shadow: var(--mv-shadow-phase-glow-sm) var(--mv-primitive-aurora-purple);
  }

  .phase-impl .phase-indicator {
    background: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-phase-glow-sm) var(--mv-primitive-aurora-green);
  }

  .phase-verify .phase-indicator {
    background: var(--mv-primitive-aurora-yellow);
    box-shadow: var(--mv-shadow-phase-glow-sm) var(--mv-primitive-aurora-yellow);
  }

  /* ノードラベル */
  .node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    text-shadow: var(--mv-text-shadow-subtle);
  }

  /* 進捗バー */
  .progress-container {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-left: auto;
  }

  .progress-text {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-muted);
    min-width: var(--mv-progress-text-width-lg);
    text-align: right;
  }

  .retry-info {
    font-size: var(--mv-font-size-xs);
    color: var(--mv-color-status-retry-wait-text);
    margin-left: var(--mv-spacing-xs);
    font-family: var(--mv-font-mono);
    opacity: 0.9;
    white-space: nowrap;
  }
</style>
