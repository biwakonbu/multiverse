# Note Package - Task Note生成・保存

このパッケージはタスク実行の完全な履歴をMarkdown形式のTask Noteとして記録し、監査証跡とデバッグの重要なアーティファクトを生成します。

## 概要

- **writer.go**: TaskContext → Markdown Task Note変換、ファイルシステム操作
- **writer_test.go**: テンプレート出力検証

## Task Note の役割

### 監査証跡（Audit Trail）
- タスク開始時刻・終了時刻を記録
- PRD全文を保存（要件の確定）
- Acceptance Criteria（AC）の合格状況
- Meta-agentの呼び出し履歴（リクエスト/レスポンス全文）
- Worker実行ログ（コマンド出力、終了コード）

### デバッグ・再現性
- タスク失敗時に何が起こったかを完全に追跡可能
- Meta-agentの判定理由をYAMLで確認
- Worker実行結果の詳細出力
- タスク再実行時の参考資料

### ドキュメント生成
- 開発変更履歴を Markdown で記録
- GitHub リポジトリに コミット可能
- Team 内での共有・レビュー用

## ファイルシステム構造

### ディレクトリ作成

```go
dir := filepath.Join(taskCtx.RepoPath, ".agent-runner")
os.MkdirAll(dir, 0755)
```

**ディレクトリ構成**:
```
<repository-root>/
└── .agent-runner/
    ├── task-TASK-001.md
    ├── task-TASK-002.md
    ├── task-TASK-003.md
    └── ...
```

**特徴**:
- `.agent-runner/` はリポジトリルート直下に作成
- ファイル名: `task-{TaskContext.ID}.md`
- ID が同じタスク = ファイル上書き（再実行で履歴更新）
- `.gitignore` に追加推奨：`*.agent-runner/` または 個別に管理

### ファイル権限

```go
os.MkdirAll(dir, 0755)  // Directory: rwxr-xr-x (755)
os.Create(path)         // File: -rw-r--r-- (644, os default)
```

- **クロスプラットフォーム対応**: Unix権限は Windows では無視される
- **パーミッション**: デフォルトでリードフレンドリー

## Task Note テンプレート構造

### テンプレート全体像（lines 34-97）

```markdown
# Task Note - <ID> - <Title>

- Task ID: <ID>
- Title: <Title>
- Started At: <StartedAt>
- Finished At: <FinishedAt>
- State: <State>

---

## 1. PRD Summary
<details>
<summary>PRD Text</summary>
```text
<PRD全文>
```
</details>

---

## 2. Acceptance Criteria
- [x] AC-1: <説明>
- [ ] AC-2: <説明>
- ...

---

## 3. Execution Log

### 3.1 Meta Calls
#### plan_task at <timestamp>
```yaml
<Request YAML>
```
```yaml
<Response YAML>
```

#### next_action at <timestamp>
...

### 3.2 Worker Runs
#### Run <ID> (ExitCode=0) at <timestamp>
Summary: <Summary>
```text
<RawOutput>
```
...
```

### セクション別の詳細

#### 1. ヘッダー（Lines 35-41）
- Task ID, Title, 開始・終了時刻、最終状態
- 一目で「何が実行されたか」「いつ」「どの状態で終わったか」を確認可能

#### 2. PRD Summary（Lines 45-54）
```html
<details>
<summary>PRD Text</summary>
...
</details>
```

**特徴**:
- `<details>` タグで折りたたみ可能（GitHub Markdown対応）
- 大規模 PRD でも Task Note が見やすく保つ
- タイトル「PRD Text」クリックで展開

**バッククォート処理** (Lines 50, 52):
```go
` + "```" + `text
{{ .PRDText }}
` + "```" + `
```

**重要**: Goテンプレートがバッククォート `\`\`\`` を誤処理しないよう、文字列連結で対応

