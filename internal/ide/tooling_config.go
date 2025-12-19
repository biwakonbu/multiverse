package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/biwakonbu/agent-runner/pkg/config"
)

// ToolingConfigStore は tooling 設定の永続化を担当する。
type ToolingConfigStore struct {
	configPath string
}

// NewToolingConfigStore は新しい ToolingConfigStore を作成する。
// baseDir は通常 $HOME/.multiverse。
func NewToolingConfigStore(baseDir string) *ToolingConfigStore {
	configDir := filepath.Join(baseDir, "config")
	return &ToolingConfigStore{
		configPath: filepath.Join(configDir, "tooling.json"),
	}
}

// DefaultToolingConfig はおすすめの tooling デフォルト設定を返す。
func DefaultToolingConfig() *config.ToolingConfig {
	return &config.ToolingConfig{
		ActiveProfile: "balanced",
		Profiles: []config.ToolProfile{
			{
				ID:   "balanced",
				Name: "Balanced",
				Categories: map[string]config.ToolCategoryConfig{
					"meta": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 40},
							{Tool: "claude-code", Model: "claude-sonnet-4-5-20250929", Weight: 30},
							{Tool: "gemini-cli", Model: "gemini-3-pro-preview", Weight: 20},
							{Tool: "openai-chat", Model: "gpt-5.2", Weight: 10},
						},
					},
					"plan": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "claude-code", Model: "claude-sonnet-4-5-20250929", Weight: 45},
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 30},
							{Tool: "gemini-cli", Model: "gemini-3-pro-preview", Weight: 25},
						},
					},
					"task": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "claude-code", Model: "claude-sonnet-4-5-20250929", Weight: 40},
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 40},
							{Tool: "gemini-cli", Model: "gemini-3-pro-preview", Weight: 20},
						},
					},
					"execution": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 40},
							{Tool: "claude-code", Model: "claude-sonnet-4-5-20250929", Weight: 30},
							{Tool: "gemini-cli", Model: "gemini-3-pro-preview", Weight: 20},
							{Tool: "openai-chat", Model: "gpt-5.2", Weight: 10},
						},
					},
					"worker": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.1-codex", Weight: 60},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 25},
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 15},
						},
					},
				},
			},
			{
				ID:   "fast",
				Name: "Fast",
				Categories: map[string]config.ToolCategoryConfig{
					"meta": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 45},
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 35},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 20},
						},
					},
					"plan": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 40},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 30},
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 30},
						},
					},
					"task": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 40},
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 30},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 30},
						},
					},
					"execution": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.2", Weight: 45},
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 35},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 20},
						},
					},
					"worker": {
						Strategy:            "weighted",
						FallbackOnRateLimit: true,
						CooldownSec:         120,
						Candidates: []config.ToolCandidate{
							{Tool: "codex-cli", Model: "gpt-5.1-codex", Weight: 70},
							{Tool: "gemini-cli", Model: "gemini-3-flash-preview", Weight: 20},
							{Tool: "claude-code", Model: "claude-haiku-4-5-20251001", Weight: 10},
						},
					},
				},
			},
		},
		Force: config.ToolForce{
			Enabled: false,
			Tool:    "",
			Model:   "",
		},
	}
}

// Load は tooling 設定を読み込み、無ければデフォルトを返す。
func (s *ToolingConfigStore) Load() (*config.ToolingConfig, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultToolingConfig(), nil
		}
		return nil, fmt.Errorf("failed to read tooling config: %w", err)
	}

	var cfg config.ToolingConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tooling config: %w", err)
	}
	return &cfg, nil
}

// Save は tooling 設定を保存する。
func (s *ToolingConfigStore) Save(cfg *config.ToolingConfig) error {
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tooling config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tooling config: %w", err)
	}
	return nil
}
