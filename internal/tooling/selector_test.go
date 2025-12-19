package tooling

import (
	"testing"

	"github.com/biwakonbu/agent-runner/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestSelector_ForceCandidateOverrides(t *testing.T) {
	cfg := &config.ToolingConfig{
		Force: config.ToolForce{
			Enabled: true,
			Tool:    "mock",
			Model:   "force-model",
		},
		Profiles: []config.ToolProfile{
			{
				ID: "p1",
				Categories: map[string]config.ToolCategoryConfig{
					CategoryMeta: {
						Strategy: "round_robin",
						Candidates: []config.ToolCandidate{
							{Tool: "mock", Model: "candidate-model"},
						},
					},
				},
			},
		},
	}

	selector := NewSelector(cfg)
	candidate, ok := selector.Select(CategoryMeta)
	require.True(t, ok)
	require.Equal(t, "mock", candidate.Tool)
	require.Equal(t, "force-model", candidate.Model)
}

func TestSelector_CategoryFallbackToMeta(t *testing.T) {
	cfg := &config.ToolingConfig{
		Profiles: []config.ToolProfile{
			{
				ID: "p1",
				Categories: map[string]config.ToolCategoryConfig{
					CategoryMeta: {
						Strategy: "round_robin",
						Candidates: []config.ToolCandidate{
							{Tool: "mock", Model: "meta"},
						},
					},
				},
			},
		},
	}

	selector := NewSelector(cfg)
	category, ok := selector.Category(CategoryTask)
	require.True(t, ok)
	require.Equal(t, "round_robin", category.Strategy)
	require.Len(t, category.Candidates, 1)
	require.Equal(t, "meta", category.Candidates[0].Model)
}

func TestSelector_RoundRobinCooldown(t *testing.T) {
	cfg := &config.ToolingConfig{
		Profiles: []config.ToolProfile{
			{
				ID: "p1",
				Categories: map[string]config.ToolCategoryConfig{
					CategoryWorker: {
						Strategy: "round_robin",
						Candidates: []config.ToolCandidate{
							{Tool: "mock", Model: "alpha"},
							{Tool: "mock", Model: "beta"},
						},
					},
				},
			},
		},
	}

	selector := NewSelector(cfg)

	first, ok := selector.Select(CategoryWorker)
	require.True(t, ok)
	second, ok := selector.Select(CategoryWorker)
	require.True(t, ok)
	require.NotEqual(t, first.Model, second.Model)

	selector.MarkRateLimited(CategoryWorker, first, 60)

	next, ok := selector.Select(CategoryWorker)
	require.True(t, ok)
	require.Equal(t, second.Model, next.Model)
}
