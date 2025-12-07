package meta

import (
	"testing"
)

func TestExtractYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain YAML",
			input:    "key: value",
			expected: "key: value",
		},
		{
			name:     "Markdown YAML block",
			input:    "Here is the yaml:\n```yaml\nkey: value\n```",
			expected: "key: value",
		},
		{
			name:     "Markdown generic block",
			input:    "```\nkey: value\n```",
			expected: "key: value",
		},
		{
			name:     "Surrounding text",
			input:    "Prefix\n```yaml\nkey: value\n```\nSuffix",
			expected: "key: value",
		},
		{
			name:     "Multiple blocks (takes first)",
			input:    "```yaml\nfirst: 1\n```\n```yaml\nsecond: 2\n```",
			expected: "first: 1",
		},
		{
			name:     "No block, just backticks",
			input:    "`key: value`",
			expected: "`key: value`", // Should not strip single backticks
		},
		{
			name:     "Leading/Trailing backticks without language",
			input:    "```\nkey: value\n```",
			expected: "key: value",
		},
		{
			name: "Codex CLI output with header",
			input: `OpenAI Codex v0.65.0 (research preview)
--------
workdir: /path/to/project
model: gpt-5.1
provider: openai
--------
user
Some prompt text here
codex
type: decompose
version: 1
payload:
  understanding: "test"`,
			expected: `type: decompose
version: 1
payload:
  understanding: "test"`,
		},
		{
			name: "Codex CLI output with thinking section",
			input: `OpenAI Codex v0.65.0
--------
thinking **Some thinking**
type: plan_task
version: 1
payload:
  task_id: "TASK-001"`,
			expected: `type: plan_task
version: 1
payload:
  task_id: "TASK-001"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractYAML(tt.input)
			if got != tt.expected {
				t.Errorf("extractYAML() = %q, want %q", got, tt.expected)
			}
		})
	}
}
