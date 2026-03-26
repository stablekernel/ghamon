package github_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ghclient "ghamon/internal/github"
)

// MockClient is a mock implementation of the Client interface.
type MockClient struct {
	mock.Mock
}

func (m *MockClient) GetWorkflowStatuses(ctx context.Context, owner, repo, workflowFile string) ([]ghclient.WorkflowRun, error) {
	args := m.Called(ctx, owner, repo, workflowFile)
	runs, _ := args.Get(0).([]ghclient.WorkflowRun)
	return runs, args.Error(1)
}

func TestMockClient_Success(t *testing.T) {
	m := &MockClient{}
	now := time.Now()
	expected := []ghclient.WorkflowRun{
		{
			Repo:       "owner/repo",
			Workflow:   "ci.yml",
			Status:     "completed",
			Conclusion: "success",
			UpdatedAt:  now,
			URL:        "https://github.com/owner/repo/actions/runs/1",
		},
	}
	m.On("GetWorkflowStatuses", mock.Anything, "owner", "repo", "ci.yml").Return(expected, nil)

	runs, err := m.GetWorkflowStatuses(context.Background(), "owner", "repo", "ci.yml")
	require.NoError(t, err)
	assert.Equal(t, expected, runs)
	m.AssertExpectations(t)
}

func TestMockClient_Error(t *testing.T) {
	m := &MockClient{}
	m.On("GetWorkflowStatuses", mock.Anything, "owner", "repo", "").Return(nil, assert.AnError)

	runs, err := m.GetWorkflowStatuses(context.Background(), "owner", "repo", "")
	assert.Error(t, err)
	assert.Nil(t, runs)
	m.AssertExpectations(t)
}

func TestWorkflowRun_DisplayStatus(t *testing.T) {
	tests := []struct {
		name string
		run  ghclient.WorkflowRun
		want string
	}{
		{"completed_success", ghclient.WorkflowRun{Status: "completed", Conclusion: "success"}, "success"},
		{"completed_failure", ghclient.WorkflowRun{Status: "completed", Conclusion: "failure"}, "failure"},
		{"completed_cancelled", ghclient.WorkflowRun{Status: "completed", Conclusion: "cancelled"}, "cancelled"},
		{"in_progress", ghclient.WorkflowRun{Status: "in_progress"}, "in progress"},
		{"queued", ghclient.WorkflowRun{Status: "queued"}, "queued"},
		{"completed_no_conc", ghclient.WorkflowRun{Status: "completed"}, "completed"},
		{"empty_status", ghclient.WorkflowRun{}, "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.run.DisplayStatus())
		})
	}
}
