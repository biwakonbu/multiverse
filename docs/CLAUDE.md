# docs/CLAUDE.md - ドキュメント整理ガイド

このファイルはドキュメント整理・命名規則・拡張に関するメモリです。

## ドキュメント整理方針

### 配置原則

```
根ルート (AI 向けルール非対象)：
  ├── CLAUDE.md          # プロジェクトメモリ（AI開発操作ガイド）
  ├── GEMINI.md          # プロジェクト概要・背景（変更不可）
  ├── README.md          # エンドユーザー向けドキュメント（変更不可）
  ├── TESTING.md         # テストベストプラクティス
  ├── CODEX_TEST_README.md # Codex統合ガイド
  └── docs/              # 設計・仕様ドキュメント（プロダクト指向）
        └── CLAUDE.md    # このファイル（ドキュメント整理ルール）

コア実装層メモリ（各内部パッケージ）：
  internal/{core,meta,worker,note,mock}/ → CLAUDE.md
  pkg/config/                            → CLAUDE.md
  test/                                  → CLAUDE.md
```

### 責任分担

| ドキュメント              | 対象読者                 | 更新主体   | 用途                               |
| ------------------------- | ------------------------ | ---------- | ---------------------------------- |
| **CLAUDE.md** (root)      | Claude Code              | AI 開発者  | プロジェクト全体・開発操作ガイド   |
| **docs/CLAUDE.md**        | Claude Code              | AI 開発者  | ドキュメント体系・メンテナンス規則 |
| **docs/\*.md**            | アーキテクト・レビュアー | 人間開発者 | 設計・仕様・ユースケース           |
| **internal/\*/CLAUDE.md** | Claude Code              | AI 開発者  | パッケージ固有の設計・実装パターン |
| **GEMINI.md / README.md** | エンドユーザー           | 人間開発者 | プロジェクト概要（変更禁止）       |
| **TESTING.md**            | 開発者・テスター         | 人間開発者 | テストベストプラクティス           |

## ドキュメント命名規則

### docs/ ドキュメント

**設計・仕様・実装層**:

```
docs/{scope}-{target}-{version}.md

scope    : AgentRunner | Component | Design | Guide
target   : architecture | spec | impl-design | protocol
version  : v1, v2... (オプション、仕様に応じて)

例：
  - AgentRunner-architecture.md        # 全体アーキテクチャ
  - agentrunner-spec-v1.md             # MVP仕様書（v1）
  - AgentRunner-impl-design-v1.md      # Go実装設計（v1）
  - design-protocol-meta-agent.md      # Meta-agentプロトコル（例：将来拡張）
```

**開発ガイド層**:

```
docs/{topic}_*README.md または docs/{topic}.md

例：
  - TESTING.md                         # テストベストプラクティス（単一・更新頻繁）
  - CODEX_TEST_README.md               # Codex統合テスト実行ガイド
  - SANDBOX_GUIDE.md                   # Docker Sandbox操作ガイド（例：将来）
```

### 内部 CLAUDE.md

```
{package}/CLAUDE.md

常に同じ構成・項目で統一：
  - パッケージ概要（1-2行）
  - 主要概念・設計パターン
  - 実装ガイダンス
  - 拡張・カスタマイズ方法
  - テスト戦略
  - 既知問題・制約
```

## ドキュメント体系

### docs/ 配下の分類

```
docs/
├── README.md                            # ドキュメント全体のナビゲーション
├── CLAUDE.md                            # ドキュメント整理ルール・命名規則・責務分担
│
├── specifications/                      # 仕様ドキュメント（確定仕様）
│   ├── README.md                        # 仕様ドキュメントの概要
│   ├── core-specification.md            # コア仕様（YAML、プロトコル、FSM）
│   ├── meta-protocol.md                 # Meta-agentプロトコル仕様
│   └── worker-interface.md              # Worker実行仕様
│
├── design/                              # 設計ドキュメント
│   ├── README.md                        # 設計ドキュメントの概要
│   ├── architecture.md                  # システムアーキテクチャ
│   ├── implementation-guide.md          # 実装ガイド（Go固有）
│   └── data-flow.md                     # データフロー設計
│
└── guides/                              # 開発ガイド
    ├── README.md                        # ガイドの概要
    ├── testing.md                       # テストガイド
    └── codex-integration.md             # Codex統合ガイド
```

### 各層の役割

