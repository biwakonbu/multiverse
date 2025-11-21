package worker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// SandboxProvider defines the interface for sandbox management
type SandboxProvider interface {
	StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error)
	Exec(ctx context.Context, containerID string, cmd []string) (int, string, error)
	StopContainer(ctx context.Context, containerID string) error
}

type SandboxManager struct {
	cli *client.Client
}

func NewSandboxManager() (*SandboxManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &SandboxManager{cli: cli}, nil
}

func (s *SandboxManager) StartContainer(ctx context.Context, image string, repoPath string, env map[string]string) (string, error) {
	// Pull image (simplified, assuming it exists or pull if missing)
	// _, err := s.cli.ImagePull(ctx, image, types.ImagePullOptions{})
	// if err != nil { ... }

	var envSlice []string
	for k, v := range env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// Prepare mounts
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: repoPath,
			Target: "/workspace/project",
		},
	}

	// Mount Codex auth.json if it exists
	// Check for ~/.codex/auth.json
	homeDir, err := os.UserHomeDir()
	if err == nil {
		codexAuthPath := filepath.Join(homeDir, ".codex", "auth.json")
		if _, err := os.Stat(codexAuthPath); err == nil {
			// auth.json exists, mount it
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   codexAuthPath,
				Target:   "/root/.codex/auth.json",
				ReadOnly: true,
			})
		}
	}

	// If auth.json doesn't exist, check for CODEX_API_KEY env var
	if codexAPIKey := os.Getenv("CODEX_API_KEY"); codexAPIKey != "" {
		envSlice = append(envSlice, fmt.Sprintf("CODEX_API_KEY=%s", codexAPIKey))
	}

	resp, err := s.cli.ContainerCreate(ctx, &container.Config{
		Image:      image,
		Tty:        true, // Keep running
		Env:        envSlice,
		Cmd:        []string{"tail", "-f", "/dev/null"}, // Keep alive
		WorkingDir: "/workspace/project",
	}, &container.HostConfig{
		Mounts: mounts,
	}, nil, nil, "")
	if err != nil {
		return "", err
	}

	if err := s.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (s *SandboxManager) Exec(ctx context.Context, containerID string, cmd []string) (int, string, error) {
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false, // Use false to separate stdout/stderr if needed, but true is easier for reading.
		// Let's use false and stdcopy to be robust.
		WorkingDir: "/workspace/project",
	}

	resp, err := s.cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		return 0, "", err
	}

	hijacked, err := s.cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, "", err
	}
	defer hijacked.Close()

	var outBuf, errBuf bytes.Buffer
	// Copy output
	// This blocks until the stream is closed (command finishes)
	_, err = stdcopy.StdCopy(&outBuf, &errBuf, hijacked.Reader)
	if err != nil {
		// It might be that Tty=true was used? No, we set false.
		// If it fails, maybe just read all?
		// For now, return error.
		return 0, "", fmt.Errorf("failed to copy output: %w", err)
	}

	// Inspect to get exit code
	inspectResp, err := s.cli.ContainerExecInspect(ctx, resp.ID)
	if err != nil {
		return 0, "", err
	}

	output := outBuf.String() + "\n" + errBuf.String()
	return inspectResp.ExitCode, output, nil
}

func (s *SandboxManager) StopContainer(ctx context.Context, containerID string) error {
	timeout := 0 // Force kill
	return s.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}
