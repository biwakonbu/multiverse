package note

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/biwakonbu/agent-runner/internal/core"
)

type Writer struct{}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(taskCtx *core.TaskContext) error {
	// Ensure .agent-runner directory exists
	dir := filepath.Join(taskCtx.RepoPath, ".agent-runner")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filename := fmt.Sprintf("task-%s.md", taskCtx.ID)
	path := filepath.Join(dir, filename)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	tmplStr := `
# Task Note - {{ .ID }} {{ if .Title }}- {{ .Title }}{{ end }}

- Task ID: {{ .ID }}
- Title: {{ .Title }}
- Started At: {{ .StartedAt }}
- Finished At: {{ .FinishedAt }}
- State: {{ .State }}

---

## 1. PRD Summary

<details>
<summary>PRD Text</summary>

` + "```" + `text
{{ .PRDText }}
` + "```" + `

</details>

---

## 2. Acceptance Criteria

{{ range .AcceptanceCriteria }}
- [{{ if .Passed }}x{{ else }} {{ end }}] {{ .ID }}: {{ .Description }}
{{ end }}

---

## 3. Execution Log

### 3.1 Meta Calls

{{ range .MetaCalls }}
#### {{ .Type }} at {{ .Timestamp }}

` + "```" + `yaml
{{ .RequestYAML }}
` + "```" + `

` + "```" + `yaml
{{ .ResponseYAML }}
` + "```" + `

{{ end }}

### 3.2 Worker Runs

{{ range .WorkerRuns }}
#### Run {{ .ID }} (ExitCode={{ .ExitCode }}) at {{ .StartedAt }}

Summary: {{ .Summary }}

` + "```" + `text
{{ .RawOutput }}
` + "```" + `

{{ end }}

---
`

	tmpl, err := template.New("task-note").Parse(tmplStr)
	if err != nil {
		return err
	}

	return tmpl.Execute(f, taskCtx)
}
