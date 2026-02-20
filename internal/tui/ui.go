package tui

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/config"
	"github.com/marcellolins/mossy/internal/git"
	"github.com/marcellolins/mossy/internal/tui/components/footer"
	"github.com/marcellolins/mossy/internal/tui/components/repopicker"
	"github.com/marcellolins/mossy/internal/tui/components/tabs"
	"github.com/marcellolins/mossy/internal/tui/components/worktreelist"
	"github.com/marcellolins/mossy/internal/tui/context"
)

type configLoadedMsg struct {
	repos []config.Repository
	err   error
}

type tickMsg time.Time
type uiTickMsg time.Time

type repoWorktreeResult struct {
	path      string
	worktrees []git.Worktree
	err       error
}

type allWorktreesFetchedMsg struct {
	results []repoWorktreeResult
}

const refreshInterval = 30 * time.Second

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func uiTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return uiTickMsg(t)
	})
}

type viewState int

const (
	viewNormal viewState = iota
	viewRepoPicker
	viewConfirmDelete
)

type Model struct {
	ctx          *context.ProgramContext
	tabs         tabs.Model
	footer       footer.Model
	repoPicker   repopicker.Model
	worktreeList worktreelist.Model
	view         viewState
}

func New() Model {
	ctx := &context.ProgramContext{
		ActiveRepo:  -1,
		AutoRefresh: true,
	}
	return Model{
		ctx:          ctx,
		tabs:         tabs.New(ctx),
		footer:       footer.New(ctx),
		worktreeList: worktreelist.New(ctx),
		view:         viewNormal,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("mossy"),
		func() tea.Msg {
			cfg, err := config.Load()
			return configLoadedMsg{repos: cfg.Repos, err: err}
		},
	)
}

func (m Model) fetchActiveWorktrees() tea.Cmd {
	if m.ctx.ActiveRepo < 0 || m.ctx.ActiveRepo >= len(m.ctx.Repos) {
		return nil
	}
	return worktreelist.FetchWorktrees(m.ctx.Repos[m.ctx.ActiveRepo].Path)
}

func (m Model) saveRepos() tea.Cmd {
	repos := make([]config.Repository, len(m.ctx.Repos))
	for i, r := range m.ctx.Repos {
		repos[i] = config.Repository{Name: r.Name, Path: r.Path}
	}
	return func() tea.Msg {
		_ = config.Save(config.Config{Repos: repos})
		return nil
	}
}

