package e2e

import (
	"context"
	"testing"

	"github.com/biwakonbu/agent-runner/internal/agenttools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAgentToolProviders verifies that all registered providers can build valid execution plans.
// This ensures that the integration with the agenttools registry is working correctly.
func TestAgentToolProviders(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		kind       string
		req        agenttools.Request
		wantCmd    string
		wantArgs   []string
		wantStdin  bool
		wantToFail bool
	}{
		{
			kind: "codex-cli",
			req: agenttools.Request{
				Prompt: "hello",
				Mode:   "exec",
			},
			wantCmd:   "codex",
			wantArgs:  []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "-C", "/workspace/project", "--json", "-m", "gpt-5.2-codex", "-c", "reasoning_effort=medium", "hello"},
			wantStdin: false,
		},
		{
			kind: "gemini-cli",
			req: agenttools.Request{
				Prompt: "hello",
				Mode:   "exec",
			},
			wantCmd:   "gemini", // assuming default in provider config
			wantStdin: false,
		},
		{
			kind: "claude-code",
			req: agenttools.Request{
				Prompt: "hello",
				Mode:   "exec",
			},
			wantCmd:   "claude",
			wantStdin: false,
		},
		{
			kind: "claude-code-cli",
			req: agenttools.Request{
				Prompt: "hello",
				Mode:   "exec",
			},
			wantCmd:   "claude",
			wantStdin: false,
		},
		{
			kind: "cursor-cli",
			req: agenttools.Request{
				Prompt: "hello",
				Mode:   "exec",
			},
			wantCmd:   "cursor",
			wantStdin: false,
		},
		{
			kind: "claude-code",
			req: agenttools.Request{
				Prompt:   "hello piped",
				Mode:     "exec",
				UseStdin: true,
			},
			wantCmd:   "claude",
			wantStdin: true,
		},
		{
			kind: "claude-code-cli",
			req: agenttools.Request{
				Prompt:   "hello piped",
				Mode:     "exec",
				UseStdin: true,
			},
			wantCmd:   "claude",
			wantStdin: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			provider, err := agenttools.New(tt.kind, agenttools.ProviderConfig{
				CLIPath: tt.wantCmd, // Injecting expected cmd as path for verification
			})
			if tt.wantToFail {
				if err == nil {
					// It might be registered, but let's check build
				} else {
					return // Expected failure
				}
			}
			require.NoError(t, err, "Provider %s should be registered", tt.kind)

			plan, err := provider.Build(ctx, tt.req)
			require.NoError(t, err)

			assert.Equal(t, tt.wantCmd, plan.Command)
			if tt.wantStdin {
				assert.NotEmpty(t, plan.Stdin)
				assert.Equal(t, tt.req.Prompt, plan.Stdin)
			} else {
				assert.Empty(t, plan.Stdin)
				// partial check for args
				if len(tt.wantArgs) > 0 {
					assert.Subset(t, plan.Args, tt.wantArgs)
				}
			}
		})
	}
}
