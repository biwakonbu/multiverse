package agenttools

import (
	"testing"
)

func TestIsValidToolKind(t *testing.T) {
	tests := []struct {
		kind string
		want bool
	}{
		{"claude-code", true},
		{"codex-cli", true},
		{"gemini-cli", true},
		{"openai-chat", true},
		{"mock", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := IsValidToolKind(tt.kind)
			if got != tt.want {
				t.Errorf("IsValidToolKind(%q) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}

func TestGetToolInfo(t *testing.T) {
	t.Run("existing tool", func(t *testing.T) {
		info := GetToolInfo("claude-code")
		if info == nil {
			t.Fatal("GetToolInfo(\"claude-code\") returned nil")
		}
		if info.Kind != ToolKindClaudeCode {
			t.Errorf("Kind = %q, want %q", info.Kind, ToolKindClaudeCode)
		}
		if len(info.SupportedModelGroups) == 0 {
			t.Error("SupportedModelGroups should not be empty")
		}
	})

	t.Run("non-existing tool", func(t *testing.T) {
		info := GetToolInfo("invalid")
		if info != nil {
			t.Error("GetToolInfo(\"invalid\") should return nil")
		}
	})
}

func TestGetModelsForTool(t *testing.T) {
	tests := []struct {
		toolKind       string
		expectedGroups []ModelGroup
	}{
		{
			toolKind:       "claude-code",
			expectedGroups: []ModelGroup{ModelGroupAuto, ModelGroupAnthropic},
		},
		{
			toolKind:       "codex-cli",
			expectedGroups: []ModelGroup{ModelGroupAuto, ModelGroupOpenAI},
		},
		{
			toolKind:       "gemini-cli",
			expectedGroups: []ModelGroup{ModelGroupAuto, ModelGroupGoogle},
		},
		{
			toolKind:       "openai-chat",
			expectedGroups: []ModelGroup{ModelGroupAuto, ModelGroupOpenAI},
		},
	}

	for _, tt := range tests {
		t.Run(tt.toolKind, func(t *testing.T) {
			models := GetModelsForTool(tt.toolKind)
			if len(models) == 0 {
				t.Fatal("GetModelsForTool returned empty slice")
			}

			// 返されたモデルのグループが期待通りかチェック
			for _, m := range models {
				found := false
				for _, expectedGroup := range tt.expectedGroups {
					if m.Group == expectedGroup {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Model %q has unexpected group %q", m.ID, m.Group)
				}
			}
		})
	}

	t.Run("invalid tool returns nil", func(t *testing.T) {
		models := GetModelsForTool("invalid")
		if models != nil {
			t.Error("GetModelsForTool(\"invalid\") should return nil")
		}
	})
}

func TestIsValidToolModelCombination(t *testing.T) {
	tests := []struct {
		tool    string
		model   string
		want    bool
		comment string
	}{
		// 空モデル（Default）は常に有効
		{"claude-code", "", true, "empty model is always valid"},
		{"codex-cli", "", true, "empty model is always valid"},

		// 正しい組み合わせ
		{"claude-code", "claude-opus-4-5", true, "Claude Code + Anthropic model"},
		{"codex-cli", "gpt-5.2", true, "Codex CLI + OpenAI model"},
		{"gemini-cli", "gemini-3-pro", true, "Gemini CLI + Google model"},
		{"openai-chat", "gpt-5.2", true, "OpenAI Chat + OpenAI model"},

		// 不正な組み合わせ
		{"claude-code", "gpt-5.2", false, "Claude Code cannot use OpenAI model"},
		{"codex-cli", "claude-opus-4-5", false, "Codex CLI cannot use Anthropic model"},
		{"gemini-cli", "gpt-5.2", false, "Gemini CLI cannot use OpenAI model"},

		// Mock は全モデル対応
		{"mock", "claude-opus-4-5", true, "Mock can use any model"},
		{"mock", "gpt-5.2", true, "Mock can use any model"},
		{"mock", "gemini-3-pro", true, "Mock can use any model"},
	}

	for _, tt := range tests {
		t.Run(tt.comment, func(t *testing.T) {
			got := IsValidToolModelCombination(tt.tool, tt.model)
			if got != tt.want {
				t.Errorf("IsValidToolModelCombination(%q, %q) = %v, want %v",
					tt.tool, tt.model, got, tt.want)
			}
		})
	}
}

func TestKnownToolsHaveSupportedModelGroups(t *testing.T) {
	for _, tool := range KnownTools {
		t.Run(string(tool.Kind), func(t *testing.T) {
			if len(tool.SupportedModelGroups) == 0 {
				t.Errorf("Tool %q has no SupportedModelGroups", tool.Kind)
			}
			// 全ツールは Auto グループをサポートすべき
			hasAuto := false
			for _, g := range tool.SupportedModelGroups {
				if g == ModelGroupAuto {
					hasAuto = true
					break
				}
			}
			if !hasAuto {
				t.Errorf("Tool %q should support ModelGroupAuto", tool.Kind)
			}
		})
	}
}
