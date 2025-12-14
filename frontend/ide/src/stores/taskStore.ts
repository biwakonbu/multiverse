/**
 * タスクデータ管理ストア
 *
 * タスク一覧と選択状態をグローバルに管理
 * Wails Events でリアルタイム更新
 */

import { writable, derived } from 'svelte/store';
import type { Task, TaskNode, TaskStatus, PoolSummary } from '../types';
import { grid, gridToCanvas } from '../design-system';
import { Logger } from '../services/logger';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const log = Logger.withComponent('TaskStore');

// タスク一覧ストア
function createTasksStore() {
  const { subscribe, set, update } = writable<Task[]>([]);

  return {
    subscribe,

    // タスク一覧を設定
    setTasks: (tasks: Task[]) => {
      log.info('tasks updated', { count: tasks.length });
      set(tasks);
    },

    // タスクを追加
    addTask: (task: Task) => {
      update((tasks) => {
        if (tasks.some(t => t.id === task.id)) {
            return tasks;
        }
        log.info('task added', { taskId: task.id, title: task.title });
        return [...tasks, task];
      });
    },

    // タスクを更新
    updateTask: (taskId: string, updates: Partial<Task>) => {
      log.debug('task updated', { taskId, updates });
      update((tasks) =>
        tasks.map((t) => (t.id === taskId ? { ...t, ...updates } : t))
      );
    },

    // タスクを削除
    removeTask: (taskId: string) => {
      log.info('task removed', { taskId });
      update((tasks) => tasks.filter((t) => t.id !== taskId));
    },

    // クリア
    clear: () => {
      log.info('tasks cleared');
      set([]);
    },
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
    COMPLETED: 0,
    FAILED: 0,
    CANCELED: 0,
    BLOCKED: 0,
    RETRY_WAIT: 0,
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

// 依存関係エッジの型
export interface TaskEdge {
  from: string; // 依存元タスクID
  to: string;   // 依存先タスクID
  satisfied: boolean; // 依存が満たされているか
}

// 依存関係エッジのリスト
export const taskEdges = derived(tasks, ($tasks): TaskEdge[] => {
  const edges: TaskEdge[] = [];
  const taskMap = new Map($tasks.map((t) => [t.id, t]));
  const completedStatuses = new Set(['SUCCEEDED', 'COMPLETED', 'CANCELED']);
  const missingDeps: string[] = [];

  for (const task of $tasks) {
    if (!task.dependencies || task.dependencies.length === 0) continue;

    for (const depId of task.dependencies) {
      const depTask = taskMap.get(depId);
      const satisfied = depTask
        ? completedStatuses.has(depTask.status)
        : false;
      if (!depTask) {
        missingDeps.push(depId);
      }

      edges.push({
        from: depId,
        to: task.id,
        satisfied,
      });
    }
  }

  return edges;
});

// ブロックされているタスク（未完了の依存がある）
export const blockedTasks = derived(
  [tasks, taskEdges],
  ([$tasks, $edges]) => {
    const blockedIds = new Set<string>();
    const unsatisfiedEdges = $edges.filter((e) => !e.satisfied);

    for (const edge of unsatisfiedEdges) {
      blockedIds.add(edge.to);
    }

    return $tasks.filter((t) => blockedIds.has(t.id));
  }
);

// 実行可能タスク（PENDINGで、全依存が満たされている）
export const readyTasks = derived(
  [tasks, blockedTasks],
  ([$tasks, $blockedTasks]) => {
    const blockedIds = new Set($blockedTasks.map((t) => t.id));
    return $tasks.filter(
      (t) => t.status === 'PENDING' && !blockedIds.has(t.id)
    );
  }
);

// Pool別サマリストア
function createPoolSummariesStore() {
  const { subscribe, set } = writable<PoolSummary[]>([]);

  return {
    subscribe,
    setSummaries: (summaries: PoolSummary[]) => set(summaries),
    clear: () => set([]),
  };
}

export const poolSummaries = createPoolSummariesStore();

// タスク状態変更イベントの型
interface TaskStateChangeEvent {
  taskId: string;
  oldStatus: TaskStatus;
  newStatus: TaskStatus;
  timestamp: string;
}

// タスク作成イベントの型
interface TaskCreatedEvent {
  task: Task;
}

// Wails イベントリスナー初期化
export function initTaskEvents(): void {
  // task:stateChange イベントをリッスン
  EventsOn('task:stateChange', (event: TaskStateChangeEvent) => {
    log.info('task state changed via event', {
      taskId: event.taskId,
      oldStatus: event.oldStatus,
      newStatus: event.newStatus,
    });
    tasks.updateTask(event.taskId, { status: event.newStatus });
  });

  // task:created イベントをリッスン
  EventsOn('task:created', (event: TaskCreatedEvent) => {
    log.info('task created via event', {
      taskId: event.task.id,
      title: event.task.title,
    });
    tasks.addTask(event.task);
  });

  log.info('task events initialized');

  // process:workerUpdate イベントをリッスン (QH-008: Artifacts)
  EventsOn('process:workerUpdate', (event: ProcessWorkerUpdateEvent) => {
    if (event.artifacts && event.artifacts.length > 0) {
      log.info('worker artifacts received', { taskId: event.taskId, count: event.artifacts.length });
      
      // 既存のTaskを取得してマージするのはストア外では難しいので、
      // updateTaskを利用してartifactsフィールドを更新する。
      // 注意: logsなど他のフィールドがある場合は上書きに注意が必要だが、現状はfilesのみ。
      tasks.updateTask(event.taskId, {
        artifacts: {
          files: event.artifacts
        }
      });
    }
  });
}

// Worker更新イベントの型
interface ProcessWorkerUpdateEvent {
  taskId: string;
  workerId: string;
  status: string;
  command: string;
  exitCode?: number;
  artifacts?: string[];
  timestamp: string;
}
