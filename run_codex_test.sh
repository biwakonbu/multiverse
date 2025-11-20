#!/bin/bash
# Script to run Codex integration test

echo "=== Codex Integration Test ==="
echo "Running agent-runner with test_codex_task.yaml"
echo ""

# Check if CODEX_API_KEY is set
if [ -z "$CODEX_API_KEY" ]; then
    echo "Warning: CODEX_API_KEY not set"
    echo "You may need to authenticate Codex first"
fi

# Run the agent-runner
go run cmd/agent-runner/main.go < test_codex_task.yaml

echo ""
echo "=== Test Complete ==="
echo "Check .agent-runner/task-TASK-CODEX-TEST.md for results"