func (m Model) fetchAllWorktrees() tea.Cmd {
	type repoRef struct {
		path string
	}
	refs := make([]repoRef, len(m.ctx.Repos))
	for i, r := range m.ctx.Repos {
		refs[i] = repoRef{path: r.Path}
	}
	if len(refs) == 0 {
		return nil
	}
	return func() tea.Msg {
		var results []repoWorktreeResult
		for _, r := range refs {
			wts, err := git.ListWorktrees(r.path)
			results = append(results, repoWorktreeResult{
				path:      r.path,
				worktrees: wts,
				err:       err,
			})
		}
		return allWorktreesFetchedMsg{results: results}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case configLoadedMsg:
		if msg.err == nil {
			for _, r := range msg.repos {
				m.ctx.Repos = append(m.ctx.Repos, context.Repository{
					Name: r.Name,
					Path: r.Path,
				})
			}
			if len(m.ctx.Repos) > 0 {
				m.ctx.ActiveRepo = 0
			}
		}
		return m, tea.Batch(m.fetchActiveWorktrees(), tickCmd(), uiTickCmd())
	case uiTickMsg:
		if m.ctx.AutoRefresh {
			return m, uiTickCmd()
		}
		return m, nil
	case tickMsg:
		if !m.ctx.AutoRefresh {
			return m, nil
		}
		return m, tea.Batch(m.fetchAllWorktrees(), tickCmd())
	case allWorktreesFetchedMsg:
		m.ctx.Loading = false
		m.ctx.LastRefresh = time.Now()
		for _, res := range msg.results {
			for i := range m.ctx.Repos {
				if m.ctx.Repos[i].Path != res.path {
					continue
				}
				if res.err == nil {
					m.ctx.Repos[i].WorktreeCount = len(res.worktrees)
				}
				if i == m.ctx.ActiveRepo {
					m.worktreeList, _ = m.worktreeList.Update(
						worktreelist.WorktreesFetchedMsg{Worktrees: res.worktrees, Err: res.err},
					)
				}
				break
			}
		}
		return m, nil
	case worktreelist.WorktreesFetchedMsg:
		m.ctx.Loading = false
		m.ctx.LastRefresh = time.Now()
		m.worktreeList, _ = m.worktreeList.Update(msg)
		if m.ctx.ActiveRepo >= 0 && m.ctx.ActiveRepo < len(m.ctx.Repos) {
			m.ctx.Repos[m.ctx.ActiveRepo].WorktreeCount = len(msg.Worktrees)
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		if m.view == viewRepoPicker {
			m.repoPicker.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if m.view == viewConfirmDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y":
				i := m.ctx.ActiveRepo
				m.ctx.Repos = append(m.ctx.Repos[:i], m.ctx.Repos[i+1:]...)
				if len(m.ctx.Repos) == 0 {
					m.ctx.ActiveRepo = -1
				} else if i >= len(m.ctx.Repos) {
					m.ctx.ActiveRepo = len(m.ctx.Repos) - 1
				}
				m.tabs.ScrollToActive()
				m.ctx.Message = ""
				m.view = viewNormal
				return m, tea.Batch(m.saveRepos(), m.fetchActiveWorktrees())
			}
			m.ctx.Message = ""
			m.view = viewNormal
		}
		return m, nil
	}

	if m.view == viewRepoPicker {
		switch msg := msg.(type) {
		case repopicker.RepoSelectedMsg:
			m.ctx.Repos = append(m.ctx.Repos, context.Repository{
				Name: msg.Name,
				Path: msg.Path,
			})
			m.ctx.ActiveRepo = len(m.ctx.Repos) - 1
			m.tabs.ScrollToActive()
			m.view = viewNormal
			return m, tea.Batch(m.saveRepos(), m.fetchActiveWorktrees())
		case repopicker.RepoPickerCancelledMsg:
			m.view = viewNormal
			return m, nil
		default:
			var cmd tea.Cmd
			m.repoPicker, cmd = m.repoPicker.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "a":
			home, err := os.UserHomeDir()
			if err != nil {
				home = "/"
			}
			var paths []string
			for _, r := range m.ctx.Repos {
				paths = append(paths, r.Path)
			}
			m.repoPicker = repopicker.New(home, m.ctx.Width, m.ctx.Height, paths)
			m.view = viewRepoPicker
			return m, nil
		case "r":
			if len(m.ctx.Repos) > 0 {
				m.ctx.Loading = true
				return m, m.fetchAllWorktrees()
			}
		case "R":
			m.ctx.AutoRefresh = !m.ctx.AutoRefresh
			if m.ctx.AutoRefresh {
				elapsed := time.Duration(int(refreshInterval.Seconds())-m.ctx.PausedRemaining) * time.Second
				m.ctx.LastRefresh = time.Now().Add(-elapsed)
				return m, tea.Batch(tickCmd(), uiTickCmd())
			}
			if !m.ctx.LastRefresh.IsZero() {
				elapsed := int(time.Since(m.ctx.LastRefresh).Seconds())
				remaining := int(refreshInterval.Seconds()) - elapsed
				if remaining < 0 {
					remaining = 0
				}
				m.ctx.PausedRemaining = remaining
			}
			return m, nil
		case "d":
			if len(m.ctx.Repos) > 0 {
				name := m.ctx.Repos[m.ctx.ActiveRepo].Name
				m.ctx.Message = fmt.Sprintf("Remove %q? (y/n)", name)
				m.view = viewConfirmDelete
				return m, nil
			}
		case "h", "left":
			if len(m.ctx.Repos) > 0 && m.ctx.ActiveRepo > 0 {
				m.ctx.ActiveRepo--
				m.tabs.ScrollToActive()
				return m, m.fetchActiveWorktrees()
			}
		case "l", "right":
			if len(m.ctx.Repos) > 0 && m.ctx.ActiveRepo < len(m.ctx.Repos)-1 {
				m.ctx.ActiveRepo++
				m.tabs.ScrollToActive()
				return m, m.fetchActiveWorktrees()
			}
		case "j", "down", "k", "up":
			var cmd tea.Cmd
			m.worktreeList, cmd = m.worktreeList.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.view == viewRepoPicker {
		return m.repoPicker.View()
	}

	top := m.tabs.View()
	foot := m.footer.View()

	topHeight := lipgloss.Height(top)
	footHeight := lipgloss.Height(foot)
	mid := m.ctx.Height - topHeight - footHeight
	if mid < 0 {
		mid = 0
	}

	var content string
	if len(m.ctx.Repos) == 0 {
		welcome := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8FBC8F")).
			Align(lipgloss.Center).
			Render("🌿\n\nWelcome to mossy\n\n" +
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("245")).
					Render("Press 'a' to add your first repository"))
		content = lipgloss.Place(m.ctx.Width, mid, lipgloss.Center, lipgloss.Center, welcome)
	} else {
		content = m.worktreeList.View(m.ctx.Width, mid)
	}

	return top + "\n" + content + "\n" + foot
}
