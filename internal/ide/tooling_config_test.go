package ide

import (
	"testing"

	"github.com/biwakonbu/agent-runner/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestToolingConfigStore_LoadDefault(t *testing.T) {
	store := NewToolingConfigStore(t.TempDir())

	cfg, err := store.Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, "balanced", cfg.ActiveProfile)
	require.NotEmpty(t, cfg.Profiles)
}

func TestToolingConfigStore_SaveAndLoad(t *testing.T) {
	store := NewToolingConfigStore(t.TempDir())

	original := &config.ToolingConfig{
		ActiveProfile: "custom",
		Profiles: []config.ToolProfile{
			{
				ID:   "custom",
				Name: "Custom",
				Categories: map[string]config.ToolCategoryConfig{
					"meta": {
						Strategy: "round_robin",
						Candidates: []config.ToolCandidate{
							{Tool: "mock", Model: "alpha"},
						},
					},
				},
			},
		},
	}

	require.NoError(t, store.Save(original))

	loaded, err := store.Load()
	require.NoError(t, err)
	require.Equal(t, original, loaded)
}
