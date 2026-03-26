package tui_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	ghclient "ghamon/internal/github"
	"ghamon/internal/tui"
)

// MockGHClient is a mock of ghclient.Client.
type MockGHClient struct {
	mock.Mock
}

func (m *MockGHClient) GetWorkflowStatuses(ctx context.Context, owner, repo, workflowFile string) ([]ghclient.WorkflowRun, error) {
	args := m.Called(ctx, owner, repo, workflowFile)
	runs, _ := args.Get(0).([]ghclient.WorkflowRun)
	return runs, args.Error(1)
}

func TestNew_FieldsSet(t *testing.T) {
	repos := []string{"owner/repo1", "owner/repo2"}
	m := tui.New(repos, "ci.yml", 30, &MockGHClient{})
	assert.Equal(t, "ci.yml", m.Workflow)
	assert.Equal(t, 30, m.Rate)
}

func TestModel_View_NotReady(t *testing.T) {
	m := tui.New([]string{}, "", 30, &MockGHClient{})
	assert.Contains(t, m.View(), "Initializing")
}

func TestModel_Update_WindowSize(t *testing.T) {
	m := tui.New([]string{"owner/repo"}, "ci.yml", 60, &MockGHClient{})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	_, ok := updated.(tui.Model)
	assert.True(t, ok)
}

func TestModel_Update_QuitKey(t *testing.T) {
	m := tui.New([]string{"owner/repo"}, "", 30, &MockGHClient{})
	// Make the model ready first.
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	// Send quit key.
	_, cmd := m2.(tui.Model).Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	assert.NotNil(t, cmd)
}
