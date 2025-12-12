package meta

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// CodexCLI のバージョンごとのレスポンス形式を検証するテスト
// Codex CLI のレスポンス形式が変更された場合、このテストで検出できる

// CodexVersionTestCase はバージョンごとのテストケースを定義する
type CodexVersionTestCase struct {
	Version      string // Codex CLI バージョン
	Description  string // テストケースの説明
	RawResponse  string // Codex CLI の生レスポンス
	ExpectedJSON string // 期待される抽出 JSON
}

// codexVersionTestCases は各バージョンのテストケースを保持する
// 新しいバージョンで形式が変わった場合、ここにテストケースを追加する
var codexVersionTestCases = []CodexVersionTestCase{
	{
		Version:     "0.65.0",
		Description: "ヘッダー + thinking セクション + JSON 形式（decompose）",
		RawResponse: `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /Users/biwakonbu/github/multiverse
model: gpt-5.2
provider: openai
approval: never
sandbox: read-only
reasoning effort: high
reasoning summaries: auto
session id: 019af8e3-f316-70d2-b602-301555ff71c8
--------
user
You are a Meta-agent that decomposes user requests into structured development tasks.
Your goal is to:
1. Understand the user's intent from their message
...
Please decompose this request into structured tasks.
mcp startup: no servers
thinking **Estimating implementation effort**
I'm considering the time estimates for various tasks...
codex
{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "ユーザーの要求を理解しました",
    "phases": [
      {
        "name": "概念設計",
        "milestone": "M1-Feature-Design",
        "tasks": [
          {
            "id": "temp-001",
            "title": "タスクタイトル",
            "description": "詳細な説明",
            "acceptance_criteria": ["達成条件1"],
            "dependencies": [],
            "wbs_level": 1,
            "estimated_effort": "small"
          }
        ]
      }
    ],
    "potential_conflicts": []
  }
}
tokens used 10,158`,
		ExpectedJSON: `{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "ユーザーの要求を理解しました",
    "phases": [
      {
        "name": "概念設計",
        "milestone": "M1-Feature-Design",
        "tasks": [
          {
            "id": "temp-001",
            "title": "タスクタイトル",
            "description": "詳細な説明",
            "acceptance_criteria": ["達成条件1"],
            "dependencies": [],
            "wbs_level": 1,
            "estimated_effort": "small"
          }
        ]
      }
    ],
    "potential_conflicts": []
  }
}`,
	},
	{
		Version:     "0.65.0",
		Description: "ヘッダー + JSON 形式（plan_task）",
		RawResponse: `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
provider: openai
--------
user
PRD content here...
codex
{
  "type": "plan_task",
  "version": 1,
  "payload": {
    "task_id": "TASK-001",
    "acceptance_criteria": [
      {
        "id": "AC-1",
        "description": "機能が実装されている",
        "type": "e2e",
        "critical": true
      }
    ]
  }
}`,
		ExpectedJSON: `{
  "type": "plan_task",
  "version": 1,
  "payload": {
    "task_id": "TASK-001",
    "acceptance_criteria": [
      {
        "id": "AC-1",
        "description": "機能が実装されている",
        "type": "e2e",
        "critical": true
      }
    ]
  }
}`,
	},
	{
		Version:     "0.65.0",
		Description: "単一行 JSON（シンプルなレスポンス）",
		RawResponse: `OpenAI Codex v0.65.0
--------
codex
{"type": "next_action", "version": 1, "payload": {"decision": {"action": "mark_complete", "reason": "完了"}}}`,
		ExpectedJSON: `{"type": "next_action", "version": 1, "payload": {"decision": {"action": "mark_complete", "reason": "完了"}}}`,
	},
	{
		Version:     "0.65.0",
		Description: "JSON が2回出力されるパターン（tokens used の後に重複）",
		RawResponse: `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
provider: openai
--------
user
Some prompt...
mcp startup: no servers
thinking **Some thinking**
codex
{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "テスト",
    "phases": [],
    "potential_conflicts": []
  }
}
tokens used 6,290
{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "テスト",
    "phases": [],
    "potential_conflicts": []
  }
}`,
		ExpectedJSON: `{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "テスト",
    "phases": [],
    "potential_conflicts": []
  }
}`,
	},
	{
		Version:     "0.65.0",
		Description: "reasoning effort と reasoning summaries 設定付きヘッダー",
		RawResponse: `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
provider: openai
approval: never
sandbox: read-only
reasoning effort: high
reasoning summaries: auto
session id: 019af8e3-f316-70d2-b602-301555ff71c8
--------
user
Prompt here
codex
{"type": "test", "version": 1, "payload": {"data": "value"}}`,
		ExpectedJSON: `{"type": "test", "version": 1, "payload": {"data": "value"}}`,
	},
}

