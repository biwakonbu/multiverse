package ide

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLMConfigStore_LoadDefault(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	config, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, "mock", config.Kind)
	assert.Equal(t, "gpt-4o", config.Model)
}

func TestLLMConfigStore_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	// 保存
	config := &LLMConfig{
		Kind:         "openai-chat",
		Model:        "gpt-4o-mini",
		BaseURL:      "https://api.example.com/v1",
		SystemPrompt: "You are a helpful assistant.",
	}
	err := store.Save(config)
	require.NoError(t, err)

	// ファイルが作成されたことを確認
	configPath := filepath.Join(tmpDir, "config", "llm.json")
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// 読み込み
	loaded, err := store.Load()
	require.NoError(t, err)
	assert.Equal(t, config.Kind, loaded.Kind)
	assert.Equal(t, config.Model, loaded.Model)
	assert.Equal(t, config.BaseURL, loaded.BaseURL)
	assert.Equal(t, config.SystemPrompt, loaded.SystemPrompt)
}

func TestLLMConfigStore_GetAPIKey_FromEnv(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	// 環境変数を設定
	t.Setenv("OPENAI_API_KEY", "sk-test-key-12345")

	key, err := store.GetAPIKey()
	require.NoError(t, err)
	assert.Equal(t, "sk-test-key-12345", key)
}

func TestLLMConfigStore_GetAPIKey_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	// 環境変数をクリア
	t.Setenv("OPENAI_API_KEY", "")

	key, err := store.GetAPIKey()
	require.NoError(t, err)
	assert.Empty(t, key)
}

func TestLLMConfigStore_HasAPIKey(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	// 環境変数なし
	t.Setenv("OPENAI_API_KEY", "")
	assert.False(t, store.HasAPIKey())

	// 環境変数あり
	t.Setenv("OPENAI_API_KEY", "sk-test")
	assert.True(t, store.HasAPIKey())
}

func TestLLMConfigStore_GetEffectiveConfig_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	// 設定ファイルを保存
	config := &LLMConfig{
		Kind:  "mock",
		Model: "gpt-4o",
	}
	err := store.Save(config)
	require.NoError(t, err)

	// 環境変数でオーバーライド
	t.Setenv("MULTIVERSE_META_KIND", "openai-chat")
	t.Setenv("MULTIVERSE_META_MODEL", "gpt-4o-mini")

	effective, err := store.GetEffectiveConfig()
	require.NoError(t, err)
	assert.Equal(t, "openai-chat", effective.Kind)
	assert.Equal(t, "gpt-4o-mini", effective.Model)
}

func TestLLMConfigStore_SetAPIKey_NotImplemented(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewLLMConfigStore(tmpDir)

	err := store.SetAPIKey("sk-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}
