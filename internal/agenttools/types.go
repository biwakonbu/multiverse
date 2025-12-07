package agenttools

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrUnsupportedKind is returned when a provider kind is not registered.
var ErrUnsupportedKind = errors.New("agent tool provider not registered")

// ErrUnsupportedMode is returned when a provider does not support the requested mode.
var ErrUnsupportedMode = errors.New("agent tool mode not supported")

// Request represents a generic agent tool request.
// It is intentionally permissive so that tool-specific options can be passed
// through without forcing a shared schema.
type Request struct {
	Prompt          string                 // Primary instruction/prompt for the tool
	Mode            string                 // Tool-dependent mode (e.g., "exec")
	Model           string                 // Optional model override
	Temperature     *float64               // Optional sampling temperature
	MaxTokens       *int                   // Optional token limit
	ReasoningEffort string                 // Reasoning effort level: "low", "medium", "high"
	Workdir         string                 // Optional working directory (tool specific)
	Timeout         time.Duration          // Optional timeout override
	ExtraEnv        map[string]string      // Additional environment variables
	Flags           []string               // Extra CLI flags to append
	ToolSpecific    map[string]interface{} // Bag for tool-specific parameters
	UseStdin        bool                   // If true, send prompt via stdin when supported
}

// ExecPlan is the resolved command plan produced by a provider.
type ExecPlan struct {
	Command string            // Binary name/path
	Args    []string          // Arguments to the binary
	Env     map[string]string // Environment variables to set for this invocation
	Workdir string            // Working directory (if supported by executor)
	Timeout time.Duration     // Optional timeout override
	Stdin   string            // Optional stdin content
}

// Capability describes high-level traits of a provider.
type Capability struct {
	Kind          string
	DefaultModel  string
	SupportsStdin bool
	Notes         string
}

// ProviderConfig describes how to construct a provider instance.
type ProviderConfig struct {
	Kind         string
	CLIPath      string
	Model        string
	SystemPrompt string
	ExtraEnv     map[string]string
	Flags        []string
	ToolSpecific map[string]interface{}
}

// AgentToolProvider resolves a Request into an ExecPlan for execution.
type AgentToolProvider interface {
	Kind() string
	Capabilities() Capability
	Build(ctx context.Context, req Request) (ExecPlan, error)
}

// mergeEnv copies maps and merges right over left.
func mergeEnv(base, override map[string]string) map[string]string {
	if len(base) == 0 && len(override) == 0 {
		return nil
	}
	out := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range override {
		out[k] = v
	}
	return out
}

// nonEmpty returns the first non-empty string.
func nonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ensurePrompt validates that a prompt is present.
func ensurePrompt(prompt string) error {
	if prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	return nil
}
