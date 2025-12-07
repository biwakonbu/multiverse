/**
 * カラートークン定義
 *
 * テーマ: Nord Deep ベース
 * - 深い背景にパステル UI が輝くデザイン
 * - 控えめなグロー効果
 * - ゲーム的な UI（Factorio 風）
 *
 * トークン階層:
 * 1. プリミティブカラー（生の色値）
 * 2. セマンティックカラー（用途別）
 *
 * セマンティックな命名で用途を明確化。
 * CSS 変数と完全に連動。
 */

// ========================================
// プリミティブカラー: Nord パレット
// ========================================

// Polar Night（深い背景）
export const polarNight = {
  nord0: '#2e3440',
  nord1: '#3b4252',
  nord2: '#434c5e',
  nord3: '#4c566a',
} as const;

// Snow Storm（明るいテキスト）
export const snowStorm = {
  nord4: '#d8dee9',
  nord5: '#e5e9f0',
  nord6: '#eceff4',
} as const;

// Frost（青系アクセント）
export const frost = {
  nord7: '#8fbcbb',
  nord8: '#88c0d0',
  nord9: '#81a1c1',
  nord10: '#5e81ac',
} as const;

// Aurora（ステータス色）
export const aurora = {
  red: '#bf616a',
  orange: '#d08770',
  yellow: '#ebcb8b',
  green: '#a3be8c',
  purple: '#b48ead',
} as const;

// Nord Deep 拡張（標準より深い背景）
export const deep = {
  0: '#16181e',
  1: '#1c1f26',
  2: '#242832',
  3: '#272b36',
  4: '#2d323e',
  5: '#363c4a',
} as const;

// パステル化した Aurora
export const pastel = {
  red: '#c97b7b',
  orange: '#d19a66',
  yellow: '#d9c28a',
  green: '#8fbf9f',
  greenMuted: '#7fa387',
  greenDark: '#6b8f71',
} as const;

// ニュートラル（グレースケール）
export const neutral = {
  400: '#6b7280',
  500: '#5c6677',
  600: '#a5afbf',
} as const;

// 全プリミティブカラーをエクスポート
export const primitiveColors = {
  polarNight,
  snowStorm,
  frost,
  aurora,
  deep,
  pastel,
  neutral,
} as const;

// ========================================
// セマンティックカラー: 用途別
// ========================================

// ステータスカラー - タスク状態の視覚表現
export const statusColors = {
  pending: {
    background: deep[4],
    border: polarNight.nord3,
    text: pastel.orange,
  },
  ready: {
    background: '#1e2a3a',
    border: frost.nord9,
    text: frost.nord8,
  },
  running: {
    background: '#1e2d26',
    border: pastel.green,
    text: aurora.green,
  },
  succeeded: {
    background: '#1a2822',
    border: pastel.greenDark,
    text: pastel.greenMuted,
  },
  failed: {
    background: '#2d1f22',
    border: pastel.red,
    text: aurora.red,
  },
  canceled: {
    background: '#22252d',
    border: polarNight.nord1,
    text: neutral[400],
  },
  blocked: {
    background: '#2d2a1e',
    border: pastel.yellow,
    text: aurora.yellow,
  },
  completed: {
    background: '#2d261a',
    border: aurora.yellow,
    text: aurora.yellow,
  },
  retryWait: {
    background: '#2d2420',
    border: aurora.orange,
    text: aurora.orange,
  },
} as const;

// サーフェスカラー - 背景・パネル
export const surfaceColors = {
  app: deep[0],
  primary: deep[1],
  secondary: deep[2],
  hover: deep[4],
  selected: deep[5],
  node: deep[3],
  overlay: 'rgba(0, 0, 0, 0.6)',
} as const;

// ボーダーカラー
export const borderColors = {
  subtle: polarNight.nord1,
  default: polarNight.nord2,
  strong: polarNight.nord3,
  focus: frost.nord8,
} as const;

// テキストカラー
export const textColors = {
  primary: snowStorm.nord6,
  secondary: snowStorm.nord4,
  muted: neutral[600],
  disabled: neutral[500],
} as const;

// グロー・シャドウカラー
export const glowColors = {
  focus: 'rgba(136, 192, 208, 0.3)',
  selected: 'rgba(136, 192, 208, 0.4)',
  error: 'rgba(191, 97, 106, 0.3)',
  running: 'rgba(143, 191, 159, 0.35)',
} as const;

// シャドウカラー
export const shadowColors = {
  elevated: 'rgba(0, 0, 0, 0.3)',
  deep: 'rgba(0, 0, 0, 0.4)',
} as const;

// インタラクティブカラー
export const interactiveColors = {
  primary: frost.nord8,
  primaryHover: frost.nord7,
  secondary: frost.nord9,
  danger: aurora.red,
} as const;

// 全セマンティックカラーをエクスポート
export const semanticColors = {
  status: statusColors,
  surface: surfaceColors,
  border: borderColors,
  text: textColors,
  glow: glowColors,
  shadow: shadowColors,
  interactive: interactiveColors,
} as const;

// 後方互換性のため colors もエクスポート
export const colors = semanticColors;

// 型定義
export type PolarNightKey = keyof typeof polarNight;
export type SnowStormKey = keyof typeof snowStorm;
export type FrostKey = keyof typeof frost;
export type AuroraKey = keyof typeof aurora;
export type DeepKey = keyof typeof deep;
export type PastelKey = keyof typeof pastel;
export type NeutralKey = keyof typeof neutral;
export type StatusKey = keyof typeof statusColors;
export type SurfaceKey = keyof typeof surfaceColors;
export type BorderKey = keyof typeof borderColors;
export type TextKey = keyof typeof textColors;
export type GlowKey = keyof typeof glowColors;
export type ShadowKey = keyof typeof shadowColors;
export type InteractiveKey = keyof typeof interactiveColors;
