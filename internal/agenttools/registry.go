package agenttools

import (
	"context"
	"fmt"
	"sync"
)

// ProviderFactory constructs a provider for a given configuration.
type ProviderFactory func(cfg ProviderConfig) (AgentToolProvider, error)

var (
	registryMu sync.RWMutex
	registry   = map[string]ProviderFactory{}
)

// Register attaches a provider factory by kind.
// It panics if the same kind is registered twice to avoid silent overrides.
func Register(kind string, factory ProviderFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registry[kind]; exists {
		panic(fmt.Sprintf("agent tool provider already registered: %s", kind))
	}
	registry[kind] = factory
}

// New creates a provider for the given kind.
func New(kind string, cfg ProviderConfig) (AgentToolProvider, error) {
	registryMu.RLock()
	factory, ok := registry[kind]
	registryMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedKind, kind)
	}
	return factory(cfg)
}

// MustNew is a helper that panics on failure (useful for tests).
func MustNew(kind string, cfg ProviderConfig) AgentToolProvider {
	p, err := New(kind, cfg)
	if err != nil {
		panic(err)
	}
	return p
}

// Build is a convenience helper: resolve a provider and build an ExecPlan.
func Build(ctx context.Context, kind string, cfg ProviderConfig, req Request) (ExecPlan, error) {
	p, err := New(kind, cfg)
	if err != nil {
		return ExecPlan{}, err
	}
	return p.Build(ctx, req)
}
