package agenttools

// ToolKind はツール（プロバイダー）の種類を表す型
type ToolKind string

// 利用可能なツールの種類（型安全な定数）
const (
	ToolKindClaudeCode ToolKind = "claude-code"
	ToolKindGeminiCLI  ToolKind = "gemini-cli"
	ToolKindCodexCLI   ToolKind = "codex-cli"
	ToolKindOpenAIChat ToolKind = "openai-chat"
	ToolKindMock       ToolKind = "mock"
)

// ToolInfo はツールのメタデータ
type ToolInfo struct {
	Kind                 ToolKind
	Name                 string
	Description          string
	SupportedModelGroups []ModelGroup // このツールでサポートされるモデルグループ
}

// KnownTools は利用可能なツールの一覧（UI表示順）
var KnownTools = []ToolInfo{
	{
		Kind:                 ToolKindClaudeCode,
		Name:                 "Claude Code",
		Description:          "Anthropic Claude CLI",
		SupportedModelGroups: []ModelGroup{ModelGroupAuto, ModelGroupAnthropic},
	},
	{
		Kind:                 ToolKindGeminiCLI,
		Name:                 "Gemini CLI",
		Description:          "Google Gemini CLI",
		SupportedModelGroups: []ModelGroup{ModelGroupAuto, ModelGroupGoogle},
	},
	{
		Kind:                 ToolKindCodexCLI,
		Name:                 "Codex CLI",
		Description:          "OpenAI Codex CLI",
		SupportedModelGroups: []ModelGroup{ModelGroupAuto, ModelGroupOpenAI},
	},
	{
		Kind:                 ToolKindOpenAIChat,
		Name:                 "OpenAI Chat",
		Description:          "OpenAI Chat API",
		SupportedModelGroups: []ModelGroup{ModelGroupAuto, ModelGroupOpenAI},
	},
	{
		Kind:                 ToolKindMock,
		Name:                 "Mock",
		Description:          "テスト用モック",
		SupportedModelGroups: []ModelGroup{ModelGroupAuto, ModelGroupAnthropic, ModelGroupOpenAI, ModelGroupGoogle},
	},
}

// IsValidToolKind は指定された文字列が有効なツール種類かどうかを返す
func IsValidToolKind(kind string) bool {
	for _, t := range KnownTools {
		if string(t.Kind) == kind {
			return true
		}
	}
	return false
}

// GetToolInfo は指定されたツール種類の情報を返す
func GetToolInfo(kind string) *ToolInfo {
	for i := range KnownTools {
		if string(KnownTools[i].Kind) == kind {
			return &KnownTools[i]
		}
	}
	return nil
}

// GetModelsForTool は指定されたツールでサポートされるモデル一覧を返す
func GetModelsForTool(toolKind string) []ModelInfo {
	toolInfo := GetToolInfo(toolKind)
	if toolInfo == nil {
		return nil
	}

	// サポートされるモデルグループをマップに変換
	supportedGroups := make(map[ModelGroup]bool)
	for _, g := range toolInfo.SupportedModelGroups {
		supportedGroups[g] = true
	}

	// サポートされるモデルをフィルタリング
	var models []ModelInfo
	for _, m := range KnownModels {
		if supportedGroups[m.Group] {
			models = append(models, m)
		}
	}
	return models
}

// IsValidToolModelCombination はツールとモデルの組み合わせが有効かどうかを返す
func IsValidToolModelCombination(toolKind, modelID string) bool {
	// 空のモデルID（Default）は常に有効
	if modelID == "" {
		return true
	}

	models := GetModelsForTool(toolKind)
	for _, m := range models {
		if m.ID == modelID {
			return true
		}
	}
	return false
}

// ModelGroup はモデルのグループ（プロバイダー）を表す型
type ModelGroup string

// モデルグループの定数
const (
	ModelGroupAuto      ModelGroup = "Auto"
	ModelGroupAnthropic ModelGroup = "Anthropic"
	ModelGroupOpenAI    ModelGroup = "OpenAI"
	ModelGroupGoogle    ModelGroup = "Google"
)

// ModelInfo はモデルのメタデータ
type ModelInfo struct {
	ID    string
	Name  string
	Group ModelGroup
}

// KnownModels は利用可能なモデルの一覧（UI表示用）
var KnownModels = []ModelInfo{
	// Auto（デフォルト）
	{ID: "", Name: "Default", Group: ModelGroupAuto},

	// Anthropic Claude（Claude Code 用）
	{ID: "claude-opus-4-5", Name: "Claude Opus 4.5", Group: ModelGroupAnthropic},
	{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5", Group: ModelGroupAnthropic},
	{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5", Group: ModelGroupAnthropic},

	// OpenAI（Codex CLI 用）
	{ID: "gpt-5.2", Name: "GPT-5.2", Group: ModelGroupOpenAI},
	{ID: "gpt-5.2-codex", Name: "GPT-5.2 Codex", Group: ModelGroupOpenAI},
	{ID: "gpt-5.1-codex-mini", Name: "GPT-5.1 Codex Mini", Group: ModelGroupOpenAI},

	// Google Gemini（Gemini CLI 用）
	{ID: "gemini-3-pro", Name: "Gemini 3 Pro", Group: ModelGroupGoogle},
	{ID: "gemini-3-flash", Name: "Gemini 3 Flash", Group: ModelGroupGoogle},
}
