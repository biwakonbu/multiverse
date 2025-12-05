/**
 * アニメーショントークン定義
 *
 * 状態遷移とフィードバックのタイミングを統一。
 */

// デュレーション（ミリ秒）
export const duration = {
  // 即座（ホバー、フォーカス）
  instant: 100,
  // 短い（小さな状態変化）
  fast: 150,
  // 標準（状態遷移）
  normal: 300,
  // 長い（大きな遷移）
  slow: 500,
  // パルスサイクル（実行中アニメーション）
  pulse: 2000,
} as const;

// イージング関数
export const easing = {
  // 標準（滑らかな開始と終了）
  default: 'ease-in-out',
  // 入り（加速）
  in: 'ease-in',
  // 出（減速）
  out: 'ease-out',
  // リニア（一定速度）
  linear: 'linear',
  // スプリング風（弾むような動き）
  spring: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
} as const;

// パルスアニメーション設定（実行中ノード用）
export const pulse = {
  // グロー効果の最大サイズ
  glowSize: 8,
  // グロー効果の色（RGBA）
  glowColor: 'rgba(68, 187, 68, 0.4)',
  // アニメーションサイクル
  duration: duration.pulse,
  easing: easing.default,
} as const;

// ズーム設定
export const zoom = {
  min: 0.25,
  max: 3.0,
  default: 1.0,
  step: 0.1,
  // ホイールズーム時の倍率変化
  wheelFactor: 0.1,
} as const;

// トランジションプリセット
export const transitions = {
  // ホバー効果
  hover: `${duration.fast}ms ${easing.out}`,
  // 状態変化
  state: `${duration.normal}ms ${easing.default}`,
  // 変形（ズーム、パン）
  transform: `${duration.fast}ms ${easing.out}`,
} as const;

export const animationTokens = {
  duration,
  easing,
  pulse,
  zoom,
  transitions,
} as const;
