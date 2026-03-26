package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ghclient "ghamon/internal/github"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	headerInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	colHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("33")).
			Underline(true)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusStyles = map[string]lipgloss.Style{
		"success":      lipgloss.NewStyle().Foreground(lipgloss.Color("40")),
		"failure":      lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		"in progress":  lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		"queued":       lipgloss.NewStyle().Foreground(lipgloss.Color("220")),
		"cancelled":    lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		"timed_out":    lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		"skipped":      lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		"no runs":      lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		"no workflows": lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
		"error":        lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
	}

	defaultStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
)

const (
	headerHeight = 4
	footerHeight = 1
)

// ── Messages ──────────────────────────────────────────────────────────────────

type tickMsg time.Time

type fetchCompleteMsg struct {
	runs []ghclient.WorkflowRun
	err  error
}

// ── Model ─────────────────────────────────────────────────────────────────────

// Model is the top-level Bubble Tea model for ghamon.
type Model struct {
	repos    []string
	Workflow string
	Rate     int

	client ghclient.Client

	runs     []ghclient.WorkflowRun
	loading  bool
	fetchErr error

	prog progress.Model
	vp   viewport.Model

	width  int
	height int
	ready  bool
}

// New creates a new Model.
func New(repos []string, workflow string, rate int, client ghclient.Client) Model {
	return Model{
		repos:    repos,
		Workflow: workflow,
		Rate:     rate,
		client:   client,
		loading:  true,
		prog:     progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
	}
}

// ── Init ──────────────────────────────────────────────────────────────────────

// Init starts the initial data fetch and the auto-refresh ticker.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.doFetch(),
		m.tick(),
	)
}

func (m Model) tick() tea.Cmd {
	d := time.Duration(m.Rate) * time.Second
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) doFetch() tea.Cmd {
	repos := m.repos
	workflow := m.Workflow
	client := m.client

	return func() tea.Msg {
		if len(repos) == 0 {
			return fetchCompleteMsg{}
		}
		ctx := context.Background()
		var all []ghclient.WorkflowRun
		for _, full := range repos {
			parts := strings.SplitN(full, "/", 2)
			runs, err := client.GetWorkflowStatuses(ctx, parts[0], parts[1], workflow)
			if err != nil {
				all = append(all, ghclient.WorkflowRun{Repo: full, Status: "error"})
				continue
			}
			all = append(all, runs...)
		}
		return fetchCompleteMsg{runs: all}
	}
}

// ── Update ────────────────────────────────────────────────────────────────────

// Update handles incoming messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.prog.Width = m.width - 4
		vpH := m.height - headerHeight - footerHeight
		if vpH < 1 {
			vpH = 1
		}
		if !m.ready {
			m.vp = viewport.New(m.width, vpH)
			m.ready = true
		} else {
			m.vp.Width = m.width
			m.vp.Height = vpH
		}
		m.vp.SetContent(m.content())

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "ctrl+c":
			return m, tea.Quit
		case "r", "R":
			m.loading = true
			m.resetProgress()
			cmds = append(cmds, m.doFetch())
		}

	case tickMsg:
		m.loading = true
		m.resetProgress()
		cmds = append(cmds, m.doFetch(), m.tick())

	case fetchCompleteMsg:
		m.loading = false
		m.fetchErr = msg.err
		if msg.err == nil {
			m.runs = msg.runs
		}
		cmds = append(cmds, m.prog.SetPercent(1.0))
		if m.ready {
			m.vp.SetContent(m.content())
		}

	case progress.FrameMsg:
		pm, cmd := m.prog.Update(msg)
		m.prog = pm.(progress.Model)
		cmds = append(cmds, cmd)
	}

	if m.ready {
		vp, cmd := m.vp.Update(msg)
		m.vp = vp
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// ── View ──────────────────────────────────────────────────────────────────────

// View renders the full TUI.
func (m Model) View() string {
	if !m.ready {
		return "\nInitializing…"
	}
	return m.header() + "\n" + m.vp.View() + "\n" + m.footer()
}

func (m Model) header() string {
	wf := m.Workflow
	if wf == "" {
		wf = "all"
	}
	title := titleStyle.Render("GHA Monitor (ghamon)")
	info := headerInfoStyle.Render(fmt.Sprintf("Workflow: %-20s  Rate: %ds", wf, m.Rate))
	return strings.Join([]string{title, info, m.prog.View()}, "\n")
}

func (m Model) content() string {
	if len(m.repos) == 0 {
		return "  No repositories configured. Specify repos via -c or as arguments.\n"
	}
	if m.fetchErr != nil {
		return fmt.Sprintf("  Error: %v\n", m.fetchErr)
	}
	if len(m.runs) == 0 {
		return "  Fetching data…\n"
	}

	repoW, wfW, statusW := 40, 30, 15
	for _, r := range m.runs {
		if len(r.Repo) > repoW {
			repoW = len(r.Repo) + 2
		}
		if len(r.Workflow) > wfW {
			wfW = len(r.Workflow) + 2
		}
	}

	var sb strings.Builder
	hdr := fmt.Sprintf("  %-*s  %-*s  %-*s", repoW, "REPOSITORY", wfW, "WORKFLOW", statusW, "STATUS")
	sb.WriteString(colHeaderStyle.Render(hdr))
	sb.WriteByte('\n')

	for _, r := range m.runs {
		ds := r.DisplayStatus()
		style, ok := statusStyles[ds]
		if !ok {
			style = defaultStatusStyle
		}
		wf := r.Workflow
		if wf == "" {
			wf = "-"
		}
		sb.WriteString(fmt.Sprintf("  %-*s  %-*s  ", repoW, r.Repo, wfW, wf))
		sb.WriteString(style.Render(ds))
		sb.WriteByte('\n')
	}
	return sb.String()
}

// resetProgress replaces the progress model with a fresh one at 0%.
// This avoids in-flight backward animation frames from a SetPercent(0) cmd
// racing with the forward animation queued when the fetch completes.
func (m *Model) resetProgress() {
	m.prog = progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage())
	m.prog.Width = m.width - 4
}

func (m Model) footer() string {
	hints := footerStyle.Render("  q: quit   r: refresh")
	pad := m.width - lipgloss.Width(hints)
	if pad < 0 {
		pad = 0
	}
	return hints + strings.Repeat(" ", pad)
}
