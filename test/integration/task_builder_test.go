package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/biwakonbu/agent-runner/internal/chat"
	"github.com/biwakonbu/agent-runner/internal/ide"
	"github.com/biwakonbu/agent-runner/internal/meta"
	"github.com/biwakonbu/agent-runner/internal/orchestrator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// MockMetaClient for GT-1
type MockMetaClient struct{}

func (m *MockMetaClient) Decompose(ctx context.Context, req *meta.DecomposeRequest) (*meta.DecomposeResponse, error) {
	// Simulate decomposing "TODO アプリを作成して"
	return &meta.DecomposeResponse{
		Understanding: "TODO アプリ作成の依頼を理解しました。",
		Phases: []meta.DecomposedPhase{
			{
				Name: "Implementation",
				Tasks: []meta.DecomposedTask{
					{
						ID:          "temp-task-1",
						Title:       "TODO アプリを作成して",
						Description: "TODO アプリを作成して。技術スタックや実装方針、検証方法はあなたの判断に任せます。",
						WBSLevel:    1,
						AcceptanceCriteria: []string{
							"アプリが起動すること",
						},
					},
				},
			},
		},
	}, nil
}

func TestTaskBuilder_Golden(t *testing.T) {
	// 1. Setup specific test workspace
	tempDir, err := os.MkdirTemp("", "multiverse-gt1-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	wsStore := ide.NewWorkspaceStore(filepath.Join(tempDir, "workspaces"))
	projectRoot := filepath.Join(tempDir, "project")
	err = os.MkdirAll(projectRoot, 0755)
	require.NoError(t, err)

	ws := &ide.Workspace{
		ProjectRoot: projectRoot,
		Version:     "1.0",
		DisplayName: "Golden Test 1",
	}
	err = wsStore.SaveWorkspace(ws)
	require.NoError(t, err)

	wsID := wsStore.GetWorkspaceID(projectRoot)
	wsDir := wsStore.GetWorkspaceDir(wsID)

	// 2. Setup Components
	taskStore := orchestrator.NewTaskStore(wsDir)
	sessionStore := chat.NewChatSessionStore(wsDir)
	metaClient := &MockMetaClient{}

	// Create a dummy event emitter (can be nil or mock if needed)
	handler := chat.NewHandler(metaClient, taskStore, sessionStore, wsID, projectRoot, nil)

	// 3. Setup Executor with Capture Script
	captureScriptPath := filepath.Join(tempDir, "capture_runner.sh")
	capturedYamlPath := filepath.Join(tempDir, "captured.yaml")

	// Script that writes stdin to captured.yaml
	scriptContent := fmt.Sprintf("#!/bin/sh\ncat > %s\nexit 0\n", capturedYamlPath)
	err = os.WriteFile(captureScriptPath, []byte(scriptContent), 0755)
	require.NoError(t, err)

	executor := orchestrator.NewExecutor(captureScriptPath, projectRoot)

	// 4. Run Chat Flow
	ctx := context.Background()
	session, err := handler.CreateSession(ctx)
	require.NoError(t, err)

	resp, err := handler.HandleMessage(ctx, session.ID, "TODO アプリを作成して")
	require.NoError(t, err)
	require.Len(t, resp.GeneratedTasks, 1)

	task := resp.GeneratedTasks[0]
	assert.Equal(t, "TODO アプリを作成して", task.Title)

	// 5. Run Executor (Simulate Orchestrator picking up the task)
	attempt, err := executor.ExecuteTask(ctx, &task)
	require.NoError(t, err)
	assert.Equal(t, orchestrator.AttemptStatusSucceeded, attempt.Status)

	// 6. Verify Captured YAML
	yamlBytes, err := os.ReadFile(capturedYamlPath)
	require.NoError(t, err, "captured.yaml should exist")

	var parsed map[string]interface{}
	err = yaml.Unmarshal(yamlBytes, &parsed)
	require.NoError(t, err, "Should be valid YAML")

	// Verify GT-1 Assertions
	// task.id, title, instructions (in prd.text), project.root_dir (repo)

	taskMap, ok := parsed["task"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, task.ID, taskMap["id"])
	assert.Equal(t, "TODO アプリを作成して", taskMap["title"])
	assert.Equal(t, ".", taskMap["repo"]) // Executor sets this to "."

	prdMap, ok := taskMap["prd"].(map[string]interface{})
	require.True(t, ok)
	prdText, ok := prdMap["text"].(string)
	require.True(t, ok)

	assert.Contains(t, prdText, "TODO アプリを作成して")
	assert.Contains(t, prdText, "Acceptance Criteria")
	assert.Contains(t, prdText, "アプリが起動すること")

	runnerMap, ok := parsed["runner"].(map[string]interface{})
	require.True(t, ok)
	workerMap, ok := runnerMap["worker"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "codex", workerMap["cli"])
}
