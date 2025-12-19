package meta

import (
	"context"
	"log/slog"
	"strings"

	"github.com/biwakonbu/agent-runner/internal/agenttools"
	"github.com/biwakonbu/agent-runner/internal/logging"
)

type Client struct {
	kind         string
	apiKey       string
	model        string
	systemPrompt string
	logger       *slog.Logger
	provider     Provider // Unified Provider Interface
}

// NewClient creates a new Meta client.
// It prioritizes CLI providers (codex, claude, gemini) over HTTP providers if configured or defaulted.
func NewClient(kind, apiKey, model, systemPrompt string) *Client {
	if model == "" {
		if kind == "" || kind == "openai-chat" || strings.Contains(kind, "codex") {
			model = agenttools.DefaultMetaModel // Meta-agent default model
		}
	}
	c := &Client{
		kind:         kind,
		apiKey:       apiKey,
		model:        model,
		systemPrompt: systemPrompt,
		logger:       logging.WithComponent(slog.Default(), "meta-client"),
	}

	// Initialize Provider based on kind
	switch {
	case kind == "mock":
		c.provider = NewMockClient().provider
	case strings.Contains(kind, "codex") || strings.Contains(kind, "claude") || strings.Contains(kind, "gemini"):
		// CLI based providers
		cliProvider := NewCLIProvider(kind, model, systemPrompt)
		cliProvider.SetLogger(c.logger)
		c.provider = cliProvider
	case kind == "openai-chat":
		openaiProvider := NewOpenAIProvider(apiKey, model, systemPrompt)
		openaiProvider.SetLogger(c.logger)
		c.provider = openaiProvider
	default:
		// Fallback to OpenAI if unknown, logic in app.go should prevent this but for safety
		openaiProvider := NewOpenAIProvider(apiKey, model, systemPrompt)
		openaiProvider.SetLogger(c.logger)
		c.provider = openaiProvider
	}

	return c
}

// SetLogger sets a custom logger for the client
func (c *Client) SetLogger(logger *slog.Logger) {
	c.logger = logging.WithComponent(logger, "meta-client")
	if c.provider != nil {
		if p, ok := c.provider.(interface{ SetLogger(*slog.Logger) }); ok {
			p.SetLogger(c.logger)
		}
	}
}

// TestConnection verifies the provider connection
func (c *Client) TestConnection(ctx context.Context) error {
	if c.provider == nil {
		return nil
	}
	return c.provider.TestConnection(ctx)
}

func (c *Client) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
	return c.provider.PlanTask(ctx, prdText)
}

func (c *Client) NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error) {
	return c.provider.NextAction(ctx, taskSummary)
}

func (c *Client) CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error) {
	return c.provider.CompletionAssessment(ctx, taskSummary)
}

// Decompose decomposes user input into tasks (v2.0 Chat Driven)
func (c *Client) Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error) {
	return c.provider.Decompose(ctx, req)
}

// PlanPatch generates patch operations from user input (v1.0)
func (c *Client) PlanPatch(ctx context.Context, req *PlanPatchRequest) (*PlanPatchResponse, error) {
	return c.provider.PlanPatch(ctx, req)
}

