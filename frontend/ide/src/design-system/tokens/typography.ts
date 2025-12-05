/**
 * タイポグラフィトークン定義
 *
 * フォント、サイズ、ウェイトを統一管理。
 */

// フォントファミリー
export const fontFamily = {
  // システムフォント（高速読み込み）
  sans: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  // モノスペース（コード、ID表示）
  mono: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
} as const;

// フォントサイズ（px）
export const fontSize = {
  xs: 10,
  sm: 12,
  md: 14,
  lg: 16,
  xl: 18,
  xxl: 24,
} as const;

// フォントウェイト
export const fontWeight = {
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
} as const;

// 行の高さ
export const lineHeight = {
  tight: 1.2,
  normal: 1.5,
  relaxed: 1.75,
} as const;

// ノード内テキストスタイル
export const nodeText = {
  // ステータスバッジ
  status: {
    size: fontSize.xs,
    weight: fontWeight.bold,
    lineHeight: lineHeight.tight,
  },
  // タイトル
  title: {
    size: fontSize.sm,
    weight: fontWeight.semibold,
    lineHeight: lineHeight.normal,
  },
  // Pool ID
  pool: {
    size: fontSize.xs,
    weight: fontWeight.normal,
    lineHeight: lineHeight.tight,
  },
} as const;

// ズームレベル別の表示閾値
export const zoomVisibility = {
  // タイトル非表示になるズームレベル
  hideTitle: 0.4,
  // 詳細情報表示されるズームレベル
  showDetails: 1.2,
} as const;

export const typographyTokens = {
  fontFamily,
  fontSize,
  fontWeight,
  lineHeight,
  nodeText,
  zoomVisibility,
} as const;
