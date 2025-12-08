<!--
  TaskNodePreview - Storybook 用の TaskNode プレビューコンポーネント
  
  GridNodePreview.svelte と同じ Crystal HUD スタイルを使用。
  Storybook での単独表示・テスト用途。
-->
<script lang="ts">
  import type { TaskStatus, PhaseName } from "../../../types";
  import { stripMarkdown } from "../../utils/markdown";

  interface Props {
    id?: string;
    title?: string;
    status?: TaskStatus;
    phaseName?: PhaseName;
    poolId?: string;
    dependencies?: string[];
    zoomLevel?: number;
    selected?: boolean;
  }

  let {
    id = "task-1",
    title = "タスク名",
    status = "PENDING",
    phaseName = "概念設計",
    poolId = "codegen",
    dependencies = [],
    zoomLevel = 1.5,
    selected = false,
  }: Props = $props();

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
  let statusClass = $derived(status.toLowerCase());
  let hasDependencies = $derived(dependencies && dependencies.length > 0);
  let showTitle = $derived(zoomLevel >= 0.4);
  let showDetails = $derived(zoomLevel >= 1.2);

  // 3段階のノードサイズ（タイトル長に応じて）
  // small: ≤15文字, medium: ≤30文字, large: >30文字
  function getSizeClass(text: string): string {
    const len = text.length;
    if (len <= 15) return "size-small";
    if (len <= 30) return "size-medium";
    return "size-large";
  }

  let sizeClass = $derived(getSizeClass(title));
</script>

<div
  class="node status-{statusClass} {sizeClass}"
  class:selected
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
    <div class="title" {title}>
      {stripMarkdown(title)}
    </div>
  {/if}

  <!-- 詳細情報（高ズームレベルで表示） -->
  {#if showDetails}
    <div class="details">
      <span class="pool">{poolId}</span>
      {#if hasDependencies}
        <span class="deps">↳ {dependencies.length}</span>
      {/if}
    </div>
  {/if}
</div>

<style>
  /* GridNodePreview.svelte から完全コピーしたスタイル */
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

  /* ステータス別スタイル - 微妙な背景色変化 */
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