// TestCodexResponseFormatByVersion はバージョンごとのレスポンス形式を検証する
func TestCodexResponseFormatByVersion(t *testing.T) {
	for _, tc := range codexVersionTestCases {
		t.Run(tc.Version+"_"+tc.Description, func(t *testing.T) {
			// JSON を抽出
			got := extractJSON(tc.RawResponse)

			// 抽出結果が期待値と一致するか確認
			if got != tc.ExpectedJSON {
				t.Errorf("extractJSON() mismatch for version %s\ngot:\n%s\nwant:\n%s",
					tc.Version, got, tc.ExpectedJSON)
			}

			// 抽出した JSON が有効かどうか確認
			var parsed interface{}
			if err := json.Unmarshal([]byte(got), &parsed); err != nil {
				t.Errorf("extractJSON() returned invalid JSON for version %s: %v\nJSON:\n%s",
					tc.Version, err, got)
			}
		})
	}
}

// TestCodexJSONToDecomposeResponse は抽出した JSON が DecomposeResponse に変換できることを確認
func TestCodexJSONToDecomposeResponse(t *testing.T) {
	// v0.65.0 の decompose レスポンスをテスト
	rawResponse := `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.2
--------
codex
{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "テスト用の理解内容",
    "phases": [
      {
        "name": "概念設計",
        "milestone": "M1-Test",
        "tasks": [
          {
            "id": "temp-001",
            "title": "テストタスク",
            "description": "テスト説明",
            "acceptance_criteria": ["条件1", "条件2"],
            "dependencies": [],
            "wbs_level": 1,
            "estimated_effort": "small"
          }
        ]
      },
      {
        "name": "実装",
        "milestone": "M2-Test",
        "tasks": [
          {
            "id": "temp-002",
            "title": "実装タスク",
            "description": "実装説明",
            "acceptance_criteria": ["実装条件"],
            "dependencies": ["temp-001"],
            "wbs_level": 3,
            "estimated_effort": "medium"
          }
        ]
      }
    ],
    "potential_conflicts": [
      {
        "file": "src/test.ts",
        "tasks": ["temp-002"],
        "warning": "競合の可能性"
      }
    ]
  }
}`

	// JSON 抽出
	jsonStr := extractJSON(rawResponse)

	// YAML に変換
	yamlStr, err := jsonToYAML(jsonStr)
	if err != nil {
		t.Fatalf("jsonToYAML failed: %v", err)
	}

	// MetaMessage としてパース
	var msg MetaMessage
	if err := yamlUnmarshal([]byte(yamlStr), &msg); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	// payload を DecomposeResponse に変換
	payloadBytes, err := yamlMarshal(msg.Payload)
	if err != nil {
		t.Fatalf("yaml.Marshal payload failed: %v", err)
	}

	var decompose DecomposeResponse
	if err := yamlUnmarshal(payloadBytes, &decompose); err != nil {
		t.Fatalf("yaml.Unmarshal decompose failed: %v", err)
	}

	// 検証
	if decompose.Understanding != "テスト用の理解内容" {
		t.Errorf("Understanding = %q, want %q", decompose.Understanding, "テスト用の理解内容")
	}

	if len(decompose.Phases) != 2 {
		t.Fatalf("len(Phases) = %d, want 2", len(decompose.Phases))
	}

	if decompose.Phases[0].Name != "概念設計" {
		t.Errorf("Phases[0].Name = %q, want %q", decompose.Phases[0].Name, "概念設計")
	}

	if len(decompose.Phases[0].Tasks) != 1 {
		t.Fatalf("len(Phases[0].Tasks) = %d, want 1", len(decompose.Phases[0].Tasks))
	}

	task := decompose.Phases[0].Tasks[0]
	if task.ID != "temp-001" {
		t.Errorf("task.ID = %q, want %q", task.ID, "temp-001")
	}
	if task.Title != "テストタスク" {
		t.Errorf("task.Title = %q, want %q", task.Title, "テストタスク")
	}
	if len(task.AcceptanceCriteria) != 2 {
		t.Errorf("len(task.AcceptanceCriteria) = %d, want 2", len(task.AcceptanceCriteria))
	}
	if task.WBSLevel != 1 {
		t.Errorf("task.WBSLevel = %d, want 1", task.WBSLevel)
	}
	if task.EstimatedEffort != "small" {
		t.Errorf("task.EstimatedEffort = %q, want %q", task.EstimatedEffort, "small")
	}

	// 依存関係の検証
	task2 := decompose.Phases[1].Tasks[0]
	if len(task2.Dependencies) != 1 || task2.Dependencies[0] != "temp-001" {
		t.Errorf("task2.Dependencies = %v, want [temp-001]", task2.Dependencies)
	}

	// potential_conflicts の検証
	if len(decompose.PotentialConflicts) != 1 {
		t.Fatalf("len(PotentialConflicts) = %d, want 1", len(decompose.PotentialConflicts))
	}
	conflict := decompose.PotentialConflicts[0]
	if conflict.File != "src/test.ts" {
		t.Errorf("conflict.File = %q, want %q", conflict.File, "src/test.ts")
	}
}