#### 3. Acceptance Criteria（Lines 60-62）
```markdown
{{ range .AcceptanceCriteria }}
- [{{ if .Passed }}x{{ else }} {{ end }}] {{ .ID }}: {{ .Description }}
{{ end }}
```

**チェックボックス表示**:
- `[x]` : Passed=true
- `[ ]` : Passed=false

**用途**: Meta-agentの完了判定を可視化

#### 4. Meta Calls（Lines 70-81）
```markdown
#### {{ .Type }} at {{ .Timestamp }}
```yaml
{{ .RequestYAML }}
```
```yaml
{{ .ResponseYAML }}
```
```

**記録内容**:
- Type: "plan_task" or "next_action"
- Timestamp: 呼び出し時刻
- RequestYAML: Meta への要求（system prompt + user prompt）
- ResponseYAML: Meta からの応答（decision, worker_call等）

**利用シーン**:
- Meta-agentの判定根拠を確認
- YAML パースエラーが発生した場合、原因調査
- LLM レスポンスの品質評価

#### 5. Worker Runs（Lines 85-94）
```markdown
#### Run {{ .ID }} (ExitCode={{ .ExitCode }}) at {{ .StartedAt }}
Summary: {{ .Summary }}
```text
{{ .RawOutput }}
```
```

**記録内容**:
- Run ID: 実行一意識別子（unix timestamp基の ID）
- ExitCode: プロセス終了コード（0=成功、非0=失敗）
- Summary: 簡潔な説明（"Worker executed" など）
- RawOutput: 標準出力・標準エラー全体

**デバッグ用途**:
- コマンド実行失敗時の error message確認
- 出力ログから実装の進捗確認

## テンプレート処理の実装詳細

### Parse と Execute（Lines 99-104）

```go
tmpl, err := template.New("task-note").Parse(tmplStr)
if err != nil {
    return err
}
return tmpl.Execute(f, taskCtx)
```

**プロセス**:
1. テンプレート文字列を `template.New()` で 初期化
2. `Parse(tmplStr)` で コンパイル
3. `Execute(file, TaskContext)` で レンダリング・ファイル出力

**エラーハンドリング**:
- Parse エラー → 返り値のエラーハンドラで catch
- Execute エラー → ファイルI/O失敗 or メモリ不足の可能性

## テンプレート拡張ガイド

### 新フィールドの追加

**例: TestResult を追加する場合**

1. TaskContext にフィールド追加（internal/core/context.go）
   ```go
   type TaskContext struct {
       ...
       TestResult *TestResult
   }
   ```

2. テンプレート に セクション追加
   ```markdown
   ## 4. Test Results
   {{ if .TestResult }}
   - Command: {{ .TestResult.Command }}
   - Exit Code: {{ .TestResult.ExitCode }}
   - Summary: {{ .TestResult.Summary }}
   ```text
   {{ .TestResult.RawOutput }}
   ```
   {{ end }}
   ```

3. テンプレート内での注意点
   - `{{ if .Field }}` で nil チェック（フィールド存在確認）
   - バッククォート使用時は `"` + "```" + "` で文字列連結
   - `{{ range }}` ブロック でスライス反復処理

### マルチフォーマット対応への設計パターン

**現在**: Markdown のみ

**将来対応案** (JSON, HTML等):
```go
type Writer interface {
    Write(taskCtx *core.TaskContext) error
}

type MarkdownWriter struct {}
func (w *MarkdownWriter) Write(taskCtx *core.TaskContext) error { ... }

type JSONWriter struct {}
func (w *JSONWriter) Write(taskCtx *core.TaskContext) error { ... }
```

**利点**:
- Formatter を差し替え可能
- Runner から format パラメータで選択

## テストベストプラクティス

### テンプレート出力の検証（writer_test.go）

**テスト項目**:
1. **テンプレート パース成功**: `Parse()` がエラーを返さないか
2. **HTML エスケープ**: Markdown 中の `<script>` タグが誤実行されないか
3. **フィールド出力**: 全フィールドが正確に出力されるか
4. **ファイル作成**: `.agent-runner/` ディレクトリが生成されるか
5. **権限設定**: ファイル権限が 0644 で作成されるか

