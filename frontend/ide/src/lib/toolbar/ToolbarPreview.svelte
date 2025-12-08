<script lang="ts">
  import type { TaskStatus, PoolSummary } from "../../types";
  import BrandText from "../components/brand/BrandText.svelte";
  import ProgressBar from "../wbs/ProgressBar.svelte";
  import Button from "../../design-system/components/Button.svelte";
  import { Network, ListTree, Settings } from "lucide-svelte";

  interface Props {
    // Props (matching Stores in Toolbar.svelte)
    poolSummaries?: PoolSummary[];
    overallProgress?: any;
    viewMode?: "graph" | "wbs";
    taskCountsByStatus?: Record<TaskStatus, number>;
    onviewmodechange?: (mode: "graph" | "wbs") => void;
  }

  let {
    poolSummaries = [],
    overallProgress = { percentage: 0, completed: 0, total: 0 },
    viewMode = "graph",
    taskCountsByStatus = {
      PENDING: 0,
      READY: 0,
      RUNNING: 0,
      SUCCEEDED: 0,
      COMPLETED: 0,
      FAILED: 0,
      CANCELED: 0,
      BLOCKED: 0,
      RETRY_WAIT: 0,
    },
    onviewmodechange,
  }: Props = $props();

  // ステータスサマリの表示設定（フォールバック用）
  const statusDisplay: {
    key: TaskStatus;
    label: string;
    showCount: boolean;
    cssClass: string;
  }[] = [
    {
      key: "RUNNING",
      label: "RUNNING",
      showCount: true,
      cssClass: "running",
    },
    {
      key: "PENDING",
      label: "PENDING",
      showCount: true,
      cssClass: "pending",
    },
    {
      key: "FAILED",
      label: "FAILED",
      showCount: true,
      cssClass: "failed",
    },
  ];

  function setGraphMode() {
    onviewmodechange?.("graph");
  }

  function setWBSMode() {
    onviewmodechange?.("wbs");
  }

  // Pool別サマリがある場合はそれを表示、なければステータス別サマリを表示
  let hasPoolSummaries = $derived(poolSummaries.length > 0);
  let isGraphMode = $derived(viewMode === "graph");
</script>