const decomposeSystemPrompt = `You are a Meta-agent that decomposes user requests into structured development tasks.

Your goal is to:
1. Understand the user's intent from their message
2. Break down the request into phases: 概念設計 (Conceptual Design) → 実装設計 (Implementation Design) → 実装 (Implementation)
3. Create detailed tasks with clear acceptance criteria
4. Identify dependencies between tasks
5. Flag potential file conflicts

IMPORTANT: Output MUST be valid JSON. Do NOT use ellipsis (...) or placeholder syntax like [...].
All arrays and values must contain actual data appropriate to the request.

Output MUST be a JSON block with the following structure:
{
  "type": "decompose",
  "version": 1,
  "payload": {
    "understanding": "ユーザーの要求を理解した内容を記述",
    "phases": [
      {
        "name": "概念設計",
        "milestone": "M1-Feature-Design",
        "tasks": [
          {
            "id": "temp-001",
            "title": "要件分析と設計ドキュメント作成",
            "description": "ユーザー要求を分析し、設計ドキュメントを作成する",
            "acceptance_criteria": [
              "設計ドキュメントが作成されている",
              "要件が明確に定義されている"
            ],
            "dependencies": [],
            "wbs_level": 1,
            "estimated_effort": "small"
          }
        ]
      },
      {
        "name": "実装設計",
        "milestone": "M1-Feature-Design",
        "tasks": [
          {
            "id": "temp-002",
            "title": "技術設計と実装計画",
            "description": "実装に必要な技術設計と計画を策定する",
            "acceptance_criteria": [
              "技術設計書が完成している",
              "実装計画が明確である"
            ],
            "dependencies": ["temp-001"],
            "wbs_level": 2,
            "estimated_effort": "medium"
          }
        ]
      },
      {
        "name": "実装",
        "milestone": "M2-Feature-Impl",
        "tasks": [
          {
            "id": "temp-003",
            "title": "機能実装",
            "description": "設計に基づいて機能を実装する",
            "acceptance_criteria": [
              "機能が実装されている",
              "テストが通過している"
            ],
            "dependencies": ["temp-002"],
            "wbs_level": 3,
            "estimated_effort": "large",
            "suggested_impl": {
              "language": "go",
              "file_paths": ["internal/feature/new.go"],
              "constraints": ["Keep backward compatibility"]
            }
          }
        ]
      }
    ],
    "potential_conflicts": [
      {
        "file": "src/example.ts",
        "tasks": ["temp-003"],
        "warning": "既存ファイルを変更する可能性があります"
      }
    ]
  }
}

Guidelines:
- WBS levels: 1=概念設計, 2=実装設計, 3=実装
- Estimated effort: small (< 1 hour), medium (1-4 hours), large (> 4 hours)
- Task IDs must start with "temp-" (they will be replaced with permanent IDs)
- Dependencies can reference other temp IDs from the same batch
- Be specific about acceptance criteria - they should be verifiable
- Consideration existing tasks to avoid duplication
- NEVER use [...] or ellipsis in your output - always provide complete, valid JSON
- For implementation tasks, ALWAYS provide 'suggested_impl' to guide the worker.
`

// planPatchSystemPrompt is for updating existing plans
const planPatchSystemPrompt = `You are a Meta-agent that maintains and edits a development plan (task list + WBS).

You are authorized to propose plan edits using the following operations:
- create: create a new task (use temp_id, Core will assign the real ID)
- update: update an existing task by task_id (partial update)
- delete: remove a task from the active plan by task_id (optionally cascade to descendants)
- move: change a task's position/parent in the WBS (and optionally update phase/milestone/wbs_level)

IMPORTANT: Output MUST be valid JSON. Do NOT use ellipsis (...) or placeholder syntax like [...].

Output MUST be a JSON block with the following structure:
{
  "type": "plan_patch",
  "version": 1,
  "payload": {
    "understanding": "ユーザーの意図を簡潔に要約",
    "operations": [
      {
        "op": "create",
        "temp_id": "temp-001",
        "title": "...",
        "description": "...",
        "acceptance_criteria": ["..."],
        "wbs_level": 1,
        "phase_name": "...",
        "parent_id": "...",
        "suggested_impl": { ... }
      },
      {
        "op": "update",
        "task_id": "TASK-123",
        "title": "New Title"
      },
      {
        "op": "delete",
        "task_id": "TASK-999",
        "cascade": true
      },
      {
        "op": "move",
        "task_id": "TASK-123",
        "parent_id": "TASK-001",
        "wbs_level": 2
      }
    ],
    "potential_conflicts": [
      {
        "file": "src/main.go",
        "tasks": ["TASK-123"],
        "warning": "..."
      }
    ]
  }
}
`
