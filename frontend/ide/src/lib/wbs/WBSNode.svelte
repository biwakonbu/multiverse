<script lang="ts">
  import { run } from 'svelte/legacy';

  import type { WBSNode } from "../../stores/wbsStore";
  import { expandedNodes } from "../../stores/wbsStore";
  import { selectedTaskId } from "../../stores";
  import { phaseToCssClass } from "../../schemas";
  import StatusBadge from "./StatusBadge.svelte";
  import ProgressBar from "./ProgressBar.svelte";
  import { getProgressColor } from "./utils";
  import { onDestroy } from "svelte";
  import { stripMarkdown } from "../utils/markdown";

  
  interface Props {
    // Props
    node: WBSNode;
    expanded?: boolean;
    index?: number;
  }

  let { node, expanded = true, index = 0 }: Props = $props();

  let isMilestone = $derived(node.type === "milestone");
  let isPhase = $derived(node.type === "phase");
  let isTask = $derived(node.type === "task");
  let hasChildren = $derived(node.children.length > 0);
  let phaseClass = $derived(phaseToCssClass(node.phaseName));
  let isSelected = $derived(node.task && $selectedTaskId === node.task.id);
  let progressPercent = $derived(node.progress.percentage);
  let progressColor = $derived(getProgressColor(progressPercent));
  let isOdd = $derived(index % 2 !== 0); // Check for zebra striping

  // Indentation with fixed width unit
  const INDENT_WIDTH = 20;
  const INDENT_BASE = 12;
  let indentStyle = $derived(`padding-left: ${node.level * INDENT_WIDTH + INDENT_BASE}px`);

  // Calculate guides for levels 0 to node.level - 1
  let indentGuides = $derived(Array(Math.max(0, node.level)).fill(0));

  // Retry Logic
  let isRetryWait = $derived(node.task?.status === "RETRY_WAIT");
  let attemptCount = $derived(node.task?.attemptCount || 0);
  let nextRetryAt = $derived(node.task?.nextRetryAt);

  let timeRemaining = $state("");
  let interval: any = $state();

  function updateTimeRemaining() {
    if (!nextRetryAt) {
      timeRemaining = "";
      return;
    }
    const now = new Date();
    const target = new Date(nextRetryAt);
    const diff = target.getTime() - now.getTime();

    if (diff <= 0) {
      timeRemaining = "";
    } else {
      const seconds = Math.ceil(diff / 1000);
      timeRemaining = `${seconds}s`;
    }
  }

  run(() => {
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
  });

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

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<div
  class="wbs-node"
  class:is-phase={isPhase}
  class:is-milestone={isMilestone}
  class:is-task={isTask}
  class:selected={isSelected}
  class:phase-concept={phaseClass === "phase-concept"}
  class:phase-design={phaseClass === "phase-design"}
  class:phase-impl={phaseClass === "phase-impl"}
  class:phase-verify={phaseClass === "phase-verify"}
  class:is-odd={isOdd}
  style={indentStyle}
  role={isPhase || isMilestone ? "treeitem" : "button"}
  tabindex="0"
  aria-expanded={(isPhase || isMilestone) && hasChildren ? expanded : undefined}
  aria-label={node.label}
  onclick={handleClick}
  onkeydown={handleKeydown}
