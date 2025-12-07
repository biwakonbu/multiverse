package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LLMConfig は LLM プロバイダの設定
type LLMConfig struct {
	Kind         string `json:"kind"`                   // mock, openai-chat
	Model        string `json:"model"`                  // gpt-4o, gpt-4o-mini
	BaseURL      string `json:"baseUrl,omitempty"`      // カスタムエンドポイント
	SystemPrompt string `json:"systemPrompt,omitempty"` // カスタムシステムプロンプト
}

// DefaultLLMConfig はデフォルトの LLM 設定を返す
func DefaultLLMConfig() *LLMConfig {
	return &LLMConfig{
		Kind:  "mock",
		Model: "gpt-4o",
	}
}

// LLMConfigStore は LLM 設定の永続化を担当する
type LLMConfigStore struct {
	configPath string
}

// NewLLMConfigStore は新しい LLMConfigStore を作成する
// baseDir は通常 $HOME/.multiverse
func NewLLMConfigStore(baseDir string) *LLMConfigStore {
	configDir := filepath.Join(baseDir, "config")
	return &LLMConfigStore{
		configPath: filepath.Join(configDir, "llm.json"),
	}
}

// Load は LLM 設定を読み込む
func (s *LLMConfigStore) Load() (*LLMConfig, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 設定ファイルがない場合はデフォルトを返す
			return DefaultLLMConfig(), nil
		}
		return nil, fmt.Errorf("failed to read llm config: %w", err)
	}

	var config LLMConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal llm config: %w", err)
	}
	return &config, nil
}

// Save は LLM 設定を保存する
func (s *LLMConfigStore) Save(config *LLMConfig) error {
	// ディレクトリを作成
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal llm config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write llm config: %w", err)
	}

	return nil
}

// GetAPIKey は API キーを取得する
// 優先順位: 環境変数 OPENAI_API_KEY > 設定ファイル（将来: OS keychain）
func (s *LLMConfigStore) GetAPIKey() (string, error) {
	// まず環境変数を確認
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key, nil
	}

	// 将来的には OS keychain からの読み込みを実装
	// 現時点では環境変数のみサポート
	return "", nil
}

// SetAPIKey は API キーを保存する
// 注意: 現時点では環境変数を推奨。設定ファイルへの保存はセキュリティリスクがあるため未実装
func (s *LLMConfigStore) SetAPIKey(_ string) error {
	// 将来的には OS keychain への保存を実装
	// 現時点では環境変数 OPENAI_API_KEY の設定を推奨
	return fmt.Errorf("API key storage not implemented: please set OPENAI_API_KEY environment variable")
}

// HasAPIKey は API キーが設定されているかを確認する
func (s *LLMConfigStore) HasAPIKey() bool {
	key, _ := s.GetAPIKey()
	return key != ""
}

// GetEffectiveConfig は環境変数を考慮した実効設定を返す
func (s *LLMConfigStore) GetEffectiveConfig() (*LLMConfig, error) {
	config, err := s.Load()
	if err != nil {
		return nil, err
	}

	// 環境変数でのオーバーライドを適用
	if kind := os.Getenv("MULTIVERSE_META_KIND"); kind != "" {
		config.Kind = kind
	}
	if model := os.Getenv("MULTIVERSE_META_MODEL"); model != "" {
		config.Model = model
	}
	if prompt := os.Getenv("MULTIVERSE_META_SYSTEM_PROMPT"); prompt != "" {
		config.SystemPrompt = prompt
	}

	return config, nil
}
