/**
 * WBS（Work Breakdown Structure）ビュー用ストア
 *
 * タスクをマイルストーン・フェーズ別にツリー構造化し、
 * 折りたたみ状態と進捗率を管理
 */

import { writable, derived } from 'svelte/store';
import { tasks } from './taskStore';
import type { Task, PhaseName, TaskStatus } from '../types';

// WBS ツリーノードの型
export interface WBSNode {
  id: string;
  type: 'phase' | 'task';
  label: string;
  phaseName?: PhaseName;
  task?: Task;
  children: WBSNode[];
  level: number;
  progress: {
    completed: number;
    total: number;
    percentage: number;
  };
}

// フェーズの表示順序
const phaseOrder: PhaseName[] = ['概念設計', '実装設計', '実装', '検証', ''];

// フェーズのラベル
// フェーズのラベル
const phaseLabels: Record<PhaseName, string> = {
  概念設計: 'Concept Design',
  実装設計: 'Architecture Design',
  実装: 'Implementation',
  検証: 'Verification',
  '': 'Other',
};

// 折りたたみ状態ストア
function createExpandedStore() {
  const { subscribe, update, set } = writable<Set<string>>(new Set());

  return {
    subscribe,

    // ノードを展開/折りたたみ切り替え
    toggle: (nodeId: string) => {
      update((expanded) => {
        const newSet = new Set(expanded);
        if (newSet.has(nodeId)) {
          newSet.delete(nodeId);
        } else {
          newSet.add(nodeId);
        }
        return newSet;
      });
    },

    // ノードを展開
    expand: (nodeId: string) => {
      update((expanded) => {
        const newSet = new Set(expanded);
        newSet.add(nodeId);
        return newSet;
      });
    },

    // ノードを折りたたむ
    collapse: (nodeId: string) => {
      update((expanded) => {
        const newSet = new Set(expanded);
        newSet.delete(nodeId);
        return newSet;
      });
    },

    // 全て展開
    expandAll: () => {
      // ストア値から全ノードIDを展開
      // 実際にはWBSツリーから全IDを取得する必要があるが、
      // ここではフェーズIDのみを初期展開
      set(new Set(phaseOrder.map((p) => `phase-${p}`)));
    },

    // 全て折りたたむ
    collapseAll: () => {
      set(new Set());
    },

    // 初期状態にリセット（全フェーズを展開）
    reset: () => {
      set(new Set(phaseOrder.map((p) => `phase-${p}`)));
    },
  };
}

export const expandedNodes = createExpandedStore();

// ビューモード（Graph or WBS）
export type ViewMode = 'graph' | 'wbs';

function createViewModeStore() {
  const { subscribe, set } = writable<ViewMode>('graph');

  return {
    subscribe,
    setGraph: () => set('graph'),
    setWBS: () => set('wbs'),
    toggle: () => {
      let current: ViewMode;
      subscribe((v) => (current = v))();
      set(current! === 'graph' ? 'wbs' : 'graph');
    },
  };
}

export const viewMode = createViewModeStore();

// タスクの完了判定
function isTaskCompleted(status: TaskStatus): boolean {
  return status === 'SUCCEEDED' || status === 'COMPLETED' || status === 'CANCELED';
}

// 進捗を計算
function calculateProgress(tasks: Task[]): {
  completed: number;
  total: number;
  percentage: number;
} {
  const total = tasks.length;
  const completed = tasks.filter((t) => isTaskCompleted(t.status)).length;
  const percentage = total > 0 ? Math.round((completed / total) * 100) : 0;
  return { completed, total, percentage };
}

// WBS ツリー構造を生成
export const wbsTree = derived(tasks, ($tasks): WBSNode[] => {
  // フェーズ別にタスクをグループ化
  const tasksByPhase = new Map<PhaseName, Task[]>();

  // 初期化（空の配列を設定）
  for (const phase of phaseOrder) {
    tasksByPhase.set(phase, []);
  }

  // タスクを振り分け
  for (const task of $tasks) {
    const phase = (task.phaseName || '') as PhaseName;
    const phaseKey = phaseOrder.includes(phase) ? phase : '';
    tasksByPhase.get(phaseKey)!.push(task);
  }

  // ツリー構造を生成
  const tree: WBSNode[] = [];

  for (const phase of phaseOrder) {
    const phaseTasks = tasksByPhase.get(phase) || [];

    // タスクがないフェーズはスキップ（'その他' 以外）
    if (phaseTasks.length === 0 && phase !== '') {
      continue;
    }

    // タスクがない場合（'その他' 含む）もスキップ
    if (phaseTasks.length === 0) {
      continue;
    }

    const phaseProgress = calculateProgress(phaseTasks);

    const phaseNode: WBSNode = {
      id: `phase-${phase}`,
      type: 'phase',
      label: phaseLabels[phase],
      phaseName: phase,
      children: phaseTasks.map((task) => ({
        id: task.id,
        type: 'task' as const,
        label: task.title,
        task,
        children: [],
        level: 1,
        progress: {
          completed: isTaskCompleted(task.status) ? 1 : 0,
          total: 1,
          percentage: isTaskCompleted(task.status) ? 100 : 0,
        },
      })),
      level: 0,
      progress: phaseProgress,
    };

    tree.push(phaseNode);
  }

  return tree;
});

// 全体の進捗率
export const overallProgress = derived(tasks, ($tasks) => {
  return calculateProgress($tasks);
});

// フラット化したWBSノードリスト（展開状態を考慮）
export const flattenedWBSNodes = derived(
  [wbsTree, expandedNodes],
  ([$tree, $expanded]): WBSNode[] => {
    const result: WBSNode[] = [];

    function flatten(nodes: WBSNode[], parentExpanded: boolean) {
      for (const node of nodes) {
        if (parentExpanded) {
          result.push(node);
        }

        if (node.children.length > 0) {
          const isExpanded = $expanded.has(node.id);
          flatten(node.children, parentExpanded && isExpanded);
        }
      }
    }

    flatten($tree, true);
    return result;
  }
);
