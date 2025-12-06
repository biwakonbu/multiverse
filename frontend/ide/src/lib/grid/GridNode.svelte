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
    <!-- ステータスヘッダー（コンパクト） -->
    <div class="node-header">
      <span class="status-dot"></span>
      <span class="status-text">{statusLabels[task.status]}</span>
    </div>

    <!-- タイトル（メインコンテンツ） -->
    {#if showTitle}
      <div class="title" title={task.title}>
        {task.title}
      </div>
    {/if}

    <!-- フッター（フェーズタグ + メタ情報） -->
    <div class="node-footer">
      {#if task.phaseName}
        <span class="phase-tag">{phaseLabels[task.phaseName] || ""}</span>
      {/if}
      <div class="meta-info">
        {#if showDetails}
          <span class="pool-id">{task.poolId}</span>
        {/if}
        {#if hasDependencies && showDetails}
          <div class="deps-indicator" title="Dependencies">
            <span class="deps-count">↳{task.dependencies?.length || 0}</span>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .node {
    position: absolute;
    width: var(--mv-grid-cell-width);
    height: var(--mv-grid-cell-height);
    border-radius: var(--mv-radius-lg);
    cursor: pointer;
    transition:
      transform var(--mv-duration-fast) var(--mv-easing-out),
      box-shadow var(--mv-duration-fast) var(--mv-easing-out),
      z-index 0s;
    user-select: none;
    z-index: 10;
  }

  /* ホバー時は少し浮く */
  .node:hover {
    transform: translateY(-3px) scale(1.02);
    z-index: 20;
  }

  .node:active {
    transform: translateY(0) scale(1);
  }

  /* ガラス背景 - Crystal HUD スタイル */
  .node-glass {
    position: absolute;
    inset: 0;
    border-radius: var(--mv-radius-lg);
    background: var(--mv-glass-bg-chat);

    /* Multi-layer border for depth */
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-light);
    border-bottom: var(--mv-border-width-thin) solid
      var(--mv-glass-border-bottom);

    box-shadow: var(--mv-shadow-glass-panel-full);

    transition: all var(--mv-duration-fast) var(--mv-easing-out);
  }

  .node:hover .node-glass {
    background: var(--mv-glass-hover);
    border-color: var(--mv-glass-border-hover);
    box-shadow: var(--mv-shadow-glass-panel-with-glow);
  }

  /* 選択状態 - 強調されたグロー */
  .node.selected .node-glass {
    border-color: var(--mv-shadow-glow-accent-border);
    box-shadow: var(--mv-shadow-floating-with-accent-inset);
    background: var(--mv-glow-frost-2-lighter);
  }

  /* ノードグロー (実行中など) - より洗練されたエフェクト */
  .node-glow {
    position: absolute;
    inset: calc(-1 * var(--mv-spacing-xs));
    border-radius: var(--mv-radius-xl, 16px);
    opacity: 0;
    transition: opacity var(--mv-duration-normal);
    pointer-events: none;
    background: radial-gradient(
      ellipse at center,
      var(--mv-color-glow-focus) 0%,
      transparent 60%
    );
    filter: blur(4px);
  }

  .node.selected .node-glow {
    opacity: 0.35;
  }

  .node.status-running .node-glow {
    opacity: 0.5;
    background: radial-gradient(
      ellipse at center,
      var(--mv-color-status-running-glow) 0%,
      transparent 60%
    );
    animation: pulse-glow 2s infinite ease-in-out;
  }

  @keyframes pulse-glow {
    0%,
    100% {
      opacity: 0.35;
      transform: scale(0.98);
      filter: blur(4px);
    }
    50% {
      opacity: 0.7;
      transform: scale(1.03);
      filter: blur(6px);
    }
  }

  .node-content {
    position: relative;
    z-index: 1;
    height: 100%;
    padding: var(--mv-spacing-sm) var(--mv-spacing-md);
    padding-top: var(--mv-spacing-md);
    display: flex;
    flex-direction: column;
    gap: var(--mv-spacing-xxs);
  }

  /* フェーズインジケーター（上部細線）- 強化されたグロー */
  .phase-indicator-top {
    position: absolute;
    top: 0;
    left: var(--mv-spacing-md);
    right: var(--mv-spacing-md);
    height: var(--mv-size-progress-height);
    background: var(--mv-glass-border-light);
    border-radius: 0 0 var(--mv-radius-progress) var(--mv-radius-progress);
    transition: all var(--mv-duration-fast);
  }

  .phase-concept .phase-indicator-top {
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--mv-primitive-frost-3) 20%,
      var(--mv-primitive-frost-3) 80%,
      transparent 100%
    );
    box-shadow: var(--mv-shadow-phase-concept);
  }
  .phase-design .phase-indicator-top {
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--mv-primitive-aurora-purple) 20%,
      var(--mv-primitive-aurora-purple) 80%,
      transparent 100%
    );
    box-shadow: var(--mv-shadow-phase-design);
  }
  .phase-impl .phase-indicator-top {
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--mv-primitive-aurora-green) 20%,
      var(--mv-primitive-aurora-green) 80%,
      transparent 100%
    );
    box-shadow: var(--mv-shadow-phase-impl);
  }
  .phase-verify .phase-indicator-top {
    background: linear-gradient(
      90deg,
      transparent 0%,
      var(--mv-primitive-aurora-yellow) 20%,
      var(--mv-primitive-aurora-yellow) 80%,
      transparent 100%
    );
    box-shadow: var(--mv-shadow-phase-verify);
  }

  /* ヘッダー（コンパクト・ステータスのみ） */
  .node-header {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xxs);
    flex-shrink: 0;
  }

  .status-dot {
    width: var(--mv-status-dot-size);
    height: var(--mv-status-dot-size);
    border-radius: var(--mv-radius-full);
    background: var(--mv-color-text-muted);
    transition: all var(--mv-duration-fast);
  }

  .status-text {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-bold);
    letter-spacing: var(--mv-letter-spacing-count);
    color: var(--mv-color-text-muted);
    text-transform: uppercase;
    transition: all var(--mv-duration-fast);
  }

  /* ステータス別の色 - 強化されたグロー */
  .node.status-pending .status-dot {
    background: var(--mv-color-status-pending-text);
    box-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-color-status-pending-text);
  }
  .node.status-pending .status-text {
    color: var(--mv-color-status-pending-text);
    text-shadow: var(--mv-text-shadow-orange);
  }

  .node.status-ready .status-dot {
    background: var(--mv-color-status-ready-text);
    box-shadow: var(--mv-shadow-badge-glow-md) var(--mv-color-status-ready-text);
  }
  .node.status-ready .status-text {
    color: var(--mv-color-status-ready-text);
    text-shadow: var(--mv-text-shadow-cyan-content);
  }

  .node.status-running .status-dot {
    background: var(--mv-color-status-running-text);
    box-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-color-status-running-text);
    animation: dot-pulse 1.5s infinite ease-in-out;
  }
  .node.status-running .status-text {
    color: var(--mv-color-status-running-text);
    text-shadow: var(--mv-text-shadow-green);
  }

  @keyframes dot-pulse {
    0%,
    100% {
      box-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-color-status-running-text);
    }
    50% {
      box-shadow: var(--mv-shadow-glow-frost-2-md) var(--mv-color-status-running-text);
    }
  }

  .node.status-succeeded .status-dot {
    background: var(--mv-color-status-succeeded-text);
    box-shadow: var(--mv-shadow-badge-glow-md) var(--mv-color-status-succeeded-text);
  }
  .node.status-succeeded .status-text {
    color: var(--mv-color-status-succeeded-text);
    text-shadow: var(--mv-text-shadow-cyan-content);
  }

  .node.status-failed .status-dot {
    background: var(--mv-color-status-failed-text);
    box-shadow: var(--mv-shadow-badge-glow-md) var(--mv-color-status-failed-text);
  }
  .node.status-failed .status-text {
    color: var(--mv-color-status-failed-text);
    text-shadow: var(--mv-shadow-badge-glow-lg) var(--mv-glow-failed);
  }

  .node.status-blocked .status-dot {
    background: var(--mv-color-status-blocked-text);
    box-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-color-status-blocked-text);
  }
  .node.status-blocked .status-text {
    color: var(--mv-color-status-blocked-text);
    text-shadow: var(--mv-text-shadow-purple-content);
  }

  /* フェーズタグ - ガラススタイル */
  .phase-tag {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-bold);
    padding: var(--mv-spacing-badge-x) var(--mv-spacing-xs);
    border-radius: var(--mv-radius-sm);
    background: var(--mv-glass-bg-dark);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-count);
    transition: all var(--mv-duration-fast);
  }

  .phase-concept .phase-tag {
    color: var(--mv-primitive-frost-3);
    border-color: var(--mv-glow-frost-3);
    text-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-glow-frost-3-strong);
  }
  .phase-design .phase-tag {
    color: var(--mv-primitive-aurora-purple);
    border-color: var(--mv-glow-purple);
    text-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-glow-purple-strong);
  }
  .phase-impl .phase-tag {
    color: var(--mv-primitive-aurora-green);
    border-color: var(--mv-glow-green);
    text-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-glow-green-strong);
  }
  .phase-verify .phase-tag {
    color: var(--mv-primitive-aurora-yellow);
    border-color: var(--mv-glow-yellow);
    text-shadow: var(--mv-shadow-badge-glow-sm) var(--mv-glow-yellow-strong);
  }

  /* タイトル - 強化されたテキストスタイル */
  .title {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-primary);
    line-height: var(--mv-line-height-tight);

    /* 3行まで表示 */
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
    flex: 1;

    text-shadow: var(--mv-text-shadow-base-white);
  }

  .node:hover .title {
    color: var(--mv-primitive-snow-storm-2);
    text-shadow: var(--mv-text-shadow-hover-white);
  }

  /* フッター - フェーズタグとメタ情報 */
  .node-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--mv-spacing-xs);
    margin-top: auto;
    padding-top: var(--mv-spacing-xs);
    border-top: var(--mv-border-width-thin) solid var(--mv-glass-border-subtle);
    min-height: var(--mv-min-height-footer);
  }

  .meta-info {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
    font-family: var(--mv-font-mono);
    font-size: var(--mv-font-size-xxs);
    color: var(--mv-color-text-disabled);
    margin-left: auto;
  }

  .pool-id {
    opacity: 0.7;
    transition: all var(--mv-duration-fast);
  }

  .node:hover .pool-id {
    opacity: 1;
    color: var(--mv-primitive-frost-2);
  }

  .deps-indicator {
    display: flex;
    align-items: center;
    opacity: 0.8;
    color: var(--mv-primitive-frost-2);
  }

  .deps-count {
    font-weight: var(--mv-font-weight-bold);
    font-size: var(--mv-font-size-xxs);
  }
</style>