>
  <!-- Indentation Guides -->
  <!-- stylelint-disable-next-line scale-unlimited/declaration-strict-value -->
  {#each indentGuides as _, i}
    <div
      class="indent-guide"
      style:left="{i * INDENT_WIDTH + INDENT_BASE + INDENT_WIDTH / 2 - 1}px"
    ></div>
  {/each}

  <!-- 展開/折りたたみトグル -->
  {#if (isPhase || isMilestone) && hasChildren}
    <button
      class="toggle-btn"
      onclick={handleToggle}
      aria-label={expanded ? "折りたたむ" : "展開する"}
    >
      <svg
        class="chevron"
        class:expanded
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <polyline points="9 18 15 12 9 6" />
      </svg>
    </button>
  {:else}
    <div class="spacer"></div>
  {/if}

  <!-- フェーズバー（カラーインジケーター） -->
  {#if phaseClass}
    <span class="phase-indicator" aria-hidden="true"></span>
  {/if}

  <!-- ラベル -->
  <span class="node-label">
    {stripMarkdown(node.label)}
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
  {#if isPhase || isMilestone}
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
    padding-right: var(--mv-spacing-md);
    cursor: pointer;
    border-bottom: var(--mv-border-width-thin) solid transparent;
    transition:
      background-color var(--mv-transition-hover),
      border-color var(--mv-transition-hover),
      box-shadow var(--mv-transition-hover),
      transform var(--mv-transition-hover);
    user-select: none;
    position: relative;
    color: var(--mv-color-text-secondary);
  }

  /* Zebra Striping (Subtle) */
  .wbs-node.is-odd {
    background: var(--mv-glass-active);
  }

  /* Indentation Guide */
  .indent-guide {
    position: absolute;
    top: 0;
    bottom: 0;
    width: var(--mv-spacing-xxxs);
    background: linear-gradient(
      to bottom,
      transparent,
      var(--mv-glass-border-subtle) 20%,
      var(--mv-glass-border-subtle) 80%,
      transparent
    );
    opacity: 0.5;
    pointer-events: none;
    z-index: 1;
  }

  /* Hover Effects */
  .wbs-node:hover {
    background: var(--mv-glass-hover);
    color: var(--mv-color-text-primary);
  }

  .wbs-node:focus {
    outline: none;
    background: var(--mv-glass-hover-strong);
    box-shadow: var(--mv-shadow-inset-focus) var(--mv-color-interactive-primary);
  }

  /* Selected State */
  .wbs-node.selected {
    background: var(--mv-glass-active);
    color: var(--mv-color-text-primary);
    box-shadow: var(--mv-shadow-wbs-node-selected);
  }

  /* Phase / Milestone Styling */
  .wbs-node.is-phase {
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-subtle);
  }

  .wbs-node.is-milestone {
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-strong);
    background: var(--mv-glass-bg-dark);
  }

  .wbs-node.is-phase:hover {
    background: var(--mv-glass-hover);
  }

  .wbs-node.is-milestone:hover {
    background: var(--mv-glass-hover-strong);
  }

  .wbs-node.is-phase .node-label {
    font-family: var(--mv-font-display);
    font-size: var(--mv-font-size-md);
    letter-spacing: var(--mv-letter-spacing-wide);
    text-shadow: var(--mv-text-shadow-subtle);
  }

  .wbs-node.is-milestone .node-label {
    font-family: var(--mv-font-display);
    font-size: var(--mv-font-size-lg);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  /* Task Node Styling */
  .wbs-node.is-task {
    font-weight: var(--mv-font-weight-normal);
    font-family: var(--mv-font-sans);
  }

  .wbs-node.is-task:hover {
    transform: translateX(2px);
  }

  /* Toggle Button */
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
    transition: color var(--mv-transition-hover);
  }

  .toggle-btn:hover {
    color: var(--mv-color-interactive-primary);
  }

  .chevron {
    width: var(--mv-icon-size-xs);
    height: var(--mv-icon-size-xs);
    transition: transform var(--mv-duration-fast) var(--mv-easing-spring);
  }

  .chevron.expanded {
    transform: rotate(90deg);
  }

  .spacer {
    width: var(--mv-wbs-toggle-size);
  }

  /* Phase Indicator */
  .phase-indicator {
    width: var(--mv-spacing-xxxs);
    height: var(--mv-icon-size-xs);
    border-radius: var(--mv-spacing-xxxs);
    flex-shrink: 0;
    opacity: 0.9;
    box-shadow: var(--mv-shadow-phase-glow-sm) var(--phase-color);
  }

  .phase-concept .phase-indicator {
    background: var(--mv-primitive-frost-3);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-primitive-frost-3);
  }
  .phase-design .phase-indicator {
    background: var(--mv-primitive-aurora-purple);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-primitive-aurora-purple);
  }
  .phase-impl .phase-indicator {
    background: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-primitive-aurora-green);
  }
  .phase-verify .phase-indicator {
    background: var(--mv-primitive-aurora-yellow);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-primitive-aurora-yellow);
  }

  /* Node Label */
  .node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Progress Container */
  .progress-container {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
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
