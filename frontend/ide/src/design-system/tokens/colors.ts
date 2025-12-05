/**
 * カラートークン定義
 *
 * セマンティックな命名で用途を明確化。
 * 生の色値はここでのみ定義し、他では変数経由で参照。
 */

// ステータスカラー - タスク状態の視覚表現
export const statusColors = {
  // 待機中: 控えめなグレー
  pending: {
    background: '#3a3a3a',
    border: '#666666',
    text: '#ff9800',
  },
  // 準備完了: アクション可能を示す青
  ready: {
    background: '#2a2a4a',
    border: '#5588ff',
    text: '#5588ff',
  },
  // 実行中: 活性を示す緑
  running: {
    background: '#2a3a2a',
    border: '#44bb44',
    text: '#4caf50',
  },
  // 成功: 完了の落ち着いた緑
  succeeded: {
    background: '#1a3a1a',
    border: '#228822',
    text: '#228822',
  },
  // 失敗: 警告の赤
  failed: {
    background: '#3a1a1a',
    border: '#cc2222',
    text: '#f44336',
  },
  // キャンセル: 無効化のグレー
  canceled: {
    background: '#2a2a2a',
    border: '#555555',
    text: '#888888',
  },
  // ブロック: 警告の黄
  blocked: {
    background: '#3a3a1a',
    border: '#ccaa22',
    text: '#ccaa22',
  },
} as const;

// サーフェスカラー - 背景・パネル
export const surfaceColors = {
  // アプリケーション背景
  app: '#1a1a1a',
  // プライマリパネル
  primary: '#1e1e1e',
  // セカンダリパネル（サイドバー等）
  secondary: '#252525',
  // ホバー状態
  hover: '#2a2a2a',
  // 選択状態
  selected: '#333333',
  // ノードのデフォルト背景
  node: '#2d2d2d',
} as const;

// ボーダーカラー
export const borderColors = {
  subtle: '#333333',
  default: '#444444',
  strong: '#666666',
  focus: '#4caf50',
} as const;

// テキストカラー
export const textColors = {
  primary: '#eeeeee',
  secondary: '#aaaaaa',
  muted: '#888888',
  disabled: '#555555',
} as const;

// 全カラートークンをエクスポート
export const colors = {
  status: statusColors,
  surface: surfaceColors,
  border: borderColors,
  text: textColors,
} as const;

export type StatusKey = keyof typeof statusColors;
export type SurfaceKey = keyof typeof surfaceColors;
export type BorderKey = keyof typeof borderColors;
export type TextKey = keyof typeof textColors;
