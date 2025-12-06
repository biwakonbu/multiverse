# Design System - multiverse IDE

## 責務

multiverse IDE のビジュアル言語を一元管理するデザインシステム。
SF/Sci-Fi 風 Crystal HUD スタイルの視覚的一貫性を保証する。

## テーマ: Nord Deep

**コンセプト**: 深い背景にパステル UI が輝くデザイン

- **ベース**: Nord パレットをベースに深い背景色を拡張
- **UI**: Aurora パレットをパステル化したステータス色
- **グロー**: 控えめな効果（IDE としての実用性重視）
- **スタイル**: SF/Sci-Fi 風の洗練された UI

### Glassmorphism（ガラスモーフィズム）

**Phantom Glass** スタイルを全体に適用:

- **背景**: 半透明ガラス効果（`--mv-glass-bg`: `rgba(22, 24, 30, 0.4)`）
- **ボーダー**: 微細な白いハイライト（`--mv-glass-border-subtle`: `rgba(255, 255, 255, 0.05)`）
- **シャドウ**: アンビエントシャドウ（`--mv-shadow-ambient-lg`）とインナーハイライト
- **ホバー**: 強調されたガラス効果（`--mv-glass-hover`）

### Crystal HUD

ツールバー、パネル、チャットウィンドウなど主要 UI に適用:

- **透明感**: 背景が透けて見える洗練されたパネル
- **グロー効果**: Frost 青系のアクセントグロー（`--mv-shadow-glow-accent`）
- **フォント**: Orbitron（ブランド用）、Rajdhani（ディスプレイ用）
- **レイヤー効果**: 複合シャドウ（`--mv-shadow-glass-panel`）

## 設計思想

### なぜデザインシステムか

- 100 個以上のノードを扱う UI では視覚的一貫性が生命線
- コンポーネント間でスタイルが散らばると保守不能になる
- デザイントークンを一箇所で管理し、変更を容易にする

### トークン階層

```
1. プリミティブカラー    # 生の色値（Nord パレット）
       ↓
2. セマンティックカラー  # 用途別（status-*, surface-*, border-*）
       ↓
3. CSS 変数            # コンポーネントで使用
```

### レイヤー構造

```
tokens/          # 最下層: 生の値（色、サイズ、時間）
  ├── colors.ts        # プリミティブ + セマンティックカラー
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
variables.css    # CSS変数としてのトークン出力（プリミティブ + セマンティック）
```

### Storybook

コンポーネントは全て Storybook で確認可能:

```bash
pnpm storybook    # http://localhost:6006 で起動
```

## デザイントークン体系

### カラートークン

#### プリミティブカラー（生の色値）

Nord パレットをベースに拡張:

| カテゴリ    | 用途                       | 変数プレフィックス             |
| ----------- | -------------------------- | ------------------------------ |
| Polar Night | 深い背景（nord0-3）        | `--mv-primitive-polar-night-*` |
| Snow Storm  | 明るいテキスト（nord4-6）  | `--mv-primitive-snow-storm-*`  |
| Frost       | 青系アクセント（nord7-10） | `--mv-primitive-frost-*`       |
| Aurora      | ステータス色（nord11-15）  | `--mv-primitive-aurora-*`      |
| Deep        | Nord より深い背景（拡張）  | `--mv-primitive-deep-*`        |
| Pastel      | パステル化した Aurora      | `--mv-primitive-pastel-*`      |
| Neutral     | グレースケール             | `--mv-primitive-neutral-*`     |

#### セマンティックカラー（用途で命名）

プリミティブを参照して定義:

- `--mv-color-status-*`: タスクステータス表現
- `--mv-color-surface-*`: 背景・パネル
- `--mv-color-border-*`: 境界線
- `--mv-color-text-*`: テキスト
- `--mv-color-glow-*`: グロー効果
- `--mv-color-shadow-*`: シャドウ
- `--mv-color-interactive-*`: インタラクティブ要素

#### ステータスカラーの原則

| ステータス | カラー                      | 意味           |
| ---------- | --------------------------- | -------------- |
| running    | パステル緑（nord14 ベース） | 活性を表現     |
| succeeded  | 暗い緑                      | 完了の落ち着き |
| failed     | パステル赤（nord11 ベース） | 警告・注意喚起 |
| pending    | グレー/オレンジ             | 控えめな待機   |
| ready      | Frost 青系（nord8/9）       | アクション可能 |
| blocked    | パステル黄（nord13 ベース） | 警告           |
| canceled   | グレー                      | 無効化         |

### スペーシング（Spacing）

**グリッドシステム**:

- `cell-width`: ノードの幅
- `cell-height`: ノードの高さ
- `cell-gap`: ノード間の余白

**UI スペーシング**:

- 4px 単位のスケール（4, 8, 12, 16, 24, 32, 48）

### アニメーション（Animation）

**原則**:

- 実行中のみアニメーション（パルス）
- 状態遷移は控えめなトランジション
- パフォーマンスを考慮（will-change 活用）

**タイミング**:

- 短い: 150ms（ホバー、フォーカス）
- 標準: 300ms（状態遷移）
- 長い: 2000ms（パルスサイクル）

### タイポグラフィ（Typography）

**原則**:

- システムフォント優先（高速読み込み）
- ノード内は最小限の情報（タイトル、ステータス、Pool）
- ズームレベルに応じた表示/非表示

## CSS 変数の命名規則

```css
--mv-{category}-{variant}-{state}

例:
--mv-color-status-running
--mv-color-surface-primary
--mv-spacing-cell-width
--mv-glass-bg
```

`mv-` プレフィックスで multiverse 固有であることを明示。

### 主要な変数カテゴリ

| プレフィックス     | 用途                                         |
| ------------------ | -------------------------------------------- |
| `--mv-primitive-*` | 生の色値（Nord パレット + 拡張）             |
| `--mv-color-*`     | セマンティックカラー（用途別）               |
| `--mv-glass-*`     | ガラスモーフィズム（背景・ボーダー・ホバー） |
| `--mv-shadow-*`    | シャドウ・グロー効果                         |
| `--mv-spacing-*`   | スペーシング                                 |
| `--mv-font-*`      | タイポグラフィ                               |
| `--mv-duration-*`  | アニメーション時間                           |
| `--mv-grid-*`      | グリッドレイアウト                           |
| `--mv-brand-*`     | ブランドコンポーネント                       |

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

- **ハードコード禁止**: 生の色値（#ff0000, rgba(...)等）をコンポーネントに直接書かない
- **トークン階層を守る**: プリミティブ → セマンティック → CSS 変数の流れ
- **セマンティック命名**: 用途で命名（色名ではなく機能名）
- **CSS 変数を使用**: コンポーネントでは `var(--mv-*)` を使用
- **フォールバック不要**: CSS 変数は必ず定義されているためフォールバック値は書かない
- **新カテゴリ追加時**: この CLAUDE.md を更新
