package meta

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPlanTaskResponse_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    PlanTaskResponse
		wantErr bool
	}{
		{
			name: "minimal plan task response",
			yaml: `
task_id: "TASK-001"
acceptance_criteria: []
`,
			want: PlanTaskResponse{
				TaskID:             "TASK-001",
				AcceptanceCriteria: []AcceptanceCriterion{},
			},
			wantErr: false,
		},
		{
			name: "plan task response with criteria",
			yaml: `
task_id: "TASK-002"
acceptance_criteria:
  - id: "AC-1"
    description: "User can create account"
    type: "functional"
    critical: true
    passed: false
  - id: "AC-2"
    description: "Password validation works"
    type: "functional"
    critical: true
    passed: false
`,
			want: PlanTaskResponse{
				TaskID: "TASK-002",
				AcceptanceCriteria: []AcceptanceCriterion{
					{
						ID:          "AC-1",
						Description: "User can create account",
						Type:        "functional",
						Critical:    true,
						Passed:      false,
					},
					{
						ID:          "AC-2",
						Description: "Password validation works",
						Type:        "functional",
						Critical:    true,
						Passed:      false,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "criteria with optional fields",
			yaml: `
task_id: "TASK-003"
acceptance_criteria:
  - id: "AC-1"
    description: "Minimal criterion"
  - id: "AC-2"
    description: "With type"
    type: "non-functional"
`,
			want: PlanTaskResponse{
				TaskID: "TASK-003",
				AcceptanceCriteria: []AcceptanceCriterion{
					{
						ID:          "AC-1",
						Description: "Minimal criterion",
						Type:        "",
						Critical:    false,
						Passed:      false,
					},
					{
						ID:          "AC-2",
						Description: "With type",
						Type:        "non-functional",
						Critical:    false,
						Passed:      false,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp PlanTaskResponse
			err := yaml.Unmarshal([]byte(tt.yaml), &resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if resp.TaskID != tt.want.TaskID {
				t.Errorf("TaskID = %s, want %s", resp.TaskID, tt.want.TaskID)
			}
			if len(resp.AcceptanceCriteria) != len(tt.want.AcceptanceCriteria) {
				t.Errorf("AcceptanceCriteria length = %d, want %d", len(resp.AcceptanceCriteria), len(tt.want.AcceptanceCriteria))
			}
			for i, ac := range resp.AcceptanceCriteria {
				if ac.ID != tt.want.AcceptanceCriteria[i].ID {
					t.Errorf("AcceptanceCriteria[%d].ID = %s, want %s", i, ac.ID, tt.want.AcceptanceCriteria[i].ID)
				}
				if ac.Description != tt.want.AcceptanceCriteria[i].Description {
					t.Errorf("AcceptanceCriteria[%d].Description = %s, want %s", i, ac.Description, tt.want.AcceptanceCriteria[i].Description)
				}
				if ac.Critical != tt.want.AcceptanceCriteria[i].Critical {
					t.Errorf("AcceptanceCriteria[%d].Critical = %v, want %v", i, ac.Critical, tt.want.AcceptanceCriteria[i].Critical)
				}
			}
		})
	}
}

func TestNextActionResponse_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    NextActionResponse
		wantErr bool
	}{
		{
			name: "run_worker action",
			yaml: `
decision:
  action: "run_worker"
  reason: "Starting development"
worker_call:
  worker_type: "codex-cli"
  mode: "workspace-write"
  prompt: "Create a simple calculator"
`,
			want: NextActionResponse{
				Decision: Decision{
					Action: "run_worker",
					Reason: "Starting development",
				},
				WorkerCall: WorkerCall{
					WorkerType: "codex-cli",
					Mode:       "workspace-write",
					Prompt:     "Create a simple calculator",
				},
			},
			wantErr: false,
		},
		{
			name: "mark_complete action without worker call",
			yaml: `
decision:
  action: "mark_complete"
  reason: "All criteria satisfied"
`,
			want: NextActionResponse{
				Decision: Decision{
					Action: "mark_complete",
					Reason: "All criteria satisfied",
				},
				WorkerCall: WorkerCall{},
			},
			wantErr: false,
		},
		{
			name: "abort action",
			yaml: `
decision:
  action: "abort"
  reason: "Task failed - impossible requirement"
`,
			want: NextActionResponse{
				Decision: Decision{
					Action: "abort",
					Reason: "Task failed - impossible requirement",
				},
			},
			wantErr: false,
		},
		{
			name: "ask_human action",
			yaml: `
decision:
  action: "ask_human"
  reason: "Unclear requirement, need clarification"
`,
			want: NextActionResponse{
				Decision: Decision{
					Action: "ask_human",
					Reason: "Unclear requirement, need clarification",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp NextActionResponse
			err := yaml.Unmarshal([]byte(tt.yaml), &resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if resp.Decision.Action != tt.want.Decision.Action {
				t.Errorf("Decision.Action = %s, want %s", resp.Decision.Action, tt.want.Decision.Action)
			}
			if resp.Decision.Reason != tt.want.Decision.Reason {
				t.Errorf("Decision.Reason = %s, want %s", resp.Decision.Reason, tt.want.Decision.Reason)
			}
			if resp.WorkerCall.WorkerType != tt.want.WorkerCall.WorkerType {
				t.Errorf("WorkerCall.WorkerType = %s, want %s", resp.WorkerCall.WorkerType, tt.want.WorkerCall.WorkerType)
			}
		})
	}
}

func TestCompletionAssessmentResponse_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    CompletionAssessmentResponse
		wantErr bool
	}{
		{
			name: "all criteria passed",
			yaml: `
all_criteria_satisfied: true
summary: "All acceptance criteria met"
by_criterion:
  - id: "AC-1"
    status: "passed"
    comment: "Successfully implemented"
  - id: "AC-2"
    status: "passed"
    comment: "Works as expected"
`,
			want: CompletionAssessmentResponse{
				AllCriteriaSatisfied: true,
				Summary:              "All acceptance criteria met",
				ByCriterion: []CriterionResult{
					{
						ID:      "AC-1",
						Status:  "passed",
						Comment: "Successfully implemented",
					},
					{
						ID:      "AC-2",
						Status:  "passed",
						Comment: "Works as expected",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "some criteria failed",
			yaml: `
all_criteria_satisfied: false
summary: "AC-2 failed to meet requirement"
by_criterion:
  - id: "AC-1"
    status: "passed"
    comment: "Done"
  - id: "AC-2"
    status: "failed"
    comment: "Timeout during execution"
`,
			want: CompletionAssessmentResponse{
				AllCriteriaSatisfied: false,
				Summary:              "AC-2 failed to meet requirement",
				ByCriterion: []CriterionResult{
					{
						ID:      "AC-1",
						Status:  "passed",
						Comment: "Done",
					},
					{
						ID:      "AC-2",
						Status:  "failed",
						Comment: "Timeout during execution",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no criteria",
			yaml: `
all_criteria_satisfied: true
summary: "No criteria to assess"
by_criterion: []
`,
			want: CompletionAssessmentResponse{
				AllCriteriaSatisfied: true,
				Summary:              "No criteria to assess",
				ByCriterion:          []CriterionResult{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp CompletionAssessmentResponse
			err := yaml.Unmarshal([]byte(tt.yaml), &resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}

			if resp.AllCriteriaSatisfied != tt.want.AllCriteriaSatisfied {
				t.Errorf("AllCriteriaSatisfied = %v, want %v", resp.AllCriteriaSatisfied, tt.want.AllCriteriaSatisfied)
			}
			if resp.Summary != tt.want.Summary {
				t.Errorf("Summary = %s, want %s", resp.Summary, tt.want.Summary)
			}
			if len(resp.ByCriterion) != len(tt.want.ByCriterion) {
				t.Errorf("ByCriterion length = %d, want %d", len(resp.ByCriterion), len(tt.want.ByCriterion))
			}
		})
	}
}

func TestTaskSummary_Creation(t *testing.T) {
	summary := TaskSummary{
		Title: "Test Task",
		State: "RUNNING",
		AcceptanceCriteria: []AcceptanceCriterion{
			{
				ID:          "AC-1",
				Description: "Criterion 1",
				Critical:    true,
				Passed:      false,
			},
		},
		WorkerRunsCount: 1,
	}

	if summary.Title != "Test Task" {
		t.Errorf("Title = %s, want 'Test Task'", summary.Title)
	}
	if summary.State != "RUNNING" {
		t.Errorf("State = %s, want 'RUNNING'", summary.State)
	}
	if summary.WorkerRunsCount != 1 {
		t.Errorf("WorkerRunsCount = %d, want 1", summary.WorkerRunsCount)
	}
}
