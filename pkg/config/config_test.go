package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTaskConfig_UnmarshalYAML_Valid(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    TaskConfig
		wantErr bool
	}{
		{
			name: "minimal valid config",
			yaml: `
version: 1
task:
  id: "TASK-001"
  repo: "."
  prd:
    text: "Simple requirement"
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
			want: TaskConfig{
				Version: 1,
				Task: TaskDetails{
					ID:   "TASK-001",
					Repo: ".",
					PRD: PRDDetails{
						Text: "Simple requirement",
					},
				},
				Runner: RunnerConfig{
					Meta: MetaConfig{
						Kind: "mock",
					},
					Worker: WorkerConfig{
						Kind: "codex-cli",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "complete config with all fields",
			yaml: `
version: 1
task:
  id: "TASK-002"
  title: "Test Task"
  repo: "/path/to/repo"
  prd:
    path: "./docs/prd.md"
  test:
    command: "npm test"
    cwd: "./"
runner:
  meta:
    kind: "openai-chat"
    model: "gpt-4-turbo"
  worker:
    kind: "codex-cli"
    docker_image: "agent-runner-codex:latest"
    max_run_time_sec: 1800
    env:
      CODEX_API_KEY: "env:CODEX_API_KEY"
      MY_VAR: "value"
`,
			want: TaskConfig{
				Version: 1,
				Task: TaskDetails{
					ID:    "TASK-002",
					Title: "Test Task",
					Repo:  "/path/to/repo",
					PRD: PRDDetails{
						Path: "./docs/prd.md",
					},
					Test: TestDetails{
						Command: "npm test",
						Cwd:     "./",
					},
				},
				Runner: RunnerConfig{
					Meta: MetaConfig{
						Kind:  "openai-chat",
						Model: "gpt-4-turbo",
					},
					Worker: WorkerConfig{
						Kind:          "codex-cli",
						DockerImage:   "agent-runner-codex:latest",
						MaxRunTimeSec: 1800,
						Env: map[string]string{
							"CODEX_API_KEY": "env:CODEX_API_KEY",
							"MY_VAR":        "value",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "PRD as text instead of path",
			yaml: `
version: 1
task:
  id: "TASK-003"
  repo: "."
  prd:
    text: |
      This is a multiline
      PRD specification
      for the task
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
			want: TaskConfig{
				Version: 1,
				Task: TaskDetails{
					ID:   "TASK-003",
					Repo: ".",
					PRD: PRDDetails{
						Text: "This is a multiline\nPRD specification\nfor the task\n",
					},
				},
				Runner: RunnerConfig{
					Meta: MetaConfig{
						Kind: "mock",
					},
					Worker: WorkerConfig{
						Kind: "codex-cli",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "both PRD path and text (path should be ignored)",
			yaml: `
version: 1
task:
  id: "TASK-004"
  repo: "."
  prd:
    path: "./docs/prd.md"
    text: "Direct text"
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
			want: TaskConfig{
				Version: 1,
				Task: TaskDetails{
					ID:   "TASK-004",
					Repo: ".",
					PRD: PRDDetails{
						Path: "./docs/prd.md",
						Text: "Direct text",
					},
				},
				Runner: RunnerConfig{
					Meta: MetaConfig{
						Kind: "mock",
					},
					Worker: WorkerConfig{
						Kind: "codex-cli",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg TaskConfig
			err := yaml.Unmarshal([]byte(tt.yaml), &cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if cfg.Version != tt.want.Version {
				t.Errorf("Version = %d, want %d", cfg.Version, tt.want.Version)
			}
			if cfg.Task.ID != tt.want.Task.ID {
				t.Errorf("Task.ID = %s, want %s", cfg.Task.ID, tt.want.Task.ID)
			}
			if cfg.Task.Title != tt.want.Task.Title {
				t.Errorf("Task.Title = %s, want %s", cfg.Task.Title, tt.want.Task.Title)
			}
			if cfg.Task.Repo != tt.want.Task.Repo {
				t.Errorf("Task.Repo = %s, want %s", cfg.Task.Repo, tt.want.Task.Repo)
			}
			if cfg.Task.PRD.Path != tt.want.Task.PRD.Path {
				t.Errorf("Task.PRD.Path = %s, want %s", cfg.Task.PRD.Path, tt.want.Task.PRD.Path)
			}
			if cfg.Task.PRD.Text != tt.want.Task.PRD.Text {
				t.Errorf("Task.PRD.Text = %s, want %s", cfg.Task.PRD.Text, tt.want.Task.PRD.Text)
			}
			if cfg.Runner.Meta.Kind != tt.want.Runner.Meta.Kind {
				t.Errorf("Runner.Meta.Kind = %s, want %s", cfg.Runner.Meta.Kind, tt.want.Runner.Meta.Kind)
			}
		})
	}
}

func TestTaskConfig_UnmarshalYAML_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		yamlStr string
		wantErr bool
	}{
		{
			name: "completely invalid YAML - unclosed quote",
			yamlStr: `
version: 1
task:
  repo: "."
  prd:
    text: "unclosed quote
runner:
`,
			wantErr: true,
		},
		{
			name: "missing required fields should still parse (empty values OK)",
			yamlStr: `
version: 1
task:
  repo: "."
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg TaskConfig
			err := yaml.Unmarshal([]byte(tt.yamlStr), &cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaskConfig_VersionCheck(t *testing.T) {
	yamlStr := `
version: 2
task:
  repo: "."
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`
	var cfg TaskConfig
	err := yaml.Unmarshal([]byte(yamlStr), &cfg)
	if err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	if cfg.Version != 2 {
		t.Errorf("Version = %d, want 2", cfg.Version)
	}
}

func TestTaskConfig_DefaultValues(t *testing.T) {
	yamlStr := `
version: 1
task:
  repo: "."
  prd:
    text: "req"
runner:
  meta:
    kind: "mock"
  worker:
    kind: "codex-cli"
`
	var cfg TaskConfig
	err := yaml.Unmarshal([]byte(yamlStr), &cfg)
	if err != nil {
		t.Fatalf("UnmarshalYAML() error = %v", err)
	}

	// Check that omitted fields have zero values
	if cfg.Task.ID != "" {
		t.Errorf("Task.ID should be empty, got %s", cfg.Task.ID)
	}
	if cfg.Task.Title != "" {
		t.Errorf("Task.Title should be empty, got %s", cfg.Task.Title)
	}
	if cfg.Runner.Worker.MaxRunTimeSec != 0 {
		t.Errorf("Worker.MaxRunTimeSec should be 0, got %d", cfg.Runner.Worker.MaxRunTimeSec)
	}
	if cfg.Runner.Worker.Env != nil {
		t.Errorf("Worker.Env should be nil when not provided, got %v", cfg.Runner.Worker.Env)
	}
}
