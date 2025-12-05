# Design System - multiverse IDE

## 責務

multiverse IDE のビジュアル言語を一元管理するデザインシステム。
Factorio風2D俯瞰UIの視覚的一貫性を保証する。

## 設計思想

### なぜデザインシステムか

- 100個以上のノードを扱うUIでは視覚的一貫性が生命線
- コンポーネント間でスタイルが散らばると保守不能になる
- デザイントークンを一箇所で管理し、変更を容易にする

### レイヤー構造

```
tokens/          # 最下層: 生の値（色、サイズ、時間）
  ├── colors.ts
  ├── spacing.ts
  ├── typography.ts
  └── animation.ts

components/      # 上層: トークンを使った再利用可能コンポーネント
  ├── Button.svelte      # ボタン（4バリアント: primary/secondary/ghost/danger）
  ├── Badge.svelte       # ステータスバッジ（7ステータス対応）
  ├── Card.svelte        # カード（3バリアント: default/elevated/outlined）
  ├── Input.svelte       # テキスト入力（ラベル、エラー状態対応）
  └── index.ts           # エクスポート集約

index.ts         # エクスポート集約
variables.css    # CSS変数としてのトークン出力
```

### Storybook

コンポーネントは全て Storybook で確認可能:

```bash
pnpm storybook    # http://localhost:6006 で起動
```

## デザイントークン体系

### 色（Colors）

**セマンティックカラー**（用途で命名）:
- `status-*`: タスクステータス表現
- `surface-*`: 背景・パネル
- `border-*`: 境界線
- `text-*`: テキスト

**ステータスカラーの原則**:
- 実行中（running）: 緑系 - 活性を表現
- 成功（succeeded）: 暗い緑系 - 完了の落ち着き
- 失敗（failed）: 赤系 - 警告・注意喚起
- 待機（pending）: グレー系 - 控えめ
- 準備完了（ready）: 青系 - アクション可能

### スペーシング（Spacing）

**グリッドシステム**:
- `cell-width`: ノードの幅
- `cell-height`: ノードの高さ
- `cell-gap`: ノード間の余白

**UIスペーシング**:
- 4px単位のスケール（4, 8, 12, 16, 24, 32, 48）

### アニメーション（Animation）

**原則**:
- 実行中のみアニメーション（パルス）
- 状態遷移は控えめなトランジション
- パフォーマンスを考慮（will-change活用）

**タイミング**:
- 短い: 150ms（ホバー、フォーカス）
- 標準: 300ms（状態遷移）
- 長い: 2000ms（パルスサイクル）

### タイポグラフィ（Typography）

**原則**:
- システムフォント優先（高速読み込み）
- ノード内は最小限の情報（タイトル、ステータス、Pool）
- ズームレベルに応じた表示/非表示

## CSS変数の命名規則

```css
--mv-{category}-{variant}-{state}

例:
--mv-color-status-running
--mv-color-surface-primary
--mv-spacing-cell-width
--mv-animation-pulse-duration
```

`mv-` プレフィックスで multiverse 固有であることを明示。

## 使用方法

```svelte
<script>
  import { colors, spacing } from '$design-system';
</script>

<style>
  .node {
    background: var(--mv-color-surface-node);
    width: var(--mv-spacing-cell-width);
  }
</style>
```

## 拡張時の注意

- トークン追加時は必ずセマンティックな命名を使う
- 生の値（#ff0000等）を直接使わない
- 新しいカテゴリ追加時はこのCLAUDE.mdを更新