### テストケース例

```go
func TestWriter_Write_Success(t *testing.T) {
    ctx := &core.TaskContext{
        ID: "TASK-001",
        Title: "テストタスク",
        State: core.StateComplete,
        ...
    }

    w := NewWriter()
    err := w.Write(ctx)

    // 検証1: エラーなし
    if err != nil { t.Fatalf("expected nil, got %v", err) }

    // 検証2: ファイル存在確認
    path := filepath.Join(ctx.RepoPath, ".agent-runner", "task-TASK-001.md")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        t.Fatalf("file not created")
    }

    // 検証3: ファイル内容確認
    content, _ := ioutil.ReadFile(path)
    if !strings.Contains(string(content), "TASK-001") {
        t.Fatalf("Task ID not found in output")
    }
}
```

## エラーハンドリング

### ファイルシステムエラー

**MkdirAll エラー**:
```go
if err := os.MkdirAll(dir, 0755); err != nil {
    return err  // 権限不足、ディスク容量不足等
}
```

**ファイル作成エラー**:
```go
f, err := os.Create(path)
if err != nil {
    return err  // 既存ファイルが読み取り専用など
}
```

### テンプレートエラー

**Parse エラー**:
```go
tmpl, err := template.New("task-note").Parse(tmplStr)
// 原因: テンプレート構文エラー（{{ range ]] など）
```

**Execute エラー**:
```go
err := tmpl.Execute(f, taskCtx)
// 原因: field access 失敗（nil pointer）など
```

**対応** (runner.go:147-150):
- Note 出力エラー → **警告のみ**、タスク失敗にしない
- 理由: Note は監査用、コア機能の失敗ではない

## ベストプラクティス

### 1. Task ID の一意性
- 各タスクに unique ID を付与
- 同一 ID = ファイル上書き（意図的な再実行）

### 2. PRD 保存の重要性
- PRD を Task Note に 埋め込む
- 後に PRD ファイルを修正しても、Task Note に原本保存
- 要件変更のトレーサビリティ確保

### 3. ディレクトリ無視設定
```gitignore
# Option 1: .agent-runner 全体を無視
.agent-runner/

# Option 2: Task Note のみ保存（任意）
!.agent-runner/.gitkeep
```

### 4. 定期的な Task Note レビュー
- CI/CD パイプラインで Task Note を成果物として保存
- 失敗タスクの原因分析に使用
- チーム内での知見共有

## 既知の課題と改善案

### 1. バッククォートのハードコーディング

**現在**:
```go
` + "```" + `text
{{ .PRDText }}
` + "```" + `
```

**課題**: テンプレート文字列が読みにくい

**改善案**:
```go
// テンプレートを外部ファイルから読み込み
data, _ := ioutil.ReadFile("templates/task_note.tmpl")
tmpl, _ := template.New("task-note").Parse(string(data))
```

### 2. タイムスタンプのフォーマット

**現在**: Go の デフォルトフォーマット（`2006-01-02 15:04:05.000000000 -0700 MST`）

**改善案**:
```go
// カスタムフォーマット
{{ .StartedAt.Format "2006-01-02 15:04:05" }}
```

### 3. 大規模出力の容量管理

**課題**: Worker出力が数MB の場合、Task Note サイズが大きくなる

**改善案**:
- 出力ログをサマリー + リンク形式に変更
- Worker実行ログは別ファイル（logs/ ディレクトリ）に分離

## 関連ドキュメント

- [core/CLAUDE.md](../core/CLAUDE.md) - TaskContext の詳細
- [meta/CLAUDE.md](../meta/CLAUDE.md) - Meta Call のYAMLフォーマット
- [worker/CLAUDE.md](../worker/CLAUDE.md) - Worker実行結果の詳細
- `/CLAUDE.md` - ファイルシステムとTask Note の位置付け
