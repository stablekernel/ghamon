package github

import (
	"context"
	"fmt"
	"strings"
	"time"

	gogithub "github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

// WorkflowRun represents the status of a single workflow run.
type WorkflowRun struct {
	Repo       string
	Workflow   string
	Status     string
	Conclusion string
	UpdatedAt  time.Time
	URL        string
}

// DisplayStatus returns a human-readable combined status string.
func (w WorkflowRun) DisplayStatus() string {
	switch w.Status {
	case "completed":
		if w.Conclusion != "" {
			return w.Conclusion
		}
		return "completed"
	case "in_progress":
		return "in progress"
	case "queued":
		return "queued"
	default:
		if w.Status != "" {
			return w.Status
		}
		return "unknown"
	}
}

// Client is the interface for fetching workflow data from GitHub.
type Client interface {
	GetWorkflowStatuses(ctx context.Context, owner, repo, workflowFile string) ([]WorkflowRun, error)
}

type ghClient struct {
	gh *gogithub.Client
}

// New creates a new GitHub API client authenticated with the provided token.
func New(token string) Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &ghClient{gh: gogithub.NewClient(tc)}
}

// GetWorkflowStatuses fetches the latest run for the specified workflow (or all
// workflows when workflowFile is "").
func (c *ghClient) GetWorkflowStatuses(ctx context.Context, owner, repo, workflowFile string) ([]WorkflowRun, error) {
	if workflowFile != "" {
		return c.getByFile(ctx, owner, repo, workflowFile)
	}
	return c.getAll(ctx, owner, repo)
}

func (c *ghClient) getByFile(ctx context.Context, owner, repo, file string) ([]WorkflowRun, error) {
	opts := &gogithub.ListWorkflowRunsOptions{
		ListOptions: gogithub.ListOptions{PerPage: 1},
	}
	runs, _, err := c.gh.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, file, opts)
	if err != nil {
		return nil, fmt.Errorf("listing workflow runs for %s/%s (%s): %w", owner, repo, file, err)
	}
	if runs == nil || len(runs.WorkflowRuns) == 0 {
		return []WorkflowRun{{
			Repo:     owner + "/" + repo,
			Workflow: file,
			Status:   "no runs",
		}}, nil
	}
	return []WorkflowRun{runFromAPI(owner, repo, file, runs.WorkflowRuns[0])}, nil
}

func (c *ghClient) getAll(ctx context.Context, owner, repo string) ([]WorkflowRun, error) {
	wfs, _, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, &gogithub.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("listing workflows for %s/%s: %w", owner, repo, err)
	}

	seen := make(map[string]bool)
	var results []WorkflowRun

	for _, wf := range wfs.Workflows {
		if wf.Path == nil {
			continue
		}
		parts := strings.Split(*wf.Path, "/")
		fileName := parts[len(parts)-1]
		if seen[fileName] {
			continue
		}
		seen[fileName] = true

		wfName := fileName
		if wf.Name != nil && *wf.Name != "" {
			wfName = *wf.Name
		}

		opts := &gogithub.ListWorkflowRunsOptions{
			ListOptions: gogithub.ListOptions{PerPage: 1},
		}
		runs, _, err := c.gh.Actions.ListWorkflowRunsByFileName(ctx, owner, repo, fileName, opts)
		if err != nil {
			results = append(results, WorkflowRun{
				Repo:     owner + "/" + repo,
				Workflow: wfName,
				Status:   "error",
			})
			continue
		}
		if runs == nil || len(runs.WorkflowRuns) == 0 {
			results = append(results, WorkflowRun{
				Repo:     owner + "/" + repo,
				Workflow: wfName,
				Status:   "no runs",
			})
			continue
		}
		results = append(results, runFromAPI(owner, repo, wfName, runs.WorkflowRuns[0]))
	}

	if len(results) == 0 {
		results = append(results, WorkflowRun{
			Repo:   owner + "/" + repo,
			Status: "no workflows",
		})
	}
	return results, nil
}

func runFromAPI(owner, repo, workflow string, r *gogithub.WorkflowRun) WorkflowRun {
	wr := WorkflowRun{
		Repo:     owner + "/" + repo,
		Workflow: workflow,
	}
	if r.Status != nil {
		wr.Status = *r.Status
	}
	if r.Conclusion != nil {
		wr.Conclusion = *r.Conclusion
	}
	if r.UpdatedAt != nil {
		wr.UpdatedAt = r.UpdatedAt.Time
	}
	if r.HTMLURL != nil {
		wr.URL = *r.HTMLURL
	}
	return wr
}
