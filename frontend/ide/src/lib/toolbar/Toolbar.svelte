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
  import ExecutionControls from "./ExecutionControls.svelte";
  import { Network, ListTree } from "lucide-svelte";
  import Panel from "../../design-system/components/Panel.svelte";
  import Flex from "../../design-system/components/Flex.svelte";

  // Badge status type (no longer directly used for badges, but conceptually for styling)
  type BadgeStatus =
    | "pending"
    | "ready"
    | "running"
    | "succeeded"
    | "failed"
    | "canceled"
    | "blocked";

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
  $: hasPoolSummaries = $poolSummaries.length > 0;
  $: isGraphMode = $viewMode === "graph";
</script>

<Panel variant="glass" padding="none" class="toolbar-panel">
  <div class="toolbar crystal-hud">
    <!-- 左側：ブランド -->
    <Flex align="center" gap="var(--mv-spacing-lg)">
      <BrandText size="sm" />
    </Flex>

    <!-- 中央：Holographic Data Strip -->
    <Flex align="center" justify="center" grow>
      <Panel
        variant="glass-strong"
        padding="sm"
        radius="full"
        class="holographic-strip-panel"
      >
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
      </Panel>
    </Flex>

    <!-- 右側：Command Capsule & View Switch -->
    <Flex align="center" justify="end" gap="var(--mv-spacing-lg)">
      <!-- Command Capsule -->
      <Panel variant="glass" padding="none" radius="md" class="command-capsule">
        <ExecutionControls />
      </Panel>

      <!-- Progress -->
      <div class="progress-module">
        <ProgressBar percentage={$overallProgress.percentage} size="mini" />
        <span class="progress-readout">{$overallProgress.percentage}%</span>
      </div>

      <!-- Segmented Crystal Switch -->
      <Panel
        variant="glass-strong"
        padding="none"
        radius="full"
        class="crystal-switch"
      >
        <button
          class="switch-item"
          class:active={isGraphMode}
          on:click={() => viewMode.setGraph()}
          title="Graph View"
        >
          <Network size="16" />
        </button>
        <button
          class="switch-item"
          class:active={!isGraphMode}
          on:click={() => viewMode.setWBS()}
          title="WBS View"
        >
          <ListTree size="16" />
        </button>
      </Panel>
    </Flex>
  </div>
</Panel>

<style>
  :global(:root) {
    --hud-font: "Rajdhani", sans-serif;
    --hud-neon-blue: var(--mv-primitive-frost-1);
    --hud-neon-green: var(--mv-primitive-aurora-green);
    --hud-neon-red: var(--mv-primitive-aurora-red);
    --hud-neon-yellow: var(--mv-primitive-aurora-yellow);
  }

  :global(.toolbar-panel) {
    width: 100%;
    z-index: 100;
  }

  /* Override Panel padding for specific layout needs if necessary, 
     but we used padding="none" on the panel and added internal div for structure 
  */
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: var(--mv-layout-toolbar-height);
    padding: 0 var(--mv-spacing-xl);
    width: 100%;
    box-sizing: border-box;
    font-family: var(--hud-font);
  }

  /* Holographic Data Strip */
  .holographic-strip {
    display: flex;
    align-items: center;
    gap: var(--mv-spacing-md);
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
  :global(.command-capsule) {
    display: flex;
    align-items: center;
    padding: var(--mv-spacing-xxxs);
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
  :global(.crystal-switch) {
    display: flex;
    padding: var(--mv-spacing-xxxs);
    gap: var(--mv-spacing-xxxs);
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
