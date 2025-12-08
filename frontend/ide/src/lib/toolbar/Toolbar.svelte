<script lang="ts">
  import {
    taskCountsByStatus,
    poolSummaries,
    viewMode,
    overallProgress,
  } from "../../stores";
  import type { TaskStatus } from "../../types";
  import BrandText from "../components/brand/BrandText.svelte";
  import ProgressBar from "../wbs/ProgressBar.svelte";
  import { Network, ListTree } from "lucide-svelte";

  // Badge status type (no longer directly used for badges, but conceptually for styling)
  type BadgeStatus =
    | "pending"
    | "ready"
    | "running"
    | "succeeded"
    | "failed"
    | "canceled"
    | "blocked";

  // TaskStatus → BadgeStatus マッピング (no longer directly used for badges)
  const statusToLower: Record<string, BadgeStatus> = {
    PENDING: "pending",
    READY: "ready",
    RUNNING: "running",
    SUCCEEDED: "succeeded",
    FAILED: "failed",
    CANCELED: "canceled",
    BLOCKED: "blocked",
  };

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

  // Pool別サマリがある場合はそれを表示、なければステータス別サマリを表示
  let hasPoolSummaries = $derived($poolSummaries.length > 0);
  let isGraphMode = $derived($viewMode === "graph");
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
        {#each $poolSummaries as pool (pool.poolId)}
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
          {#if showCount && $taskCountsByStatus[key] > 0}
            <div class="holo-stat {cssClass}">
              <span class="stat-value">{$taskCountsByStatus[key]}</span>
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
    <!-- Progress -->
    <div class="progress-module">
      <ProgressBar percentage={$overallProgress.percentage} size="mini" />
      <span class="progress-readout">{$overallProgress.percentage}%</span>
    </div>

    <!-- Segmented Crystal Switch -->
    <div class="crystal-switch">
      <button
        class="switch-item"
        class:active={isGraphMode}
        onclick={() => viewMode.setGraph()}
        title="Graph View"
      >
        <Network size="16" />
      </button>
      <button
        class="switch-item"
        class:active={!isGraphMode}
        onclick={() => viewMode.setWBS()}
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
    --hud-glass-highlight: var(--mv-glass-active);
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
    box-shadow: var(--mv-shadow-ambient-lg);

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
    padding: var(--mv-spacing-xxs) var(--mv-spacing-md);
    border-radius: var(--mv-spacing-xl);
    background: var(--mv-glass-bg-dark);
    box-shadow: var(--mv-shadow-inset-darker);
    border: var(--mv-border-width-thin) solid var(--mv-glass-border);
  }

  .holo-group {
    display: flex;
    align-items: baseline;
    gap: var(--mv-status-dot-size);
  }

  .holo-label {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    color: var(--mv-color-text-muted);
    letter-spacing: var(--mv-letter-spacing-tab);
  }

  .holo-value {
    font-size: var(--mv-font-size-lg);
    font-weight: var(--mv-font-weight-bold);
    color: var(--mv-color-text-primary);
  }

  .pool-id {
    color: var(--hud-neon-blue);
    text-shadow: var(--mv-text-shadow-cyan);
  }

  .holo-divider {
    width: var(--mv-border-width-thin);
    height: var(--mv-indicator-size-lg);
    background: var(--mv-color-border-subtle);
    opacity: 0.5;
  }

  .holo-divider-sm {
    width: var(--mv-border-width-thin);
    height: var(--mv-indicator-size-md);
    background: var(--mv-color-border-subtle);
    opacity: 0.3;
    margin: 0 var(--mv-spacing-xxs);
  }

  .holo-stats {
    display: flex;
    gap: var(--mv-spacing-md);
  }

  .holo-stat {
    display: flex;
    align-items: baseline;
    gap: var(--mv-spacing-xxs);
    line-height: 1;
  }

  .stat-value {
    font-size: var(--mv-font-size-xl);
    font-weight: var(--mv-font-weight-bold);
  }

  .stat-label {
    font-size: var(--mv-font-size-sm);
    font-weight: var(--mv-font-weight-semibold);
    opacity: 0.8;
    letter-spacing: var(--mv-letter-spacing-count);
  }

  .holo-stat.running {
    color: var(--hud-neon-green);
    text-shadow: var(--mv-text-shadow-green);
  }
  .holo-stat.pending {
    color: var(--hud-neon-yellow);
  }
  .holo-stat.failed {
    color: var(--hud-neon-red);
    text-shadow: var(--mv-text-shadow-purple);
  }
  .holo-stat.idle {
    color: var(--mv-color-text-muted);
  }

  /* Progress Module */
  .progress-module {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-xs);
  }
  .progress-readout {
    font-family: var(--hud-font);
    font-weight: var(--mv-font-weight-semibold);
    font-size: var(--mv-font-size-md);
    color: var(--mv-color-text-secondary);
    min-width: var(--mv-zoom-label-min-width);
    text-align: right;
  }

  /* Segmented Crystal Switch */
  .crystal-switch {
    display: flex;
    padding: var(--mv-spacing-xxxs);
    gap: var(--mv-spacing-xxxs);
    background: var(--mv-glass-bg-darker);
    border-radius: var(--mv-spacing-xl);
    border: var(--mv-border-width-thin) solid var(--mv-glass-active);
  }

  .switch-item {
    display: flex;
    align-items: center;
    justify-content: center;
    width: var(--mv-size-action-btn);
    height: var(--mv-size-action-btn);
    border: none;
    border-radius: var(--mv-radius-full);
    background: transparent;
    color: var(--mv-color-text-muted);
    cursor: pointer;
    transition: all var(--mv-duration-normal) cubic-bezier(0.2, 0.8, 0.2, 1);
  }

  .switch-item:hover {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
  }

  .switch-item.active {
    color: var(--mv-color-text-primary);
    background: var(--mv-glass-active);
    box-shadow: var(--mv-shadow-badge);
    border: var(--mv-border-width-thin) solid var(--mv-glass-hover);
  }

  /* Settings Button */
</style>
