# CLI Package - コマンドラインフラグ処理

このパッケージは AgentRunner Core CLI のコマンドラインフラグ処理を提供します。

## 責務

- **フラグパース**: コマンドラインフラグの解析
- **設定解決**: フラグ、YAML 設定、デフォルト値の優先順位に基づく設定解決

## ファイル構成

| ファイル | 役割 |
|---------|------|
| flags.go | フラグ定義とパース処理 |
| flags_test.go | フラグ処理のユニットテスト |

## 主要データモデル

### Flags 構造体

```go
type Flags struct {
    MetaModel string  // Meta-agent の LLM モデル ID
}
```

## API

### ParseFlags

```go
func ParseFlags(args []string, output io.Writer) (*Flags, error)
```

コマンドライン引数をパースし、Flags 構造体を返します。

**パラメータ**:
- `args`: `os.Args[1:]` 相当のコマンドライン引数
- `output`: エラーメッセージの出力先（通常 `os.Stderr`）

**サポートされるフラグ**:
- `-meta-model`: Meta-agent の LLM モデル ID

### ResolveMetaModel

```go
func ResolveMetaModel(cliModel, yamlModel string) string
```

Meta-agent のモデル ID を優先順位に基づいて解決します。

**優先順位**:
1. CLI フラグ（`-meta-model`）
2. タスク YAML 設定（`runner.meta.model`）
3. デフォルト値（`gpt-5.2`）

## 使用例

```go
// main.go での使用
flags, err := cli.ParseFlags(os.Args[1:], os.Stderr)
if err != nil {
    return err
}

// YAML 設定と組み合わせてモデル解決
metaModel := cli.ResolveMetaModel(flags.MetaModel, cfg.Runner.Meta.Model)
```

## 設計原則

### flag パッケージの使用

- 標準ライブラリの `flag` パッケージを使用
- `flag.ContinueOnError` でエラー時もプログラムを継続可能
- 出力先を注入可能にしてテスタビリティを確保

### 設定の階層化

```
CLI フラグ > YAML 設定 > デフォルト値
```

- 明示的な指定が優先される
- デフォルト値はコード内にハードコード
- 将来的に環境変数サポートも検討可能

## テスト戦略

### ユニットテスト

`flags_test.go` でカバー:

- フラグなしの場合のデフォルト動作
- 各フラグの個別パース
- 不正なフラグのエラーハンドリング
- ResolveMetaModel の優先順位テスト

### テストパターン

```go
func TestParseFlags(t *testing.T) {
    // フラグなし
    flags, err := ParseFlags([]string{}, io.Discard)
    assert.Equal(t, "", flags.MetaModel)

    // フラグあり
    flags, err := ParseFlags([]string{"-meta-model", "gpt-4o"}, io.Discard)
    assert.Equal(t, "gpt-4o", flags.MetaModel)
}

func TestResolveMetaModel(t *testing.T) {
    // CLI フラグ優先
    assert.Equal(t, "cli-model", ResolveMetaModel("cli-model", "yaml-model"))

    // YAML 設定優先
    assert.Equal(t, "yaml-model", ResolveMetaModel("", "yaml-model"))

    // デフォルト値
    assert.Equal(t, "gpt-5.2", ResolveMetaModel("", ""))
}
```

## 拡張予定

### 短期

- [ ] `-verbose` フラグ（ログレベル制御）
- [ ] `-dry-run` フラグ（実行シミュレーション）

### 長期

- [ ] 環境変数からの設定読み込み
- [ ] 設定ファイル（`~/.agent-runner.yaml`）サポート

## 関連ドキュメント

- [../../cmd/agent-runner/CLAUDE.md](../../cmd/agent-runner/CLAUDE.md): CLI エントリポイント
- [../../pkg/config/CLAUDE.md](../../pkg/config/CLAUDE.md): YAML 設定スキーマ
