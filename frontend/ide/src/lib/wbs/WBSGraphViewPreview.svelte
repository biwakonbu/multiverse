<script lang="ts">
  import WBSGraphView from "./WBSGraphView.svelte";
  import { tasks } from "../../stores/taskStore";
  import type { Task, PhaseName } from "../../types";

  export let taskCount = 5;
  export let completedRatio = 0.4;
  export let showAllPhases = true;

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
  $: {
    const sampleTasks = generateSampleTasks(taskCount, completedRatio);
    tasks.setTasks(sampleTasks);
  }
</script>

<div class="preview-container">
  <WBSGraphView />
</div>

<style>
  .preview-container {
    width: var(--mv-size-preview-width);
    height: var(--mv-size-preview-height);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-md, 8px);
    overflow: hidden;
  }
</style>
