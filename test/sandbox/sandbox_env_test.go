//go:build docker
// +build docker

package sandbox

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/biwakonbu/agent-runner/internal/worker"
)

// TestSandboxManager_EnvPrefixHandling verifies that env: prefix is resolved
func TestSandboxManager_EnvPrefixHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker tests in short mode")
	}

	// Set a host environment variable
	hostVar := "HOST_TEST_VAR"
	hostVal := "resolved_value_123"
	os.Setenv(hostVar, hostVal)
	defer os.Unsetenv(hostVar)

	sm, err := worker.NewSandboxManager()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	image := "alpine:3.19"
	tmpDir := t.TempDir()

	// Pass env var with prefix
	env := map[string]string{
		"TARGET_VAR": "env:" + hostVar,
		"NORMAL_VAR": "normal_value",
	}

	containerID, err := sm.StartContainer(ctx, image, tmpDir, env)
	if err != nil {
		t.Fatalf("StartContainer() error = %v", err)
	}
	defer sm.StopContainer(ctx, containerID)

	// Check TARGET_VAR
	exitCode, output, err := sm.Exec(ctx, containerID, []string{"sh", "-c", "echo $TARGET_VAR"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if exitCode != 0 {
		t.Errorf("Exec() exit code = %d", exitCode)
	}
	if strings.TrimSpace(output) != hostVal {
		t.Errorf("Expected TARGET_VAR to be %q, got %q", hostVal, strings.TrimSpace(output))
	}

	// Check NORMAL_VAR
	exitCode, output, err = sm.Exec(ctx, containerID, []string{"sh", "-c", "echo $NORMAL_VAR"})
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}
	if strings.TrimSpace(output) != "normal_value" {
		t.Errorf("Expected NORMAL_VAR to be 'normal_value', got %q", strings.TrimSpace(output))
	}
}
