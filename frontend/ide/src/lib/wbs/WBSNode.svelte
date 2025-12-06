<script lang="ts">
  import type { WBSNode } from '../../stores/wbsStore';
  import { expandedNodes, viewMode } from '../../stores/wbsStore';
  import { selectedTaskId } from '../../stores';
  import { statusToCssClass, statusLabels } from '../../types';
  import type { PhaseName } from '../../types';

  // Props
  export let node: WBSNode;
  export let expanded: boolean = true;

  // フェーズ名からCSSクラス名への変換
  function phaseToClass(phase: PhaseName | undefined): string {
    if (!phase) return '';
    const phaseMap: Record<string, string> = {
      概念設計: 'phase-concept',
      実装設計: 'phase-design',
      実装: 'phase-impl',
      検証: 'phase-verify',
    };
    return phaseMap[phase] || '';
  }

  $: isPhase = node.type === 'phase';
  $: isTask = node.type === 'task';
  $: hasChildren = node.children.length > 0;
  $: phaseClass = phaseToClass(node.phaseName);
  $: statusClass = node.task ? statusToCssClass(node.task.status) : '';
  $: isSelected = node.task && $selectedTaskId === node.task.id;
  $: progressPercent = node.progress.percentage;
  $: indentStyle = `padding-left: ${node.level * 24 + 12}px`;

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
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      if (isPhase && hasChildren) {
        expandedNodes.toggle(node.id);
      } else {
        handleClick();
      }
    }
  }
</script>

<div
  class="wbs-node"
  class:is-phase={isPhase}
  class:is-task={isTask}
  class:selected={isSelected}
  class:status-pending={statusClass === 'pending'}
  class:status-ready={statusClass === 'ready'}
  class:status-running={statusClass === 'running'}
  class:status-succeeded={statusClass === 'succeeded'}
  class:status-failed={statusClass === 'failed'}
  class:status-canceled={statusClass === 'canceled'}
  class:status-blocked={statusClass === 'blocked'}
  class:phase-concept={phaseClass === 'phase-concept'}
  class:phase-design={phaseClass === 'phase-design'}
  class:phase-impl={phaseClass === 'phase-impl'}
  class:phase-verify={phaseClass === 'phase-verify'}
  style={indentStyle}
  role={isPhase ? 'treeitem' : 'button'}
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
      aria-label={expanded ? '折りたたむ' : '展開する'}
    >
      <span class="toggle-icon" class:expanded>{expanded ? '▼' : '▶'}</span>
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
    <span class="status-badge">
      {statusLabels[node.task.status]}
    </span>
  {/if}

  <!-- 進捗バー -->
  {#if isPhase}
    <div class="progress-container">
      <div class="progress-bar">
        <div class="progress-fill" style:--progress="{progressPercent}%"></div>
      </div>
      <span class="progress-text">
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
    cursor: pointer;
    border-radius: var(--mv-radius-sm);
    transition: background-color var(--mv-transition-hover);
    user-select: none;
  }

  .wbs-node:hover {
    background: var(--mv-color-surface-hover);
  }

  .wbs-node:focus {
    outline: none;
    background: var(--mv-color-surface-hover);
    box-shadow: var(--mv-shadow-focus);
  }

  .wbs-node.selected {
    background: var(--mv-color-surface-selected);
  }

  /* フェーズノード */
  .wbs-node.is-phase {
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
  }

  .wbs-node.is-phase .node-label {
    font-size: var(--mv-font-size-base);
  }

  /* タスクノード */
  .wbs-node.is-task {
    font-weight: var(--mv-font-weight-normal);
    color: var(--mv-color-text-secondary);
  }

  .wbs-node.is-task:hover {
    color: var(--mv-color-text-primary);
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
    transition: background-color var(--mv-transition-hover);
  }

  .toggle-btn:hover {
    background: var(--mv-color-surface-secondary);
    color: var(--mv-color-text-primary);
  }

  .toggle-icon {
    font-size: var(--mv-wbs-toggle-icon-size);
    transition: transform var(--mv-transition-hover);
  }

  .task-bullet {
    width: var(--mv-wbs-toggle-size);
    text-align: center;
    color: var(--mv-color-text-muted);
    font-size: var(--mv-font-size-sm);
  }

  .spacer {
    width: var(--mv-wbs-toggle-size);
  }

  /* フェーズインジケーター */
  .phase-indicator {
    width: var(--mv-wbs-phase-bar-width);
    height: var(--mv-wbs-toggle-size);
    border-radius: var(--mv-radius-sm);
    flex-shrink: 0;
  }

  .phase-concept .phase-indicator {
    background: var(--mv-primitive-frost-3);
  }

  .phase-design .phase-indicator {
    background: var(--mv-primitive-aurora-purple);
  }

  .phase-impl .phase-indicator {
    background: var(--mv-primitive-aurora-green);
  }

  .phase-verify .phase-indicator {
    background: var(--mv-primitive-aurora-yellow);
  }

  /* ノードラベル */
  .node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* ステータスバッジ */
  .status-badge {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-bold);
    text-transform: uppercase;
    letter-spacing: var(--mv-letter-spacing-wide);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
  }

  .status-pending .status-badge {
    background: var(--mv-color-status-pending-bg);
    color: var(--mv-color-status-pending-text);
  }

  .status-ready .status-badge {
    background: var(--mv-color-status-ready-bg);
    color: var(--mv-color-status-ready-text);
  }

  .status-running .status-badge {
    background: var(--mv-color-status-running-bg);
    color: var(--mv-color-status-running-text);
  }

  .status-succeeded .status-badge {
    background: var(--mv-color-status-succeeded-bg);
    color: var(--mv-color-status-succeeded-text);
  }

  .status-failed .status-badge {
    background: var(--mv-color-status-failed-bg);
    color: var(--mv-color-status-failed-text);
  }

  .status-canceled .status-badge {
    background: var(--mv-color-status-canceled-bg);
    color: var(--mv-color-status-canceled-text);
  }

  .status-blocked .status-badge {
    background: var(--mv-color-status-blocked-bg);
    color: var(--mv-color-status-blocked-text);
  }

  /* 進捗バー */
  .progress-container {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    margin-left: auto;
  }

  .progress-bar {
    width: var(--mv-progress-bar-width-sm);
    height: var(--mv-progress-bar-height-sm);
    background: var(--mv-color-surface-secondary);
    border-radius: var(--mv-radius-sm);
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    width: var(--progress, 0%);
    background: var(--mv-color-status-succeeded-border);
    border-radius: var(--mv-radius-sm);
    transition: width var(--mv-transition-hover);
  }

  .progress-text {
    font-size: var(--mv-font-size-xs);
    font-family: var(--mv-font-mono);
    color: var(--mv-color-text-muted);
    min-width: var(--mv-progress-text-width-lg);
    text-align: right;
  }
</style>
