package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	footerStyle = lipgloss.NewStyle().Faint(true)
)

// Lines used by the fixed header and footer.
// Header: title line + blank line + column header line = 3
// Footer: blank line + instruction line = 2
const headerLines = 3
const footerLines = 2

type workflowInfo struct {
	Repo     string
	Workflow string
	Status   string
}

type model struct {
	workflow      string
	repos         []string
	rate          int
	client        *GitHubClient
	runs          [][]workflowInfo
	err           error
	fetching      bool
	fetchProgress int
	windowWidth   int
	windowHeight  int
	scrollOffset  int
}

type fetchNextMsg struct {
	index int
}

type fetchedRepoMsg struct {
	index int
	infos []workflowInfo
	err   error
}

type resetProgressMsg struct{}

type tickMsg time.Time

func newModel(workflow string, repos []string, rate int, token string) model {
	return model{
		workflow:  workflow,
		repos:    repos,
		rate:     rate,
		client:   NewGitHubClient(token),
		runs:     placeholderRuns(repos),
		fetching: true,
	}
}

func placeholderRuns(repos []string) [][]workflowInfo {
	runs := make([][]workflowInfo, len(repos))
	for i, repo := range repos {
		runs[i] = []workflowInfo{{Repo: repo, Status: "..."}}
	}
	return runs
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.startFetch(), m.tick())
}

func (m model) tick() tea.Cmd {
	return tea.Tick(time.Duration(m.rate)*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) startFetch() tea.Cmd {
	return func() tea.Msg {
		return fetchNextMsg{index: 0}
	}
}

func (m model) fetchRepo(index int) tea.Cmd {
	repo := m.repos[index]
	workflow := m.workflow
	client := m.client
	return func() tea.Msg {
		if workflow != "" {
			// Single-workflow mode.
			run, err := client.FetchWorkflowRun(repo, workflow)
			if err != nil {
				return fetchedRepoMsg{index: index, err: err}
			}
			var infos []workflowInfo
			if run != nil {
				infos = []workflowInfo{{
					Repo:     repo,
					Workflow: run.Name,
					Status:   formatStatus(run.Status, run.Conclusion),
				}}
			}
			return fetchedRepoMsg{index: index, infos: infos}
		}

		// All-workflows mode.
		runs, err := client.FetchWorkflowRuns(repo)
		if err != nil {
			return fetchedRepoMsg{index: index, err: err}
		}
		var infos []workflowInfo
		for _, run := range runs {
			if strings.HasPrefix(run.Name, "Graph Update") || strings.HasPrefix(run.Name, "go_modules") {
				continue
			}
			infos = append(infos, workflowInfo{
				Repo:     repo,
				Workflow: run.Name,
				Status:   formatStatus(run.Status, run.Conclusion),
			})
		}
		return fetchedRepoMsg{index: index, infos: infos}
	}
}

func formatStatus(status, conclusion string) string {
	if status == "completed" {
		return conclusion
	}
	return status
}

// flatRuns returns a flattened view of all workflow info rows across all repos.
func (m model) flatRuns() []workflowInfo {
	var flat []workflowInfo
	for _, repoRuns := range m.runs {
		flat = append(flat, repoRuns...)
	}
	return flat
}

// totalRows returns the total number of display rows across all repos.
func (m model) totalRows() int {
	total := 0
	for _, repoRuns := range m.runs {
		total += len(repoRuns)
	}
	return total
}

func (m *model) clampScroll() {
	maxRows := m.contentHeight()
	totalRows := m.totalRows()
	maxOffset := totalRows - maxRows
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

func (m model) contentHeight() int {
	h := m.windowHeight - headerLines - footerLines
	if h < 1 {
		h = 1
	}
	return h
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowWidth = msg.Width
		m.windowHeight = msg.Height
		m.clampScroll()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			if !m.fetching {
				m.fetching = true
				m.fetchProgress = 0
				return m, m.startFetch()
			}
		case "up", "k":
			m.scrollOffset--
			m.clampScroll()
		case "down", "j":
			m.scrollOffset++
			m.clampScroll()
		}
	case tickMsg:
		if !m.fetching {
			m.fetching = true
			m.fetchProgress = 0
			return m, tea.Batch(m.startFetch(), m.tick())
		}
		return m, m.tick()
	case fetchNextMsg:
		if msg.index < len(m.repos) {
			return m, m.fetchRepo(msg.index)
		}
	case resetProgressMsg:
		m.fetchProgress = 0
	case fetchedRepoMsg:
		if msg.err != nil {
			m.err = msg.err
			m.fetching = false
		} else {
			if msg.infos != nil {
				m.runs[msg.index] = msg.infos
			}
			m.fetchProgress = msg.index + 1
			if msg.index+1 < len(m.repos) {
				return m, func() tea.Msg { return fetchNextMsg{index: msg.index + 1} }
			}
			m.fetching = false
			m.err = nil
			return m, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				return resetProgressMsg{}
			})
		}
		m.clampScroll()
	}
	return m, nil
}

func renderProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}
	filled := width * current / total
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func (m model) View() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("GHA Monitor"))
	if m.workflow != "" {
		b.WriteString(fmt.Sprintf("  Workflow: %s", m.workflow))
	}
	b.WriteString(fmt.Sprintf("  Refresh: %ds", m.rate))
	b.WriteString(fmt.Sprintf("  %s %d/%d",
		renderProgressBar(m.fetchProgress, len(m.repos), 20),
		m.fetchProgress, len(m.repos)))
	b.WriteString("\n\n")

	// Content
	flat := m.flatRuns()
	if m.err != nil {
		b.WriteString(fmt.Sprintf("Error: %v\n", m.err))
	} else {
		if m.workflow != "" {
			b.WriteString(fmt.Sprintf("%-40s %s\n", "REPOSITORY", "STATUS"))
		} else {
			b.WriteString(fmt.Sprintf("%-40s %-25s %s\n", "REPOSITORY", "WORKFLOW", "STATUS"))
		}

		maxRows := m.contentHeight()
		end := m.scrollOffset + maxRows
		if end > len(flat) {
			end = len(flat)
		}
		visible := flat[m.scrollOffset:end]
		for _, r := range visible {
			if m.workflow != "" {
				b.WriteString(fmt.Sprintf("%-40s %s\n", r.Repo, r.Status))
			} else {
				b.WriteString(fmt.Sprintf("%-40s %-25s %s\n", r.Repo, r.Workflow, r.Status))
			}
		}
	}

	// Pad to push footer to the bottom
	header := headerLines
	contentRendered := 0
	if m.err != nil {
		contentRendered = 1
	} else {
		contentRendered = 1 // column header
		end := m.scrollOffset + m.contentHeight()
		if end > len(flat) {
			end = len(flat)
		}
		contentRendered += end - m.scrollOffset
	}
	usedLines := header + contentRendered + footerLines
	if pad := m.windowHeight - usedLines; pad > 0 {
		b.WriteString(strings.Repeat("\n", pad))
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("q: quit | r: refresh"))

	return b.String()
}

// RunTUI starts the TUI application.
func RunTUI(workflow string, repos []string, rate int, token string) error {
	m := newModel(workflow, repos, rate, token)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
