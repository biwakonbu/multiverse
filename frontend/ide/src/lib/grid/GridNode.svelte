<script lang="ts">
  import type { Task, PhaseName } from "../../types";
  import { gridToCanvas } from "../../design-system";
  import { statusToCssClass, statusLabels } from "../../types";
  import { selectedTaskId } from "../../stores";

  // Props
  export let task: Task;
  export let col: number;
  export let row: number;
  export let zoomLevel: number = 1;

  // フェーズ名からCSSクラス名への変換
  function phaseToClass(phase: PhaseName | undefined): string {
    if (!phase) return "";
    const phaseMap: Record<string, string> = {
      概念設計: "phase-concept",
      実装設計: "phase-design",
      実装: "phase-impl",
      検証: "phase-verify",
    };
    return phaseMap[phase] || "";
  }

  // フェーズの表示ラベル
  const phaseLabels: Record<string, string> = {
    概念設計: "CONCEPT",
    実装設計: "DESIGN",
    実装: "IMPL",
    検証: "VERIFY",
  };

  // キャンバス座標を計算
  $: position = gridToCanvas(col, row);
  $: isSelected = $selectedTaskId === task.id;
  $: statusClass = statusToCssClass(task.status);
  $: phaseClass = phaseToClass(task.phaseName);
  $: hasDependencies = task.dependencies && task.dependencies.length > 0;

  // ズームレベルに応じた表示制御
  $: showTitle = zoomLevel >= 0.4;
  $: showDetails = zoomLevel >= 1.2;

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
  class="node status-{statusClass} {phaseClass}"
  class:selected={isSelected}
  class:has-deps={hasDependencies}
  style="left: {position.x}px; top: {position.y}px;"
  on:click={handleClick}
  on:keydown={handleKeydown}
  role="button"
  tabindex="0"
  aria-label="{task.title} - {statusLabels[task.status]}"
