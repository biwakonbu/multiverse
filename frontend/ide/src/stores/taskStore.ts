/**
 * タスクデータ管理ストア
 *
 * タスク一覧と選択状態をグローバルに管理
 */

import { writable, derived } from 'svelte/store';
import type { Task, TaskNode, TaskStatus } from '../types';
import { grid, gridToCanvas } from '../design-system';

// タスク一覧ストア
function createTasksStore() {
  const { subscribe, set, update } = writable<Task[]>([]);

  return {
    subscribe,

    // タスク一覧を設定
    setTasks: (tasks: Task[]) => set(tasks),

    // タスクを追加
    addTask: (task: Task) => {
      update((tasks) => [...tasks, task]);
    },

    // タスクを更新
    updateTask: (taskId: string, updates: Partial<Task>) => {
      update((tasks) =>
        tasks.map((t) => (t.id === taskId ? { ...t, ...updates } : t))
      );
    },

    // タスクを削除
    removeTask: (taskId: string) => {
      update((tasks) => tasks.filter((t) => t.id !== taskId));
    },

    // クリア
    clear: () => set([]),
  };
}

// 選択状態ストア
function createSelectionStore() {
  const { subscribe, set } = writable<string | null>(null);

  return {
    subscribe,
    select: (taskId: string | null) => set(taskId),
    clear: () => set(null),
  };
}

export const tasks = createTasksStore();
export const selectedTaskId = createSelectionStore();

// グリッド配置されたタスクノード
// タスクを列方向に順番に配置（後で改善可能）
export const taskNodes = derived(tasks, ($tasks): TaskNode[] => {
  const columns = 6; // デフォルト列数

  return $tasks.map((task, index) => ({
    task,
    col: index % columns,
    row: Math.floor(index / columns),
  }));
});

// 選択中のタスク
export const selectedTask = derived(
  [tasks, selectedTaskId],
  ([$tasks, $selectedTaskId]) => {
    if (!$selectedTaskId) return null;
    return $tasks.find((t) => t.id === $selectedTaskId) ?? null;
  }
);

// ステータス別タスク数
export const taskCountsByStatus = derived(tasks, ($tasks) => {
  const counts: Record<TaskStatus, number> = {
    PENDING: 0,
    READY: 0,
    RUNNING: 0,
    SUCCEEDED: 0,
    FAILED: 0,
    CANCELED: 0,
    BLOCKED: 0,
  };

  for (const task of $tasks) {
    counts[task.status]++;
  }

  return counts;
});

// グリッド全体のサイズ（キャンバス座標）
export const gridBounds = derived(taskNodes, ($nodes) => {
  if ($nodes.length === 0) {
    return { width: 0, height: 0 };
  }

  const maxCol = Math.max(...$nodes.map((n) => n.col));
  const maxRow = Math.max(...$nodes.map((n) => n.row));

  const { x: maxX, y: maxY } = gridToCanvas(maxCol, maxRow);

  return {
    width: maxX + grid.cellWidth + grid.gap,
    height: maxY + grid.cellHeight + grid.gap,
  };
});
