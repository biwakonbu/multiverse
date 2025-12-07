<script lang="ts">
  import { run } from 'svelte/legacy';

  import WBSGraphView from "./WBSGraphView.svelte";
  import { tasks } from "../../stores/taskStore";
  import type { Task } from "../../types";

  type PhaseName = "概念設計" | "実装設計" | "実装" | "検証";

  interface Props {
    taskCount?: number;
    completedRatio?: number;
    showAllPhases?: boolean;
  }

  let { taskCount = 5, completedRatio = 0.4, showAllPhases = true }: Props = $props();

  const phases: PhaseName[] = ["概念設計", "実装設計", "実装", "検証"];
  const statuses: Task["status"][] = [
    "PENDING",
    "RUNNING",
    "SUCCEEDED",
    "FAILED",
  ];

  // サンプルタスクを生成
  function generateSampleTasks(count: number, completedRatio: number): Task[] {
    const result: Task[] = [];
    const completedCount = Math.floor(count * completedRatio);

    for (let i = 0; i < count; i++) {
      const phase = showAllPhases ? phases[i % phases.length] : phases[0];

      const isCompleted = i < completedCount;
      const status: Task["status"] = isCompleted
        ? "SUCCEEDED"
        : statuses[i % 3]; // PENDING, RUNNING, SUCCEEDED をローテーション

      result.push({
        id: `task-${i}`,
        title: `サンプルタスク ${i + 1}`,
        status,
        poolId: "default",
        phaseName: phase,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      });
    }

    return result;
  }

  // タスクストアを更新
  run(() => {
    const sampleTasks = generateSampleTasks(taskCount, completedRatio);
    tasks.setTasks(sampleTasks);
  });
</script>

<div class="preview-container">
  <WBSGraphView />
</div>

<style>
  .preview-container {
    width: var(--mv-size-preview-width);
    height: var(--mv-size-preview-height);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-md);
    overflow: hidden;
    border: var(--mv-border-width-sm) solid var(--mv-color-border-subtle);
  }
</style>