>
  <!-- 背景グロー（選択時/実行中用） -->
  <div class="node-glow"></div>

  <!-- ガラス背景レイヤー -->
  <div class="node-glass"></div>

  <!-- フェーズインジケータ（上部バー） -->
  {#if task.phaseName && phaseClass}
    <div class="phase-indicator-top" aria-hidden="true"></div>
  {/if}

  <!-- コンテンツコンテナ -->
  <div class="node-content">
    <!-- ステータスヘッダー -->
    <div class="node-header">
      <div class="status-badge">
        <span class="status-dot"></span>
        <span class="status-text">{statusLabels[task.status]}</span>
      </div>
      {#if task.phaseName}
        <span class="phase-tag">{phaseLabels[task.phaseName] || ""}</span>
      {/if}
    </div>

    <!-- タイトル -->
    {#if showTitle}
      <div class="title" title={task.title}>
        {task.title}
      </div>
    {/if}

    <!-- 詳細フッター -->
    {#if showDetails}
      <div class="node-footer">
        <span class="pool-id">{task.poolId}</span>
        {#if hasDependencies}
          <div class="deps-indicator" title="Dependencies">
            <span class="deps-arrow">↳</span>
            <span class="deps-count">{task.dependencies?.length || 0}</span>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  .node {
    position: absolute;
    width: var(--mv-grid-cell-width);
    height: var(--mv-grid-cell-height);
    border-radius: var(--mv-radius-md);
    cursor: pointer;
    transition:
      transform var(--mv-duration-fast) var(--mv-easing-out),
      z-index 0s;
    user-select: none;
    z-index: 10;
  }

  /* ホバー時は少し浮く */
  .node:hover {
    transform: translateY(-2px);
    z-index: 20;
  }

  .node:active {
    transform: translateY(0);
  }

  /* ガラス背景 */
  .node-glass {
    position: absolute;
    inset: 0;
    border-radius: var(--mv-radius-md);
    background: var(--mv-glass-bg);
    backdrop-filter: blur(8px);
    border: var(--mv-border-width-thin) solid var(--mv-color-border-subtle);
    box-shadow: var(--mv-shadow-card);
    transition: all var(--mv-duration-fast);
  }

  .node:hover .node-glass {
    background: var(--mv-glass-hover);
    border-color: var(--mv-color-border-default);
    box-shadow: var(--mv-shadow-ambient-lg);
  }

  /* 選択状態 */
  .node.selected .node-glass {
    border-color: var(--mv-color-interactive-primary);
    box-shadow: var(--mv-shadow-selected);
    background: var(--mv-glass-hover-strong);
  }

  /* ノードグロー (実行中など) */
  .node-glow {
    position: absolute;
    inset: calc(-1 * var(--mv-spacing-xxs));
    border-radius: var(--mv-radius-lg);
    opacity: 0;
    transition: opacity var(--mv-duration-normal);
    pointer-events: none;
    background: radial-gradient(
      circle at center,
      var(--mv-color-glow-focus) 0%,
      transparent 70%
    );
  }

  .node.selected .node-glow {
    opacity: 0.2;
  }

  .node.status-running .node-glow {
    opacity: 0.4;
    background: radial-gradient(
      circle at center,
      var(--mv-color-status-running-glow) 0%,
      transparent 70%
    );
    animation: pulse-glow 2s infinite ease-in-out;
  }

  @keyframes pulse-glow {
    0%,
    100% {
      opacity: 0.3;
      transform: scale(0.95);
    }
    50% {
      opacity: 0.6;
      transform: scale(1.05);
    }
  }

  .node-content {
    position: relative;
    z-index: 1;
    height: 100%;
    padding: var(--mv-spacing-sm);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xs);
  }

  /* フェーズインジケーター（上部細線） */
  .phase-indicator-top {
    position: absolute;
    top: 0;
    left: var(--mv-spacing-sm);
    right: var(--mv-spacing-sm);
    height: var(--mv-spacing-xxxs);
    background: var(--mv-color-border-subtle);
    box-shadow: var(--mv-shadow-badge);
    border-radius: 0 0 var(--mv-spacing-xxxs) var(--mv-spacing-xxxs);
  }

  .phase-concept .phase-indicator-top {
    background: var(--mv-primitive-frost-3);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-primitive-frost-3);
  }
  .phase-design .phase-indicator-top {
    background: var(--mv-primitive-aurora-purple);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-primitive-aurora-purple);
  }
  .phase-impl .phase-indicator-top {
    background: var(--mv-primitive-aurora-green);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-primitive-aurora-green);
  }
  .phase-verify .phase-indicator-top {
    background: var(--mv-primitive-aurora-yellow);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-primitive-aurora-yellow);
  }

  /* ヘッダー */
  .node-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: var(--mv-indicator-size-lg);
  }

  .status-badge {
    display: flex;
    align-items: center;
    gap: var(--mv-status-dot-size);
  }

  .status-dot {
    width: var(--mv-status-dot-size);
    height: var(--mv-status-dot-size);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-text-muted);
    box-shadow: var(--mv-text-shadow-subtle);
  }

  .status-text {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-count);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
  }

  /* ステータス別の色 */
  .node.status-pending .status-dot {
    background: var(--mv-color-status-pending-text);
  }
  .node.status-pending .status-text {
    color: var(--mv-color-status-pending-text);
  }

  .node.status-ready .status-dot {
    background: var(--mv-color-status-ready-text);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-color-status-ready-text);
  }
  .node.status-ready .status-text {
    color: var(--mv-color-status-ready-text);
  }

  .node.status-running .status-dot {
    background: var(--mv-color-status-running-text);
    box-shadow: var(--mv-shadow-badge-glow-md) var(--mv-color-status-running-text);
  }
  .node.status-running .status-text {
    color: var(--mv-color-status-running-text);
    text-shadow: var(--mv-text-shadow-green-content);
  }

  .node.status-succeeded .status-dot {
    background: var(--mv-color-status-succeeded-text);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-color-status-succeeded-text);
  }
  .node.status-succeeded .status-text {
    color: var(--mv-color-status-succeeded-text);
  }

  .node.status-failed .status-dot {
    background: var(--mv-color-status-failed-text);
    box-shadow: var(--mv-shadow-phase-indicator-glow) var(--mv-color-status-failed-text);
  }
  .node.status-failed .status-text {
    color: var(--mv-color-status-failed-text);
  }

  .node.status-blocked .status-dot {
    background: var(--mv-color-status-blocked-text);
  }
  .node.status-blocked .status-text {
    color: var(--mv-color-status-blocked-text);
  }

  /* フェーズタグ */
  .phase-tag {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-bold);
    padding: var(--mv-border-width-thin) var(--mv-spacing-xxs);
    border-radius: var(--mv-spacing-xxxs);
    background: var(--mv-glass-active);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-count);
  }

  .phase-concept .phase-tag {
    color: var(--mv-primitive-frost-3);
    background: var(--mv-glass-active);
  }
  .phase-design .phase-tag {
    color: var(--mv-primitive-aurora-purple);
    background: var(--mv-glass-active);
  }
  .phase-impl .phase-tag {
    color: var(--mv-primitive-aurora-green);
    background: var(--mv-glass-active);
  }
  .phase-verify .phase-tag {
    color: var(--mv-primitive-aurora-yellow);
    background: var(--mv-glass-active);
  }

  /* タイトル */
  .title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-medium);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-tight);

    /* 3行まで表示 */
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    flex: 1;

    text-shadow: var(--mv-text-shadow-base);
  }

  /* フッター */
  .node-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xxs);
    color: var(--mv-color-text-disabled);
    margin-top: auto;
  }

  .pool-id {
    opacity: 0.7;
  }

  .deps-indicator {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxxs);
    opacity: 0.8;
  }

  .deps-arrow {
    font-size: var(--mv-font-size-xs);
  }
</style>