// TestCodexJSONToPlanTaskResponse は抽出した JSON が PlanTaskResponse に変換できることを確認
func TestCodexJSONToPlanTaskResponse(t *testing.T) {
	rawResponse := `OpenAI Codex v0.65.0
--------
codex
{
  "type": "plan_task",
  "version": 1,
  "payload": {
    "task_id": "TASK-ABC",
    "acceptance_criteria": [
      {
        "id": "AC-1",
        "description": "ユニットテストが全て通過する",
        "type": "unit",
        "critical": true
      },
      {
        "id": "AC-2",
        "description": "ドキュメントが更新されている",
        "type": "docs",
        "critical": false
      }
    ]
  }
}`

	// JSON 抽出
	jsonStr := extractJSON(rawResponse)

	// YAML に変換
	yamlStr, err := jsonToYAML(jsonStr)
	if err != nil {
		t.Fatalf("jsonToYAML failed: %v", err)
	}

	// MetaMessage としてパース
	var msg MetaMessage
	if err := yamlUnmarshal([]byte(yamlStr), &msg); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	// payload を PlanTaskResponse に変換
	payloadBytes, err := yamlMarshal(msg.Payload)
	if err != nil {
		t.Fatalf("yaml.Marshal payload failed: %v", err)
	}

	var plan PlanTaskResponse
	if err := yamlUnmarshal(payloadBytes, &plan); err != nil {
		t.Fatalf("yaml.Unmarshal plan failed: %v", err)
	}

	// 検証
	if plan.TaskID != "TASK-ABC" {
		t.Errorf("TaskID = %q, want %q", plan.TaskID, "TASK-ABC")
	}

	if len(plan.AcceptanceCriteria) != 2 {
		t.Fatalf("len(AcceptanceCriteria) = %d, want 2", len(plan.AcceptanceCriteria))
	}

	ac1 := plan.AcceptanceCriteria[0]
	if ac1.ID != "AC-1" {
		t.Errorf("AC[0].ID = %q, want %q", ac1.ID, "AC-1")
	}
	if ac1.Description != "ユニットテストが全て通過する" {
		t.Errorf("AC[0].Description = %q, want %q", ac1.Description, "ユニットテストが全て通過する")
	}
	if ac1.Type != "unit" {
		t.Errorf("AC[0].Type = %q, want %q", ac1.Type, "unit")
	}
	if !ac1.Critical {
		t.Errorf("AC[0].Critical = false, want true")
	}
}

// yamlUnmarshal と yamlMarshal は yaml パッケージのラッパー
func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func yamlMarshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}