| レイヤー         | ディレクトリ      | 内容                                                    | 対象                 |
| ---------------- | ----------------- | ------------------------------------------------------- | -------------------- |
| **メモリ・管理** | `CLAUDE.md`       | ドキュメント整理ルール、命名規則、責務分担              | AI 開発者            |
| **仕様**         | `specifications/` | 確定仕様（YAML スキーマ、プロトコル、インターフェース） | 実装者・レビュアー   |
| **設計**         | `design/`         | システム設計、アーキテクチャ、実装方針                  | アーキテクト・実装者 |
| **ガイド**       | `guides/`         | テスト手法、統合ガイド、操作手順                        | テスター・開発者     |

## サブディレクトリ CLAUDE.md の標準パターン

### 構成テンプレート

```markdown
# {package}/CLAUDE.md - {パッケージ簡潔説明}

## パッケージ概要

1 行で責務を明記：「{モジュール名}は{役割}を{方法}で実現します」

## 主要概念

### 概念 A

- 定義と用途
- 関連インターフェース・型

### 概念 B

- 定義と用途

## 実装パターン

### パターン 1: {パターン名}

- 用途・シーン
- コード例
- 長所・短所

## テスト戦略

- テスト対象と方法
- モック戦略
- カバレッジ目標

## 拡張・カスタマイズ

- インターフェース追加時の手順
- 新機能追加のガイドライン

## 既知問題・制約

- 既知問題と回避策
- パフォーマンス制約
- 将来改善予定
```

### 標準化項目（全パッケージ共通）

各 internal/\*/CLAUDE.md は以下を必須含む：

1. **モジュール責務**：1-2 行で明確化
2. **主要インターフェース**：`type Interface struct { ... }`
3. **依存性図**：どのパッケージに依存するか
4. **テスト対象**：どのテストで検証するか
5. **拡張ポイント**：インターフェース、オプション設定
6. **既知課題**：実装上の制約

## ドキュメント更新ルール

### 追加・変更トリガー

| イベント               | アクション                           | 対象ドキュメント                      |
| ---------------------- | ------------------------------------ | ------------------------------------- |
| **新機能追加**         | 仕様書と実装設計を更新               | docs/spec, docs/impl-design           |
| **アーキテクチャ変更** | アーキテクチャドキュメント更新       | docs/architecture                     |
| **パッケージ追加**     | 新規 CLAUDE.md 作成・root 更新       | new internal/\*/CLAUDE.md + CLAUDE.md |
| **バグ修正・改善**     | 既知問題セクション更新               | internal/\*/CLAUDE.md                 |
| **テスト追加**         | テスト戦略セクション更新             | test/CLAUDE.md                        |
| **バージョンアップ**   | 新規 spec-v\*.md と Impl-design 更新 | docs/spec-vX, docs/impl-design-vX     |

### メンテナンス周期

- **即座更新**：仕様変更、API シグネチャ変更
- **定期更新**（リリース前）：パフォーマンス測定、既知問題整理
- **年 1 回見直し**：全体構成・命名規則・廃止項目確認

## 既存ドキュメント一覧

| ファイル                                                                     | 管理者           | 更新頻度               |
| ---------------------------------------------------------------------------- | ---------------- | ---------------------- |
| **CLAUDE.md** (このファイル)                                                 | AI 開発者        | 命名規則変更時         |
| [README.md](README.md)                                                       | AI 開発者        | ドキュメント構造変更時 |
| [specifications/core-specification.md](specifications/core-specification.md) | 人間・プロダクト | 仕様変更時             |
| [specifications/meta-protocol.md](specifications/meta-protocol.md)           | 人間・プロダクト | プロトコル変更時       |
| [specifications/worker-interface.md](specifications/worker-interface.md)     | 人間・プロダクト | Worker 仕様変更時      |
| [design/architecture.md](design/architecture.md)                             | 人間設計者       | 年 1 回以上            |
| [design/implementation-guide.md](design/implementation-guide.md)             | 人間実装者       | 重大リファクタリング時 |
| [design/data-flow.md](design/data-flow.md)                                   | 人間実装者       | データフロー変更時     |
| [guides/testing.md](guides/testing.md)                                       | 開発者           | テスト手法変更時       |
| [guides/codex-integration.md](guides/codex-integration.md)                   | 開発者           | 実行手順変更時         |
| [../CLAUDE.md](../CLAUDE.md)                                                 | AI 開発者        | 継続的更新             |
| [../test/CLAUDE.md](../test/CLAUDE.md)                                       | AI 開発者        | テスト追加時           |
| [../internal/\*/CLAUDE.md](../internal/)                                     | AI 開発者        | 実装変更時             |

## 拡張ガイド

### 新しいドキュメントを追加する場合

