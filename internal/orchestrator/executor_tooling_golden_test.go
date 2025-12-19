package orchestrator

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/agent-runner/pkg/config"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update", false, "update golden files")

func TestGenerateTaskYAML_WithTooling_Golden(t *testing.T) {
	executor := &Executor{
		ToolingConfig: &config.ToolingConfig{
			ActiveProfile: "golden",
			Profiles: []config.ToolProfile{
				{
					ID:   "golden",
					Name: "Golden",
					Categories: map[string]config.ToolCategoryConfig{
						"worker": {
							Strategy:            "round_robin",
							FallbackOnRateLimit: true,
							CooldownSec:         90,
							Candidates: []config.ToolCandidate{
								{Tool: "mock", Model: "alpha", Weight: 1},
							},
						},
					},
				},
			},
			Force: config.ToolForce{
				Enabled: false,
			},
		},
	}

	task := &Task{
		ID:                 "task-tooling-golden",
		Title:              "Tooling YAML Golden",
		Description:        "Verify tooling YAML block",
		WBSLevel:           1,
		PhaseName:          "Planning",
		Dependencies:       []string{"task-dep-1"},
		AcceptanceCriteria: []string{"AC-1", "AC-2"},
	}

	got := executor.generateTaskYAML(task)

	goldenPath := filepath.Join("testdata", "task_yaml_with_tooling.golden")
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0755))
		require.NoError(t, os.WriteFile(goldenPath, []byte(got), 0644))
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)
	require.Equal(t, string(expected), got)
}
