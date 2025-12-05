/**
 * タスク関連の型定義
 */

// タスクステータス
export type TaskStatus =
  | 'PENDING'
  | 'READY'
  | 'RUNNING'
  | 'SUCCEEDED'
  | 'FAILED'
  | 'CANCELED'
  | 'BLOCKED';

// タスクデータ（Goバックエンドから受け取る形式）
export interface Task {
  id: string;
  title: string;
  status: TaskStatus;
  poolId: string;
  createdAt: string;
  updatedAt: string;
  startedAt?: string;
  doneAt?: string;
}

// グリッド配置用のタスクノード
export interface TaskNode {
  task: Task;
  col: number;
  row: number;
}

// ステータスからCSS変数名サフィックスへの変換
export function statusToCssClass(status: TaskStatus): string {
  return status.toLowerCase();
}

// ステータスの表示名
export const statusLabels: Record<TaskStatus, string> = {
  PENDING: '待機中',
  READY: '準備完了',
  RUNNING: '実行中',
  SUCCEEDED: '成功',
  FAILED: '失敗',
  CANCELED: 'キャンセル',
  BLOCKED: 'ブロック',
};
