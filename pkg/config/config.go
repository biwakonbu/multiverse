package config

// TaskConfig represents the root configuration from task.yaml
type TaskConfig struct {
	Version int          `yaml:"version"`
	Task    TaskDetails  `yaml:"task"`
	Runner  RunnerConfig `yaml:"runner"`
}

// TaskDetails holds task-specific information
type TaskDetails struct {
	ID    string      `yaml:"id"`
	Title string      `yaml:"title"`
	Repo  string      `yaml:"repo"`
	PRD   PRDDetails  `yaml:"prd"`
	Test  TestDetails `yaml:"test"`
}

// PRDDetails holds PRD location or content
type PRDDetails struct {
	Path string `yaml:"path"`
	Text string `yaml:"text"`
}

// TestDetails holds test configuration
type TestDetails struct {
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
}

// RunnerConfig holds runner configuration
type RunnerConfig struct {
	Meta   MetaConfig   `yaml:"meta"`
	Worker WorkerConfig `yaml:"worker"`
}

// MetaConfig holds Meta agent configuration
type MetaConfig struct {
	Kind  string `yaml:"kind"`
	Model string `yaml:"model"`
}

// WorkerConfig holds Worker agent configuration
type WorkerConfig struct {
	Kind          string            `yaml:"kind"`
	DockerImage   string            `yaml:"docker_image"`
	MaxRunTimeSec int               `yaml:"max_run_time_sec"`
	Env           map[string]string `yaml:"env"`
}
