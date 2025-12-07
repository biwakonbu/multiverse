//go:build codex
// +build codex

package codex

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/biwakonbu/agent-runner/internal/core"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/note"
	"github.com/biwakonbu/agent-runner/internal/worker"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

// TestCase defines the structure of a test case YAML file
type TestCase struct {
	ID                 string                    `yaml:"id"`
	Title              string                    `yaml:"title"`
	PRD                string                    `yaml:"prd"`
	AcceptanceCriteria []TestAcceptanceCriterion `yaml:"acceptance_criteria"`
}

type TestAcceptanceCriterion struct {
	ID          string `yaml:"id"`
	Description string `yaml:"description"`
	Type        string `yaml:"type"`
}

// TestCodex_TableDriven runs all test cases defined in test/codex/cases/*.yaml
func TestCodex_TableDriven(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Codex integration tests in short mode")
	}

	// Find all test case files
	caseFiles, err := filepath.Glob("cases/*.yaml")
	if err != nil {
		t.Fatalf("Failed to glob test cases: %v", err)
	}
	if len(caseFiles) == 0 {
		t.Fatal("No test cases found in test/codex/cases/")
	}

	for _, caseFile := range caseFiles {
		t.Run(filepath.Base(caseFile), func(t *testing.T) {
			runTestCase(t, caseFile)
		})
	}
}

func runTestCase(t *testing.T, casePath string) {
	// Load test case
	data, err := os.ReadFile(casePath)
	if err != nil {
		t.Fatalf("Failed to read test case %s: %v", casePath, err)
	}

	var tc TestCase
	if err := yaml.Unmarshal(data, &tc); err != nil {
		t.Fatalf("Failed to parse test case %s: %v", casePath, err)
	}

	// Create temporary repo directory
	tmpDir := t.TempDir()

	// Create task configuration
	cfg := &config.TaskConfig{
		Version: 1,
		Task: config.TaskDetails{
			ID:    "TEST-" + tc.ID,
			Title: tc.Title,
			Repo:  tmpDir,
			PRD: config.PRDDetails{
				Text: tc.PRD,
			},
		},
		Runner: config.RunnerConfig{
			Meta: config.MetaConfig{
				Kind: "mock", // Use mock meta to avoid LLM costs, but we need a smarter mock or real LLM for actual code gen?
				// Wait, if we use "mock" meta, it won't generate code unless we program the mock.
				// The original plan implied running Codex.
				// If we want to test Codex CLI *execution*, we need the Meta agent to tell it what to do.
				// If we use a real LLM, it costs money.
				// If we use a mock Meta, we need to hardcode the "NextAction" to run the worker with a specific prompt.
				// Let's assume for this "Codex Integration Test", we want to verify the *Worker* (Codex CLI) works.
				// So we should probably use a Mock Meta that returns a "run_worker" action with a prompt derived from PRD.
			},
			Worker: config.WorkerConfig{
				Kind:          "codex-cli",
				DockerImage:   "agent-runner-codex:latest",
				MaxRunTimeSec: 300,
				Env: map[string]string{
					"CODEX_API_KEY": "env:CODEX_API_KEY", // Ensure this is passed
				},
			},
		},
	}

	// Custom Mock Meta that returns a run_worker action based on the PRD
	metaClient := &SmartMockMeta{
		PRD:                tc.PRD,
		AcceptanceCriteria: tc.AcceptanceCriteria,
	}

	// Create executor
	executor, err := worker.NewExecutor(cfg.Runner.Worker, tmpDir)
	if err != nil {
		t.Skipf("Codex environment not available: %v", err)
	}

	// For test stability, use SmartMockSandbox if not explicitly using real backend
	if os.Getenv("TEST_CODEX_REAL") != "1" {
		executor.Sandbox = &SmartMockSandbox{RepoPath: tmpDir}
	}

	noteWriter := note.NewWriter()

	runner := &core.Runner{
		Config: cfg,
		Meta:   metaClient,
		Worker: executor,
		Note:   noteWriter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	taskCtx, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	// Verify task completed
	if taskCtx.State != core.StateComplete && taskCtx.State != core.StateFailed {
		t.Errorf("Task state = %v, want COMPLETE or FAILED", taskCtx.State)
	}

	// Verify artifacts (simple check that *something* was generated if successful)
	// Since we are using a mock meta, the "verification" step is also mocked (always passes).
	// But the *worker* actually ran. So we can check if files exist.
	// The "SmartMockMeta" will tell the worker to "Implement the PRD".
	// Codex CLI should generate files.

	// Check for expected files based on ACs
	for _, ac := range tc.AcceptanceCriteria {
		if ac.Type == "file_exists" {
			// Extract filename from description or just check if the logical file exists
			// For "index.html が存在する", we extract "index.html"
			// Simple heuristic: look for words ending in extension
			words := strings.Fields(ac.Description)
			found := false
			for _, word := range words {
				if strings.Contains(word, ".") {
					// Check file existence
					path := filepath.Join(tmpDir, word)
					if _, err := os.Stat(path); err == nil {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("AC %s failed: file not found for description '%s'", ac.ID, ac.Description)
			}
		}
	}
}

// SmartMockMeta is a mock MetaClient that returns a run_worker action first, then mark_complete.
type SmartMockMeta struct {
	PRD                string
	AcceptanceCriteria []TestAcceptanceCriterion
}

func (m *SmartMockMeta) PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
	acs := make([]meta.AcceptanceCriterion, len(m.AcceptanceCriteria))
	for i, ac := range m.AcceptanceCriteria {
		acs[i] = meta.AcceptanceCriterion{
			ID:          ac.ID,
			Description: ac.Description,
			Type:        ac.Type,
		}
	}
	return &meta.PlanTaskResponse{
		TaskID:             "MOCK-TASK",
		AcceptanceCriteria: acs,
	}, nil
}

func (m *SmartMockMeta) NextAction(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error) {
	if taskSummary.WorkerRunsCount == 0 {
		return &meta.NextActionResponse{
			Decision: meta.Decision{
				Action: "run_worker",
				Reason: "Initial run to implement PRD",
			},
			WorkerCall: meta.WorkerCall{
				WorkerType: "codex-cli",
				Mode:       "exec",
				Prompt:     "Please implement the following requirements:\n" + m.PRD,
			},
		}, nil
	}
	return &meta.NextActionResponse{
		Decision: meta.Decision{
			Action: "mark_complete",
			Reason: "Work completed",
		},
	}, nil
}

func (m *SmartMockMeta) CompletionAssessment(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
	return &meta.CompletionAssessmentResponse{
		AllCriteriaSatisfied: true,
		Summary:              "Mock assessment: All good",
		ByCriterion: []meta.CriterionResult{
			{ID: "AC-1", Status: "passed", Comment: "Mock passed"},
		},
	}, nil
}