1. **責務を決定**

   - docs/ に属するか、internal/CLAUDE.md に属するか判定
   - docs/ = 設計・公開仕様、internal/ = 実装ガイダンス

2. **命名規則を適用**

   - docs/: `{scope}-{target}-{version}.md`
   - internal/: `CLAUDE.md` で統一

3. **体系に統合**

   - docs/ ドキュメント → root CLAUDE.md の「関連ドキュメント」に追加
   - internal/ CLAUDE.md → root CLAUDE.md の「サブディレクトリメモリ」に追加

4. **相互リンク**
   - 上位ドキュメント → 下位ドキュメントへのリンク
   - 下位ドキュメント → 上位ドキュメント（親）へのリンク

### ドキュメント削除・廃止

- **非推奨化**: `[DEPRECATED: v2で置き換え]` ヘッダー付与
- **保持期間**: 最低 1 メジャーバージョン
- **削除**: プロジェクト開始から 2 年経過後に相談

## トラブルシューティング

### Q. どこに何を書くか分からない

```
フローチャート：
  設計決定？
    → Yes: docs/architecture.md
  実装パターン？
    → Yes: internal/{package}/CLAUDE.md
  操作ガイダンス？
    → Yes: root CLAUDE.md
  テスト方法？
    → Yes: test/CLAUDE.md
  ユーザー向け？
    → Yes: README.md (変更禁止)
```

### Q. docs/ と internal/CLAUDE.md の区別

```
docs/：
  - 何を、なぜ（What・Why）設計したか
  - 複数バージョン維持
  - 人間アーキテクト向け

internal/CLAUDE.md：
  - どのように（How）実装するか
  - コード変更に合わせて即座更新
  - AI開発者・実装者向け
```

### Q. 古いドキュメントはどうする

- 仕様書（spec）：新版作成、旧版は参考資料化
- 実装ドキュメント（impl-design）：マージ・統合
- 内部メモリ（CLAUDE.md）：継続更新（バージョン分岐不可）

## 非コードファイル配置ルール

### examples/ - サンプル・実行スクリプト

**目的**：YAML タスク定義、参考スクリプト、実行例の一元管理

```
examples/
├── CLAUDE.md                    # サンプル管理ガイド ★詳細は examples/CLAUDE.md を参照
├── tasks/                       # YAML タスク定義サンプル
│   ├── sample_task_go.yaml      # Go開発用汎用サンプル
│   ├── test_codex_task.yaml     # Codex統合テスト定義
│   └── (将来) sample_python.yaml, sample_node.yaml
└── scripts/                     # テスト・デバッグスクリプト
    ├── run_codex_test.sh        # Codex統合テスト実行スクリプト
    └── (将来) debug_sandbox.sh, generate_report.sh
```

**配置判断フローチャート**：

```
ファイルの用途？

┌─ YAML タスク定義？
│  ├─ 汎用サンプル → examples/tasks/sample_*.yaml
│  └─ テスト用固定定義 → examples/tasks/test_*.yaml
│
├─ シェルスクリプト？
│  ├─ テスト/デバッグ用 → examples/scripts/{purpose}_*.sh
│  └─ 本番実行用 → cmd/ + ドキュメント
│
└─ その他（バイナリ、画像等）？
   └─ .gitignore で除外、または docs/assets/
```

### 具体例

| ファイル              | 配置先                                | 理由                     |
| --------------------- | ------------------------------------- | ------------------------ |
| Go プロジェクト実行例 | `examples/tasks/sample_task_go.yaml`  | 複数言語対応・再利用可能 |
| Codex 統合テスト定義  | `examples/tasks/test_codex_task.yaml` | テスト用・固定仕様       |
| テスト実行スクリプト  | `examples/scripts/run_codex_test.sh`  | テスト関連・非本番       |
| デバッグスクリプト    | `examples/scripts/debug_*.sh`         | 開発者用・非本番         |
| docs 画像・図表       | `docs/assets/`                        | ドキュメント資産         |

### 更新ルール

**トリガー**：

1. **YAML スキーマ変更** → サンプル (examples/tasks/\*.yaml) 同期更新
2. **テスト手法改善** → テスト定義 (examples/tasks/test\_\*.yaml) 更新
3. **新機能追加** → サンプル拡張、ガイド更新

**手順**：

```bash
# 1. サンプル更新
vi examples/tasks/sample_task_go.yaml

# 2. 実行確認
./agent-runner < examples/tasks/sample_task_go.yaml

# 3. examples/CLAUDE.md 更新（変更内容を記録）

# 4. 必要に応じて README.md や docs/ にリンク追加
```
