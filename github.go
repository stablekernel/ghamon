package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://api.github.com"

// WorkflowRun represents a GitHub Actions workflow run.
type WorkflowRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
}

type workflowRunsResponse struct {
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

// GitHubClient fetches workflow run data from the GitHub API.
type GitHubClient struct {
	HTTPClient *http.Client
	Token      string
	BaseURL    string
}

// NewGitHubClient creates a new GitHubClient with default settings.
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		HTTPClient: http.DefaultClient,
		Token:      token,
		BaseURL:    defaultBaseURL,
	}
}

// FetchWorkflowRun fetches the most recent run of the named workflow for a repository.
func (c *GitHubClient) FetchWorkflowRun(repo, workflow string) (*WorkflowRun, error) {
	url := fmt.Sprintf("%s/repos/%s/actions/runs?per_page=50", c.BaseURL, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching runs for %s: %w", repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, repo)
	}

	var result workflowRunsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response for %s: %w", repo, err)
	}

	// Return the most recent run matching the workflow name.
	for _, run := range result.WorkflowRuns {
		if run.Name == workflow {
			return &run, nil
		}
	}

	return nil, nil
}
