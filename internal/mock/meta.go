package mock

import (
	"context"

	"github.com/biwakonbu/agent-runner/internal/meta"
)

type MetaClient struct {
	PlanTaskFunc              func(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error)
	NextActionFunc            func(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error)
	CompletionAssessmentFunc  func(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error)
}

func (m *MetaClient) PlanTask(ctx context.Context, prdText string) (*meta.PlanTaskResponse, error) {
	if m.PlanTaskFunc != nil {
		return m.PlanTaskFunc(ctx, prdText)
	}
	return nil, nil
}

func (m *MetaClient) NextAction(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.NextActionResponse, error) {
	if m.NextActionFunc != nil {
		return m.NextActionFunc(ctx, taskSummary)
	}
	return nil, nil
}

func (m *MetaClient) CompletionAssessment(ctx context.Context, taskSummary *meta.TaskSummary) (*meta.CompletionAssessmentResponse, error) {
	if m.CompletionAssessmentFunc != nil {
		return m.CompletionAssessmentFunc(ctx, taskSummary)
	}
	return nil, nil
}

// NewMockMetaClient creates a mock MetaClient with default behavior
func NewMockMetaClient() *MetaClient {
	return &MetaClient{}
}
