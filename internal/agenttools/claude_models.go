package agenttools

// ClaudeModelsDocURL は公式のモデル一覧。
// 参照: https://platform.claude.com/docs/en/about-claude/models/overview
const ClaudeModelsDocURL = "https://platform.claude.com/docs/en/about-claude/models/overview"

// KnownClaudeModels は、公式ドキュメント上に現れたモデル ID の一覧（抜粋）。
// 将来モデルが追加されても、この一覧に含まれないことはあり得る（= バリデーション目的では使わない）。
var KnownClaudeModels = []string{
	// Claude 4.5
	"claude-haiku-4-5",
	"claude-haiku-4-5-20251001",
	"claude-haiku-4-5-20251001-v1",
	"claude-sonnet-4-5",
	"claude-sonnet-4-5-20250929",
	"claude-sonnet-4-5-20250929-v1",
	"claude-opus-4-5",
	"claude-opus-4-5-20251101",
	"claude-opus-4-5-20251101-v1",

	// Claude 4.x
	"claude-sonnet-4",
	"claude-sonnet-4-0",
	"claude-sonnet-4-20250514",
	"claude-sonnet-4-20250514-v1",
	"claude-opus-4",
	"claude-opus-4-0",
	"claude-opus-4-1",
	"claude-opus-4-1-20250805",
	"claude-opus-4-1-20250805-v1",
	"claude-opus-4-20250514",
	"claude-opus-4-20250514-v1",

	// Claude 3.x (legacy)
	"claude-3-7-sonnet",
	"claude-3-7-sonnet-20250219",
	"claude-3-7-sonnet-20250219-v1",
	"claude-3-7-sonnet-latest",
	"claude-3-5-haiku",
	"claude-3-5-haiku-20241022",
	"claude-3-5-haiku-20241022-v1",
	"claude-3-5-haiku-latest",
	"claude-3-haiku",
	"claude-3-haiku-20240307",
	"claude-3-haiku-20240307-v1",
}
