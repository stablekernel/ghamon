package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchWorkflowRuns(t *testing.T) {
	t.Run("returns most recent run per distinct workflow", func(t *testing.T) {
		response := workflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{WorkflowID: 1, Name: "CI", Status: "completed", Conclusion: "success"},
				{WorkflowID: 2, Name: "Deploy", Status: "in_progress", Conclusion: ""},
				{WorkflowID: 1, Name: "CI", Status: "completed", Conclusion: "failure"},
				{WorkflowID: 3, Name: "Lint", Status: "completed", Conclusion: "success"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{HTTPClient: server.Client(), Token: "test-token", BaseURL: server.URL}
		runs, err := client.FetchWorkflowRuns("owner/repo")
		require.NoError(t, err)
		require.Len(t, runs, 3)
		assert.Equal(t, "CI", runs[0].Name)
		assert.Equal(t, "success", runs[0].Conclusion)
		assert.Equal(t, "Deploy", runs[1].Name)
		assert.Equal(t, "Lint", runs[2].Name)
	})

	t.Run("deduplicates workflow name and file path for same workflow_id", func(t *testing.T) {
		response := workflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{WorkflowID: 1, Name: "Build", Status: "completed", Conclusion: "success"},
				{WorkflowID: 1, Name: ".github/workflows/build.yaml", Status: "completed", Conclusion: "failure"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{HTTPClient: server.Client(), Token: "test-token", BaseURL: server.URL}
		runs, err := client.FetchWorkflowRuns("owner/repo")
		require.NoError(t, err)
		require.Len(t, runs, 1)
		assert.Equal(t, "Build", runs[0].Name)
		assert.Equal(t, "success", runs[0].Conclusion)
	})

	t.Run("returns empty slice when no runs exist", func(t *testing.T) {
		response := workflowRunsResponse{WorkflowRuns: []WorkflowRun{}}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{HTTPClient: server.Client(), Token: "test-token", BaseURL: server.URL}
		runs, err := client.FetchWorkflowRuns("owner/repo")
		require.NoError(t, err)
		assert.Empty(t, runs)
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := &GitHubClient{HTTPClient: server.Client(), Token: "test-token", BaseURL: server.URL}
		_, err := client.FetchWorkflowRuns("owner/repo")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestFetchWorkflowRun(t *testing.T) {
	t.Run("returns the most recent matching workflow run", func(t *testing.T) {
		response := workflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{Name: "CI", Status: "completed", Conclusion: "success"},
				{Name: "CI", Status: "completed", Conclusion: "failure"},
				{Name: "Deploy", Status: "in_progress", Conclusion: ""},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/repos/owner/repo/actions/runs", r.URL.Path)
			assert.Equal(t, "50", r.URL.Query().Get("per_page"))
			assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{
			HTTPClient: server.Client(),
			Token:      "test-token",
			BaseURL:    server.URL,
		}

		run, err := client.FetchWorkflowRun("owner/repo", "CI")
		require.NoError(t, err)
		require.NotNil(t, run)
		assert.Equal(t, "CI", run.Name)
		assert.Equal(t, "success", run.Conclusion)
	})

	t.Run("returns nil when workflow not found", func(t *testing.T) {
		response := workflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{Name: "CI", Status: "completed", Conclusion: "success"},
			},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{
			HTTPClient: server.Client(),
			Token:      "test-token",
			BaseURL:    server.URL,
		}

		run, err := client.FetchWorkflowRun("owner/repo", "Deploy")
		require.NoError(t, err)
		assert.Nil(t, run)
	})

	t.Run("returns error on non-200 response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client := &GitHubClient{
			HTTPClient: server.Client(),
			Token:      "test-token",
			BaseURL:    server.URL,
		}

		_, err := client.FetchWorkflowRun("owner/repo", "CI")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("returns nil when no runs exist", func(t *testing.T) {
		response := workflowRunsResponse{
			WorkflowRuns: []WorkflowRun{},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := &GitHubClient{
			HTTPClient: server.Client(),
			Token:      "test-token",
			BaseURL:    server.URL,
		}

		run, err := client.FetchWorkflowRun("owner/repo", "CI")
		require.NoError(t, err)
		assert.Nil(t, run)
	})
}
