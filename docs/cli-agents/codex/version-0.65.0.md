# Codex CLI バージョン 0.65.0

確認日: 2025-12-07

## インストール

```bash
npm i -g @openai/codex
# または
brew install --cask codex
```

## サブコマンド

| コマンド | 説明 | エイリアス |
|---------|------|-----------|
| `exec` | 非対話モードで実行 | `e` |
| `review` | コードレビューを非対話モードで実行 | - |
| `login` | ログイン管理 | - |
| `logout` | 認証情報を削除 | - |
| `mcp` | MCP サーバー管理（実験的） | - |
| `mcp-server` | MCP サーバー起動（実験的） | - |
| `app-server` | アプリサーバー起動（実験的） | - |
| `completion` | シェル補完スクリプト生成 | - |
| `sandbox` | サンドボックス内でコマンド実行 | `debug` |
| `apply` | 最新の diff を適用 | `a` |
| `resume` | 前回のセッションを再開 | - |
| `cloud` | Codex Cloud からタスクを取得（実験的） | - |
| `features` | フィーチャーフラグを確認 | - |

## exec サブコマンドオプション

### 基本オプション

| フラグ | 説明 |
|--------|------|
| `-c, --config <key=value>` | 設定オーバーライド（TOML 形式） |
| `--enable <FEATURE>` | フィーチャーフラグを有効化 |
| `--disable <FEATURE>` | フィーチャーフラグを無効化 |
| `-i, --image <FILE>...` | 画像ファイルを添付 |
| `-m, --model <MODEL>` | モデルを指定 |
| `--oss` | ローカル OSS モデルを使用 |
| `--local-provider <PROVIDER>` | ローカルプロバイダ（lmstudio/ollama） |

### サンドボックス・承認オプション

| フラグ | 説明 |
|--------|------|
| `-s, --sandbox <MODE>` | `read-only` / `workspace-write` / `danger-full-access` |
| `--full-auto` | `-a on-request --sandbox workspace-write` のショートカット |
| `--dangerously-bypass-approvals-and-sandbox` | サンドボックス・承認を完全無効化（**Docker 内専用**） |

### ディレクトリ・パスオプション

| フラグ | 説明 |
|--------|------|
| `-C, --cd <DIR>` | 作業ディレクトリを指定 |
| `--add-dir <DIR>` | 追加の書き込み可能ディレクトリ |
| `--skip-git-repo-check` | Git リポジトリ外での実行を許可 |

### 出力オプション

| フラグ | 説明 |
|--------|------|
| `--json` | JSONL 形式で出力 |
| `-o, --output-last-message <FILE>` | 最後のメッセージをファイルに出力 |
| `--output-schema <FILE>` | 出力スキーマを指定 |
| `--color <COLOR>` | カラー設定（always/never/auto） |

### その他

| フラグ | 説明 |
|--------|------|
| `-p, --profile <PROFILE>` | 設定プロファイルを指定 |
| `-h, --help` | ヘルプを表示 |
| `-V, --version` | バージョンを表示 |

## 設定オーバーライド (-c)

`-c` フラグで `~/.codex/config.toml` の設定をオーバーライド可能:

```bash
# モデル指定
-c model="o3"

# サンドボックス権限
-c 'sandbox_permissions=["disk-full-read-access"]'

# 環境変数継承
-c shell_environment_policy.inherit=all

# 思考の深さ
-c reasoning_effort=medium

# サンプリング設定
-c temperature=0.5
-c max_tokens=4000
```

## stdin 入力

PROMPT 引数を省略するか `-` を指定すると stdin から読み取り:

```bash
# 省略パターン
echo "prompt" | codex exec --json

# 明示的指定
echo "prompt" | codex exec --json -
```

## 既知の制限

### プラットフォーム

- **macOS**: 完全サポート
- **Linux**: 完全サポート
- **Windows**: 実験的（WSL 推奨）

### サンドボックス

- macOS: Seatbelt 使用
- Linux: Landlock/seccomp 使用
- Docker 内: 無効化推奨（Docker が外部サンドボックスとして機能）

## 前バージョンからの変更点

### 0.58.0 → 0.65.0

- `review` サブコマンド追加
- `--local-provider` オプション追加（lmstudio 対応）
- 各種バグ修正・安定性向上

## 参考リンク

- [Codex CLI 公式ドキュメント](https://developers.openai.com/codex/cli/)
- [Codex CLI リファレンス](https://developers.openai.com/codex/cli/reference)
- [Codex セキュリティ](https://developers.openai.com/codex/security/)
- [GitHub リポジトリ](https://github.com/openai/codex)
