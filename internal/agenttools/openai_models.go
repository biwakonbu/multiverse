package agenttools

// OpenAIPricingDocURL は OpenAI 公式の Pricing ドキュメント。
// 参照: https://platform.openai.com/docs/pricing
const OpenAIPricingDocURL = "https://platform.openai.com/docs/pricing"

// OpenAIModelInfo は OpenAI/Codex 系モデルのメタデータ（抜粋）。
//
// NOTE:
// - この一覧は「バリデーション目的」では使わない（将来モデルが追加されても、この一覧に含まれないことはあり得る）。
// - 価格やコンテキスト長などは docs/pricing を根拠に更新される想定だが、取得が難しい実行環境もあるため、未設定(nil)を許容する。
type OpenAIModelInfo struct {
	// ID は OpenAI API / Codex CLI が受け付けるモデル ID（推奨: 公式表記）。
	ID string

	// Aliases は、プロダクト側で便宜的に使われる別名（短縮名など）。
	// 例: "5.1-codex-mini" → "gpt-5.1-codex-mini"
	Aliases []string

	// Use は想定用途（Meta/Worker など）を簡易に分類するためのラベル。
	Use string

	// Notes は UI/ドキュメント表示用の補足。
	Notes string

	// PricingSourceURL は価格情報の一次ソース（公式ドキュメント URL）。
	PricingSourceURL string

	// PricingUSDPerMTok は $/1M tokens を格納する（未確認の場合は nil）。
	PricingUSDPerMTok *ModelPricingUSDPerMTok
}

// ModelPricingUSDPerMTok は入出力トークンの価格（USD / 1M tokens）。
// docs/pricing の記載を人手で反映する前提。
type ModelPricingUSDPerMTok struct {
	Input  float64
	Output float64
}

// KnownOpenAIModels は、利用頻度が高いモデル ID の一覧（抜粋）。
// 公式Pricing/Docsに掲載されるモデルのうち、multiverse が「利用可能モデル」として提示したいものを登録する。
var KnownOpenAIModels = []OpenAIModelInfo{
	{
		ID:               "gpt-5.2",
		Use:              "meta",
		Notes:            "Meta-agent（計画・分解）向けのデフォルト候補。",
		PricingSourceURL: OpenAIPricingDocURL,
	},
	{
		ID:               "gpt-5.2-codex",
		Use:              "worker",
		Notes:            "Codex CLI（Worker 実行）向けのデフォルト。",
		PricingSourceURL: OpenAIPricingDocURL,
	},
	{
		ID:               "gpt-5.1-codex-mini",
		Aliases:          []string{"5.1-codex-mini"},
		Use:              "worker-fast",
		Notes:            "Codex CLI の高速/低コスト系（短縮名: 5.1-codex-mini）。",
		PricingSourceURL: OpenAIPricingDocURL,
	},
}

// ResolveOpenAIModelID は alias を正規のモデル ID に正規化する。
// 未知の値はそのまま返す（バリデーション用途では使わない）。
func ResolveOpenAIModelID(model string) string {
	if model == "" {
		return model
	}
	for _, m := range KnownOpenAIModels {
		if model == m.ID {
			return model
		}
		for _, a := range m.Aliases {
			if model == a {
				return m.ID
			}
		}
	}
	return model
}
