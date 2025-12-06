/**
 * wbsStore のユニットテスト
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import {
  expandedNodes,
  viewMode,
  wbsTree,
  overallProgress,
  flattenedWBSNodes,
} from './wbsStore';
import { tasks } from './taskStore';
import type { Task, PhaseName } from '../types';

// テスト用タスクデータ
function createTask(
  id: string,
  title: string,
  status: Task['status'],
  phaseName: PhaseName = ''
): Task {
  return {
    id,
    title,
    status,
    poolId: 'default',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    phaseName,
    dependencies: [],
  };
}

describe('expandedNodes', () => {
  beforeEach(() => {
    expandedNodes.collapseAll();
  });

  it('toggle で展開/折りたたみを切り替える', () => {
    const nodeId = 'phase-概念設計';

    // 初期状態: 折りたたまれている
    expect(get(expandedNodes).has(nodeId)).toBe(false);

    // 展開
    expandedNodes.toggle(nodeId);
    expect(get(expandedNodes).has(nodeId)).toBe(true);

    // 折りたたむ
    expandedNodes.toggle(nodeId);
    expect(get(expandedNodes).has(nodeId)).toBe(false);
  });

  it('expand で展開する', () => {
    const nodeId = 'phase-実装';

    expandedNodes.expand(nodeId);
    expect(get(expandedNodes).has(nodeId)).toBe(true);
  });

  it('collapse で折りたたむ', () => {
    const nodeId = 'phase-検証';

    expandedNodes.expand(nodeId);
    expandedNodes.collapse(nodeId);
    expect(get(expandedNodes).has(nodeId)).toBe(false);
  });

  it('expandAll で全フェーズを展開する', () => {
    expandedNodes.expandAll();

    const expanded = get(expandedNodes);
    expect(expanded.has('phase-概念設計')).toBe(true);
    expect(expanded.has('phase-実装設計')).toBe(true);
    expect(expanded.has('phase-実装')).toBe(true);
    expect(expanded.has('phase-検証')).toBe(true);
  });

  it('collapseAll で全て折りたたむ', () => {
    expandedNodes.expandAll();
    expandedNodes.collapseAll();

    expect(get(expandedNodes).size).toBe(0);
  });

  it('reset で初期状態（全フェーズ展開）にする', () => {
    expandedNodes.collapseAll();
    expandedNodes.reset();

    const expanded = get(expandedNodes);
    expect(expanded.size).toBeGreaterThan(0);
  });
});

describe('viewMode', () => {
  it('初期値は graph', () => {
    expect(get(viewMode)).toBe('graph');
  });

  it('setGraph で graph に設定', () => {
    viewMode.setWBS();
    viewMode.setGraph();
    expect(get(viewMode)).toBe('graph');
  });

  it('setWBS で wbs に設定', () => {
    viewMode.setWBS();
    expect(get(viewMode)).toBe('wbs');
  });

  it('toggle で切り替え', () => {
    viewMode.setGraph();
    viewMode.toggle();
    expect(get(viewMode)).toBe('wbs');

    viewMode.toggle();
    expect(get(viewMode)).toBe('graph');
  });
});

describe('wbsTree', () => {
  beforeEach(() => {
    tasks.clear();
  });

  it('タスクがない場合は空の配列を返す', () => {
    const tree = get(wbsTree);
    expect(tree).toEqual([]);
  });

  it('タスクをフェーズ別にグループ化する', () => {
    tasks.setTasks([
      createTask('task-1', 'Task 1', 'PENDING', '概念設計'),
      createTask('task-2', 'Task 2', 'RUNNING', '概念設計'),
      createTask('task-3', 'Task 3', 'SUCCEEDED', '実装'),
    ]);

    const tree = get(wbsTree);

    // 2 つのフェーズ
    expect(tree.length).toBe(2);

    // 概念設計フェーズ
    const conceptPhase = tree.find((n) => n.phaseName === '概念設計');
    expect(conceptPhase).toBeDefined();
    expect(conceptPhase?.children.length).toBe(2);

    // 実装フェーズ
    const implPhase = tree.find((n) => n.phaseName === '実装');
    expect(implPhase).toBeDefined();
    expect(implPhase?.children.length).toBe(1);
  });

  it('フェーズノードの進捗率を計算する', () => {
    tasks.setTasks([
      createTask('task-1', 'Task 1', 'SUCCEEDED', '概念設計'),
      createTask('task-2', 'Task 2', 'PENDING', '概念設計'),
    ]);

    const tree = get(wbsTree);
    const conceptPhase = tree.find((n) => n.phaseName === '概念設計');

    expect(conceptPhase?.progress.completed).toBe(1);
    expect(conceptPhase?.progress.total).toBe(2);
    expect(conceptPhase?.progress.percentage).toBe(50);
  });

  it('CANCELED も完了としてカウントする', () => {
    tasks.setTasks([
      createTask('task-1', 'Task 1', 'CANCELED', '実装'),
      createTask('task-2', 'Task 2', 'SUCCEEDED', '実装'),
    ]);

    const tree = get(wbsTree);
    const implPhase = tree.find((n) => n.phaseName === '実装');

    expect(implPhase?.progress.completed).toBe(2);
    expect(implPhase?.progress.percentage).toBe(100);
  });
});

describe('overallProgress', () => {
  beforeEach(() => {
    tasks.clear();
  });

  it('タスクがない場合は 0%', () => {
    const progress = get(overallProgress);
    expect(progress.completed).toBe(0);
    expect(progress.total).toBe(0);
    expect(progress.percentage).toBe(0);
  });

  it('全体の進捗率を計算する', () => {
    tasks.setTasks([
      createTask('task-1', 'Task 1', 'SUCCEEDED', '概念設計'),
      createTask('task-2', 'Task 2', 'SUCCEEDED', '実装'),
      createTask('task-3', 'Task 3', 'PENDING', '検証'),
      createTask('task-4', 'Task 4', 'RUNNING', '検証'),
    ]);

    const progress = get(overallProgress);
    expect(progress.completed).toBe(2);
    expect(progress.total).toBe(4);
    expect(progress.percentage).toBe(50);
  });
});

describe('flattenedWBSNodes', () => {
  beforeEach(() => {
    tasks.clear();
    expandedNodes.collapseAll();
  });

  it('展開状態に応じてノードをフラット化する', () => {
    tasks.setTasks([
      createTask('task-1', 'Task 1', 'PENDING', '概念設計'),
      createTask('task-2', 'Task 2', 'PENDING', '概念設計'),
    ]);

    // 折りたたまれている場合: フェーズのみ表示
    const collapsedNodes = get(flattenedWBSNodes);
    expect(collapsedNodes.length).toBe(1); // フェーズノードのみ

    // 展開した場合: フェーズ + タスク
    expandedNodes.expand('phase-概念設計');
    const expandedNodesResult = get(flattenedWBSNodes);
    expect(expandedNodesResult.length).toBe(3); // フェーズ + 2 タスク
  });
});
