package meta

import (
	"context"
	"testing"

	"github.com/biwakonbu/agent-runner/internal/tooling"
	"github.com/biwakonbu/agent-runner/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestToolingClient_Decompose_UsesToolingCandidate(t *testing.T) {
	cfg := &config.ToolingConfig{
		ActiveProfile: "alpha",
		Profiles: []config.ToolProfile{
			{
				ID: "alpha",
				Categories: map[string]config.ToolCategoryConfig{
					tooling.CategoryPlan: {
						Strategy: "round_robin",
						Candidates: []config.ToolCandidate{
							{Tool: "mock"},
						},
					},
				},
			},
		},
	}

	client := NewToolingClient(cfg, "", nil, "")
	resp, err := client.Decompose(context.Background(), &DecomposeRequest{
		UserInput: "hello",
		Context:   DecomposeContext{},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Phases)
}

func TestToolingClient_FallbackWhenNoProfile(t *testing.T) {
	cfg := &config.ToolingConfig{}
	fallback := NewMockClient()

	client := NewToolingClient(cfg, "", fallback, "")
	resp, err := client.PlanTask(context.Background(), "PRD")
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "TASK-MOCK", resp.TaskID)
}
