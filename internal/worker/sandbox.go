package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	Exec(ctx context.Context, containerID string, cmd []string, stdin io.Reader) (int, string, error)
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
	// Check if image exists and pull if missing
	_, _, err := s.cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		// Image doesn't exist, pull it
		pullReader, pullErr := s.cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if pullErr != nil {
			return "", fmt.Errorf("failed to pull image %s: %w", image, pullErr)
		}
		defer func() { _ = pullReader.Close() }()

		// Wait for pull to complete by reading all output
		_, copyErr := io.Copy(io.Discard, pullReader)
		if copyErr != nil {
			return "", fmt.Errorf("failed to complete image pull for %s: %w", image, copyErr)
		}
	}

	var envSlice []string
	var customAuthPath string

	for k, v := range env {
		if k == "__INTERNAL_CLAUDE_AUTH_PATH" {
			customAuthPath = v
			continue
		}
		val := v
		if len(v) > 4 && v[:4] == "env:" {
			val = os.Getenv(v[4:])
		}
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, val))
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

		// Mount Claude Code auth if it exists
		var claudeConfigPath string
		if customAuthPath != "" {
			if filepath.IsAbs(customAuthPath) {
				claudeConfigPath = customAuthPath
			} else {
				claudeConfigPath = filepath.Join(homeDir, customAuthPath)
			}
		} else {
			// Default path: ~/.config/claude
			claudeConfigPath = filepath.Join(homeDir, ".config", "claude")
		}

		if _, err := os.Stat(claudeConfigPath); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   claudeConfigPath,
				Target:   "/root/.config/claude",
				ReadOnly: true,
			})
		}

		// Mount Gemini CLI config if it exists (e.g. ~/.gemini/.env, settings.json)
		geminiConfigPath := filepath.Join(homeDir, ".gemini")
		if _, err := os.Stat(geminiConfigPath); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   geminiConfigPath,
				Target:   "/root/.gemini",
				ReadOnly: true,
			})
		}
	}

	// If auth.json doesn't exist, check for CODEX_API_KEY env var
	if codexAPIKey := os.Getenv("CODEX_API_KEY"); codexAPIKey != "" {
		envSlice = append(envSlice, fmt.Sprintf("CODEX_API_KEY=%s", codexAPIKey))
	}
	if geminiAPIKey := os.Getenv("GEMINI_API_KEY"); geminiAPIKey != "" {
		envSlice = append(envSlice, fmt.Sprintf("GEMINI_API_KEY=%s", geminiAPIKey))
	}
	if googleAPIKey := os.Getenv("GOOGLE_API_KEY"); googleAPIKey != "" {
		envSlice = append(envSlice, fmt.Sprintf("GOOGLE_API_KEY=%s", googleAPIKey))
	}
	if vertexFlag := os.Getenv("GOOGLE_GENAI_USE_VERTEXAI"); vertexFlag != "" {
		envSlice = append(envSlice, fmt.Sprintf("GOOGLE_GENAI_USE_VERTEXAI=%s", vertexFlag))
	}
	if projectID := os.Getenv("GOOGLE_CLOUD_PROJECT"); projectID != "" {
		envSlice = append(envSlice, fmt.Sprintf("GOOGLE_CLOUD_PROJECT=%s", projectID))
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

func (s *SandboxManager) Exec(ctx context.Context, containerID string, cmd []string, stdin io.Reader) (int, string, error) {
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  stdin != nil,
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

	// If stdin is provided, stream it to the container in a goroutine
	if stdin != nil {
		go func() {
			// Copy input to execution
			// We ignore error here because if the process exits, writes will fail
			defer func() {
				_ = hijacked.CloseWrite()
			}()
			_, _ = io.Copy(hijacked.Conn, stdin)
		}()
	}

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