<header class="toolbar crystal-hud">
  <!-- 左側：ブランド -->
  <div class="toolbar-section left">
    <BrandText size="sm" />
  </div>

  <!-- 中央：Holographic Data Strip -->
  <div class="toolbar-section center">
    <div class="holographic-strip">
      {#if hasPoolSummaries}
        {#each poolSummaries as pool (pool.poolId)}
          <!-- Pool ID -->
          <div class="holo-group">
            <span class="holo-label">POOL</span>
            <span class="holo-value pool-id">{pool.poolId}</span>
          </div>

          <div class="holo-divider"></div>

          <!-- Stats -->
          <div class="holo-stats">
            {#if pool.running > 0}
              <div class="holo-stat running">
                <span class="stat-value">{pool.running}</span>
                <span class="stat-label">RUN</span>
              </div>
            {/if}
            {#if pool.queued > 0}
              <div class="holo-stat pending">
                <span class="stat-value">{pool.queued}</span>
                <span class="stat-label">WAIT</span>
              </div>
            {/if}
            {#if pool.failed > 0}
              <div class="holo-stat failed">
                <span class="stat-value">{pool.failed}</span>
                <span class="stat-label">FAIL</span>
              </div>
            {/if}
            {#if pool.running === 0 && pool.queued === 0 && pool.failed === 0}
              <div class="holo-stat idle">
                <span class="stat-value">{pool.total}</span>
                <span class="stat-label">TASKS</span>
              </div>
            {/if}
          </div>
        {/each}
      {:else}
        <!-- Fallback Status Strip -->
        {#each statusDisplay as { key, label, showCount, cssClass }}
          {#if showCount && taskCountsByStatus[key] > 0}
            <div class="holo-stat {cssClass}">
              <span class="stat-value">{taskCountsByStatus[key]}</span>
              <span class="stat-label">{label}</span>
            </div>
            <div class="holo-divider-sm"></div>
          {/if}
        {/each}
      {/if}
    </div>
  </div>

  <!-- 右側：Command Capsule & View Switch -->
  <div class="toolbar-section right">
    <!-- Command Capsule -->
    <div class="command-capsule">
      <!-- Removed ExecutionControls -->
    </div>

    <!-- Progress -->
    <div class="progress-module">
      <ProgressBar percentage={overallProgress.percentage} size="mini" />
      <span class="progress-readout">{overallProgress.percentage}%</span>
    </div>

    <!-- Segmented Crystal Switch -->
    <div class="crystal-switch">
      <button
        class="switch-item"
        class:active={isGraphMode}
        onclick={setGraphMode}
        title="Graph View"
      >
        <Network size="16" />
      </button>
      <button
        class="switch-item"
        class:active={!isGraphMode}
        onclick={setWBSMode}
        title="WBS View"
      >
        <ListTree size="16" />
      </button>
    </div>
  </div>
</header>

<style>
  :global(:root) {
    --hud-font: "Rajdhani", sans-serif;
    --hud-glass-bg: var(--mv-glass-bg);
    --hud-glass-border: var(--mv-glass-border-strong);
    --hud-glass-highlight: var(--mv-glass-border);
    --hud-neon-blue: var(--mv-primitive-frost-1);
    --hud-neon-green: var(--mv-primitive-aurora-green);
    --hud-neon-red: var(--mv-primitive-aurora-red);
    --hud-neon-yellow: var(--mv-primitive-aurora-yellow);
  }

  .crystal-hud {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: var(--mv-layout-toolbar-height);
    padding: 0 var(--mv-spacing-xl);

    background: var(--hud-glass-bg);
    backdrop-filter: blur(16px);
    border-bottom: var(--mv-border-width-thin) solid var(--hud-glass-border);
    box-shadow: var(--mv-shadow-lg);

    font-family: var(--hud-font);
    z-index: 100;
  }

  .toolbar-section {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-lg);
  }

  .toolbar-section.left {
    flex: 0 0 auto;
  }
  .toolbar-section.center {
    flex: 1;
    justify-content: center;
  }
  .toolbar-section.right {
    flex: 0 0 auto;
    justify-content: flex-end;
  }

  /* Holographic Data Strip */
  .holographic-strip {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-md);
    padding: var(--mv-spacing-xxs) var(--mv-spacing-lg);
    border-radius: var(--mv-radius-full);
    background: var(--mv-glass-bg-strong);
    box-shadow: var(--mv-shadow-inset-sm);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .holo-group {
    display: flex;
    align-items: baseline;
    gap: var(--mv-spacing-xs);
  }

  .holo-label {
    font-size: var(--mv-font-size-xs);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-wide);
  }

  .holo-value {
    font-size: var(--mv-font-size-base);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
  }

  .pool-id {
    color: var(--hud-neon-blue);
    text-shadow: var(--mv-glow-blue);
  }

  .holo-divider {
    width: var(--mv-border-width-thin);
    height: var(--mv-spacing-lg);
    background: var(--mv-color-border-subtle);
    opacity: 0.5;
  }

  .holo-divider-sm {
    width: var(--mv-border-width-thin);
    height: var(--mv-spacing-md);
    background: var(--mv-color-border-subtle);
    opacity: 0.3;
    margin: 0 var(--mv-spacing-xxs);
  }

  .holo-stats {
    display: flex;
    gap: var(--mv-spacing-lg);
  }

  .holo-stat {
    display: flex;
    align-items: baseline;
    gap: var(--mv-spacing-xxs);
    line-height: 1;
  }

  .stat-value {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-bold);
  }

  .stat-label {
    font-size: var(--mv-font-size-xxs);
    font-weight: var(--mv-font-weight-semibold);
    opacity: 0.8;
    letter-spacing: var(--mv-letter-spacing-normal);
  }

  .holo-stat.running {
    color: var(--hud-neon-green);
    text-shadow: var(--mv-glow-green);
  }

  .holo-stat.pending {
    color: var(--hud-neon-yellow);
  }

  .holo-stat.failed {
    color: var(--hud-neon-red);
    text-shadow: var(--mv-glow-red);
  }

  .holo-stat.idle {
    color: var(--mv-color-text-muted);
  }

  /* Command Capsule */
  .command-capsule {
    display: flex;
    align-items: center;
    padding: var(--mv-spacing-xxxs);
    background: var(--mv-glass-border);
    border-radius: var(--mv-radius-md);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  /* Progress Module */
  .progress-module {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-sm);
  }

  .progress-readout {
    font-family: var(--hud-font);
    font-weight: var(--mv-font-weight-semibold);
    font-size: var(--mv-font-size-sm);
    color: var(--mv-color-text-secondary);
    min-width: var(--mv-zoom-label-min-width);
    text-align: right;
  }

  /* Segmented Crystal Switch */
  .crystal-switch {
    display: flex;
    padding: var(--mv-spacing-xxxs);
    gap: var(--mv-spacing-xxxs);
    background: var(--mv-glass-bg-strong);
    border-radius: var(--mv-radius-full);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .switch-item {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-icon-lg);
    height: var(--mv-size-icon-lg);
    border: none;
    border-radius: var(--mv-radius-full);
    background: transparent;
    color: var(--mv-color-text-muted);
    cursor: pointer;
    transition: all 0.25s cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .switch-item:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-border);
  }

  .switch-item.active {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
    box-shadow: var(--mv-shadow-sm);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border-strong);
  }
</style>
