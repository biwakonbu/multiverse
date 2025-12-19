package meta

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/biwakonbu/agent-runner/internal/logging"
	"github.com/biwakonbu/agent-runner/internal/tooling"
	"github.com/biwakonbu/agent-runner/pkg/config"
)

// ToolingClient は tooling profile に基づいて meta プロバイダを選択する。
type ToolingClient struct {
	selector            *tooling.Selector
	fallback            *Client
	apiKey              string
	defaultSystemPrompt string
	logger              *slog.Logger
}

// NewToolingClient は tooling 対応の meta クライアントを作成する。
func NewToolingClient(cfg *config.ToolingConfig, apiKey string, fallback *Client, defaultSystemPrompt string) *ToolingClient {
	client := &ToolingClient{
		selector:            tooling.NewSelector(cfg),
		fallback:            fallback,
		apiKey:              apiKey,
		defaultSystemPrompt: defaultSystemPrompt,
		logger:              logging.WithComponent(slog.Default(), "meta-tooling"),
	}
	if fallback != nil {
		fallback.SetLogger(client.logger)
	}
	return client
}

// SetLogger はカスタムロガーを設定する。
func (c *ToolingClient) SetLogger(logger *slog.Logger) {
	c.logger = logging.WithComponent(logger, "meta-tooling")
	if c.fallback != nil {
		c.fallback.SetLogger(c.logger)
	}
}

func (c *ToolingClient) Decompose(ctx context.Context, req *DecomposeRequest) (*DecomposeResponse, error) {
	result, err := c.callWithCategory(ctx, tooling.CategoryPlan, func(client *Client) (interface{}, error) {
		return client.Decompose(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*DecomposeResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected decompose response type")
	}
	return resp, nil
}

func (c *ToolingClient) PlanPatch(ctx context.Context, req *PlanPatchRequest) (*PlanPatchResponse, error) {
	result, err := c.callWithCategory(ctx, tooling.CategoryPlan, func(client *Client) (interface{}, error) {
		return client.PlanPatch(ctx, req)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*PlanPatchResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected plan patch response type")
	}
	return resp, nil
}

func (c *ToolingClient) PlanTask(ctx context.Context, prdText string) (*PlanTaskResponse, error) {
	result, err := c.callWithCategory(ctx, tooling.CategoryTask, func(client *Client) (interface{}, error) {
		return client.PlanTask(ctx, prdText)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*PlanTaskResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected plan task response type")
	}
	return resp, nil
}

func (c *ToolingClient) NextAction(ctx context.Context, taskSummary *TaskSummary) (*NextActionResponse, error) {
	result, err := c.callWithCategory(ctx, tooling.CategoryExecution, func(client *Client) (interface{}, error) {
		return client.NextAction(ctx, taskSummary)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*NextActionResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected next action response type")
	}
	return resp, nil
}

func (c *ToolingClient) CompletionAssessment(ctx context.Context, taskSummary *TaskSummary) (*CompletionAssessmentResponse, error) {
	result, err := c.callWithCategory(ctx, tooling.CategoryExecution, func(client *Client) (interface{}, error) {
		return client.CompletionAssessment(ctx, taskSummary)
	})
	if err != nil {
		return nil, err
	}
	resp, ok := result.(*CompletionAssessmentResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected completion assessment response type")
	}
	return resp, nil
}

func (c *ToolingClient) callWithCategory(ctx context.Context, category string, call func(*Client) (interface{}, error)) (interface{}, error) {
	if c.selector == nil {
		return c.callFallback(ctx, category, call)
	}
	if forced, ok := c.selector.ForceCandidate(); ok {
		client := c.clientForCandidate(forced)
		return call(client)
	}

	categoryCfg, ok := c.selector.Category(category)
	if !ok || len(categoryCfg.Candidates) == 0 {
		return c.callFallback(ctx, category, call)
	}

	var lastErr error
	for i := 0; i < len(categoryCfg.Candidates); i++ {
		candidate, ok := c.selector.Select(category)
		if !ok {
			break
		}
		client := c.clientForCandidate(candidate)
		resp, err := call(client)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if tooling.IsRateLimitError(err) && c.selector.ShouldFallbackOnRateLimit(category) {
			c.selector.MarkRateLimited(category, candidate, c.selector.CooldownSec(category))
			c.logger.Warn("rate limited; switching tooling candidate",
				slog.String("category", category),
				slog.String("tool", candidate.Tool),
				slog.String("model", candidate.Model),
			)
			continue
		}
		return nil, err
	}

	if c.fallback != nil {
		c.logger.Warn("tooling candidate unavailable; using fallback",
			slog.String("category", category),
		)
		return call(c.fallback)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("tooling client: no available candidate for category %s", category)
}

func (c *ToolingClient) callFallback(_ context.Context, category string, call func(*Client) (interface{}, error)) (interface{}, error) {
	if c.fallback == nil {
		return nil, fmt.Errorf("tooling client: fallback is nil for category %s", category)
	}
	return call(c.fallback)
}

func (c *ToolingClient) clientForCandidate(candidate config.ToolCandidate) *Client {
	tool := strings.ToLower(strings.TrimSpace(candidate.Tool))
	systemPrompt := candidate.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = c.defaultSystemPrompt
	}

	switch tool {
	case "mock":
		client := NewMockClient()
		client.SetLogger(c.logger)
		return client
	case "openai-chat":
		client := NewClient("openai-chat", c.apiKey, candidate.Model, systemPrompt)
		client.SetLogger(c.logger)
		return client
	default:
		opts := CLIProviderOptions{
			CLIPath:      candidate.CLIPath,
			Flags:        candidate.Flags,
			Env:          candidate.Env,
			ToolSpecific: candidate.ToolSpecific,
		}
		provider := NewCLIProviderWithOptions(tool, candidate.Model, systemPrompt, opts)
		provider.SetLogger(c.logger)
		client := &Client{
			kind:         tool,
			apiKey:       c.apiKey,
			model:        candidate.Model,
			systemPrompt: systemPrompt,
			logger:       logging.WithComponent(c.logger, "meta-client"),
			provider:     provider,
		}
		return client
	}
}
