/**
 * Zod スキーマ定義
 *
 * Wails が生成する any 型のバインディングを
 * ランタイムで検証し、型安全な値に変換する
 */

import { z } from 'zod';

// TaskStatus スキーマ
export const TaskStatusSchema = z.enum([
  'PENDING',
  'READY',
  'RUNNING',
  'SUCCEEDED',
  'COMPLETED',
  'FAILED',
  'CANCELED',
  'BLOCKED',
  'RETRY_WAIT',
]);

export type TaskStatus = z.infer<typeof TaskStatusSchema>;

// PhaseName スキーマ（タスク分解フェーズ）
export const PhaseNameSchema = z.enum([
  '概念設計',
  '実装設計',
  '実装',
  '検証',
  '',
]);

export type PhaseName = z.infer<typeof PhaseNameSchema>;

// フェーズ名からCSSクラス名への変換
export function phaseToCssClass(phase: PhaseName | undefined | null): string {
  if (!phase) return '';
  const phaseMap: Record<string, string> = {
    概念設計: 'phase-concept',
    実装設計: 'phase-design',
    実装: 'phase-impl',
    検証: 'phase-verify',
  };
  return phaseMap[phase] || '';
}

// Task スキーマ
export const TaskSchema = z.object({
  // 基本フィールド
  id: z.string(),
  title: z.string(),
  status: TaskStatusSchema,
  poolId: z.string(),
  createdAt: z.string().datetime({ offset: true }).or(z.string()),
  updatedAt: z.string().datetime({ offset: true }).or(z.string()),
  startedAt: z.string().datetime({ offset: true }).or(z.string()).optional(),
  doneAt: z.string().datetime({ offset: true }).or(z.string()).optional(),

  // v2.0 拡張フィールド
  description: z.string().optional(),
  dependencies: z.array(z.string()).optional(),
  parentId: z.string().optional().nullable(),
  wbsLevel: z.number().int().nonnegative().optional(),
  phaseName: PhaseNameSchema.optional(),
  milestone: z.string().optional(),
  sourceChatId: z.string().optional().nullable(),
  acceptanceCriteria: z.array(z.string()).optional(),

  // リトライ管理用 (v2.0 Extension)
  attemptCount: z.number().int().nonnegative().optional(),
  nextRetryAt: z.string().datetime({ offset: true }).or(z.string()).optional().nullable(),

  // Phase 1: Data Model Enhancements
  suggestedImpl: z.object({
    language: z.string().optional(),
    filePaths: z.array(z.string()).optional(),
    constraints: z.array(z.string()).optional(),
  }).optional(),

  artifacts: z.object({
    files: z.array(z.string()).optional(),
    logs: z.array(z.string()).optional(),
  }).optional(),
});

export type Task = z.infer<typeof TaskSchema>;

// Task 配列スキーマ
export const TaskArraySchema = z.array(TaskSchema);

// Workspace スキーマ
export const WorkspaceSchema = z.object({
  version: z.string(),
  projectRoot: z.string(),
  displayName: z.string(),
  createdAt: z.string().datetime({ offset: true }).or(z.string()),
  lastOpenedAt: z.string().datetime({ offset: true }).or(z.string()),
});

export type Workspace = z.infer<typeof WorkspaceSchema>;

// WorkspaceSummary スキーマ（一覧表示用）
export const WorkspaceSummarySchema = z.object({
  id: z.string(),
  displayName: z.string(),
  projectRoot: z.string(),
  lastOpenedAt: z.string().datetime({ offset: true }).or(z.string()),
});

export type WorkspaceSummary = z.infer<typeof WorkspaceSummarySchema>;

// グリッド配置用のタスクノード
export const TaskNodeSchema = z.object({
  task: TaskSchema,
  col: z.number().int().nonnegative(),
  row: z.number().int().nonnegative(),
});

export type TaskNode = z.infer<typeof TaskNodeSchema>;

// ステータスからCSS変数名サフィックスへの変換（kebab-case）
export function statusToCssClass(status: TaskStatus): string {
  return status.toLowerCase().replace(/_/g, '-');
}

// ステータスの表示名
export const statusLabels: Record<TaskStatus, string> = {
  PENDING: '待機中',
  READY: '準備完了',
  RUNNING: '実行中',
  SUCCEEDED: '処理成功',
  COMPLETED: '完了',
  FAILED: '失敗',
  CANCELED: 'キャンセル',
  BLOCKED: 'ブロック',
  RETRY_WAIT: 'リトライ待機',
};

// AttemptStatus スキーマ
export const AttemptStatusSchema = z.enum([
  'STARTING',
  'RUNNING',
  'SUCCEEDED',
  'FAILED',
  'TIMEOUT',
  'CANCELED',
]);

export type AttemptStatus = z.infer<typeof AttemptStatusSchema>;

// Attempt スキーマ
export const AttemptSchema = z.object({
  id: z.string(),
  taskId: z.string(),
  status: AttemptStatusSchema,
  startedAt: z.string().datetime({ offset: true }).or(z.string()),
  finishedAt: z.string().datetime({ offset: true }).or(z.string()).optional(),
  errorSummary: z.string().optional(),
});

export type Attempt = z.infer<typeof AttemptSchema>;

// AttemptStatusの表示名
export const attemptStatusLabels: Record<AttemptStatus, string> = {
  STARTING: '開始中',
  RUNNING: '実行中',
  SUCCEEDED: '成功',
  FAILED: '失敗',
  TIMEOUT: 'タイムアウト',
  CANCELED: 'キャンセル',
};

// PoolSummary スキーマ
export const PoolSummarySchema = z.object({
  poolId: z.string(),
  running: z.number(),
  queued: z.number(),
  failed: z.number(),
  total: z.number(),
  counts: z.record(z.string(), z.number()),
});

export type PoolSummary = z.infer<typeof PoolSummarySchema>;
