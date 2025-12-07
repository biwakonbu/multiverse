package agenttools

import (
	"context"
	"fmt"
)

// stubProvider is a placeholder for unsupported providers.
type stubProvider struct {
	kind string
}

func (p *stubProvider) Kind() string {
	return p.kind
}

func (p *stubProvider) Capabilities() Capability {
	return Capability{
		Kind:          p.kind,
		DefaultModel:  "",
		SupportsStdin: true,
		Notes:         "Not implemented yet; please provide a concrete implementation.",
	}
}

func (p *stubProvider) Build(_ context.Context, _ Request) (ExecPlan, error) {
	return ExecPlan{}, fmt.Errorf("agent tool provider %s is not implemented yet", p.kind)
}

func registerStub(kind string) {
	Register(kind, func(cfg ProviderConfig) (AgentToolProvider, error) {
		_ = cfg // keep signature open for future use
		return &stubProvider{kind: kind}, nil
	})
}

func init() {
	registerStub("gemini-cli")
	registerStub("claude-code")
	registerStub("cursor-cli")
}
