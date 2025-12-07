<script lang="ts">
  import { run } from 'svelte/legacy';

  import WBSListView from "./WBSListView.svelte";
  import { tasks } from "../../stores/taskStore";
  import { expandedNodes } from "../../stores/wbsStore";
  import type { Task } from "../../types";
  import type { PhaseName } from "../../schemas";
  import { onMount, onDestroy } from "svelte";

  
  interface Props {
    // Props
    taskCount?: number;
    completedRatio?: number;
    showAllPhases?: boolean;
  }

  let { taskCount = 5, completedRatio = 0.4, showAllPhases = true }: Props = $props();

  // サンプルタスクを生成
  function generateTasks(count: number, ratio: number): Task[] {
    const phases: PhaseName[] = showAllPhases
      ? ["概念設計", "実装設計", "実装", "検証"]
      : ["実装"];
    const statuses: Task["status"][] = [
      "PENDING",
      "READY",
      "RUNNING",
      "SUCCEEDED",
      "FAILED",
    ];

    const result: Task[] = [];
    const completedCount = Math.floor(count * ratio);

    for (let i = 0; i < count; i++) {
      const phase = phases[i % phases.length];
      let status: Task["status"];

      if (i < completedCount) {
        status = "SUCCEEDED";
      } else if (i === completedCount) {
        status = "RUNNING";
      } else {
        status = statuses[i % 3] as Task["status"]; // PENDING, READY, RUNNING
      }

      result.push({
        id: `task-${i + 1}`,
        title: `タスク ${i + 1}: ${phase}のサンプル`,
        status,
        poolId: "default",
        phaseName: phase,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
        dependencies: [],
      });
    }

    return result;
  }

  // マウント時にタスクを設定
  onMount(() => {
    const sampleTasks = generateTasks(taskCount, completedRatio);
    tasks.setTasks(sampleTasks);
    expandedNodes.expandAll();
  });

  // アンマウント時にクリア
  onDestroy(() => {
    tasks.clear();
    expandedNodes.collapseAll();
  });

  // Props変更時に再生成
  run(() => {
    const sampleTasks = generateTasks(taskCount, completedRatio);
    tasks.setTasks(sampleTasks);
  });
</script>

<div class="preview-container">
  <WBSListView />
</div>

<style>
  .preview-container {
    width: var(--mv-space-500, 500px);
    height: var(--mv-space-600, 600px);
    background: var(--mv-color-surface-primary);
    border-radius: var(--mv-radius-md, 8px);
    overflow: hidden;
  }
</style>
