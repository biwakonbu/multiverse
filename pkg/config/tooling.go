package config

// ToolingConfig defines profile-based tool selection for meta/worker execution.
type ToolingConfig struct {
	ActiveProfile string        `yaml:"active_profile,omitempty" json:"activeProfile,omitempty"`
	Profiles      []ToolProfile `yaml:"profiles,omitempty" json:"profiles,omitempty"`
	Force         ToolForce     `yaml:"force,omitempty" json:"force,omitempty"`
}

// ToolForce overrides all category selections when enabled.
type ToolForce struct {
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Tool    string `yaml:"tool,omitempty" json:"tool,omitempty"`
	Model   string `yaml:"model,omitempty" json:"model,omitempty"`
}

// ToolProfile groups category configs under a named profile.
type ToolProfile struct {
	ID         string                        `yaml:"id,omitempty" json:"id,omitempty"`
	Name       string                        `yaml:"name,omitempty" json:"name,omitempty"`
	Categories map[string]ToolCategoryConfig `yaml:"categories,omitempty" json:"categories,omitempty"`
}

// ToolCategoryConfig defines weighted tool selection for a category.
type ToolCategoryConfig struct {
	Strategy            string          `yaml:"strategy,omitempty" json:"strategy,omitempty"` // weighted | round_robin
	Candidates          []ToolCandidate `yaml:"candidates,omitempty" json:"candidates,omitempty"`
	FallbackOnRateLimit bool            `yaml:"fallback_on_rate_limit,omitempty" json:"fallbackOnRateLimit,omitempty"`
	CooldownSec         int             `yaml:"cooldown_sec,omitempty" json:"cooldownSec,omitempty"`
}

// ToolCandidate defines a tool/model pair and optional execution overrides.
type ToolCandidate struct {
	Tool         string                 `yaml:"tool,omitempty" json:"tool,omitempty"` // codex-cli | claude-code | gemini-cli | openai-chat | mock
	Model        string                 `yaml:"model,omitempty" json:"model,omitempty"`
	Weight       int                    `yaml:"weight,omitempty" json:"weight,omitempty"`
	CLIPath      string                 `yaml:"cli_path,omitempty" json:"cliPath,omitempty"`
	Flags        []string               `yaml:"flags,omitempty" json:"flags,omitempty"`
	Env          map[string]string      `yaml:"env,omitempty" json:"env,omitempty"`
	ToolSpecific map[string]interface{} `yaml:"tool_specific,omitempty" json:"toolSpecific,omitempty"`
	SystemPrompt string                 `yaml:"system_prompt,omitempty" json:"systemPrompt,omitempty"`
}
