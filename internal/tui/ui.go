package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/config"
	"github.com/marcellolins/mossy/internal/git"
	"github.com/marcellolins/mossy/internal/tui/components/footer"
	"github.com/marcellolins/mossy/internal/tui/components/repopicker"
	"github.com/marcellolins/mossy/internal/tui/components/sidepanel"
	"github.com/marcellolins/mossy/internal/tui/components/tabs"
	"github.com/marcellolins/mossy/internal/tui/components/worktreecreate"
	"github.com/marcellolins/mossy/internal/tui/components/worktreelist"
	"github.com/marcellolins/mossy/internal/tui/components/worktreeremove"
	"github.com/marcellolins/mossy/internal/tui/context"
)

type configLoadedMsg struct {
	repos []config.Repository
	err   error
}

type tickMsg time.Time
type uiTickMsg time.Time

type worktreeCreatedMsg struct {
	path string
	err  error
}

type worktreeRemovedMsg struct {
	name string
	err  error
}

type repoWorktreeResult struct {
	path      string
	worktrees []git.Worktree
	err       error
}

type commitsFetchedMsg struct {
	commits []git.Commit
	err     error
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
	viewCreateWorktree
	viewRemoveWorktree
)

type Model struct {
	ctx            *context.ProgramContext
	tabs           tabs.Model
	footer         footer.Model
	repoPicker     repopicker.Model
	worktreeCreate worktreecreate.Model
	worktreeRemove worktreeremove.Model
	worktreeList   worktreelist.Model
	sidePanel      sidepanel.Model
	view           viewState
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
		sidePanel:    sidepanel.New(),
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

func (m Model) fetchCommits() tea.Cmd {
	wt, ok := m.worktreeList.SelectedWorktree()
	if !ok || m.ctx.ActiveRepo < 0 || m.ctx.ActiveRepo >= len(m.ctx.Repos) {
		return nil
	}
	repoPath := m.ctx.Repos[m.ctx.ActiveRepo].Path
	branch := wt.Branch
	if branch == "" || branch == "(detached)" {
		return nil
	}
	return func() tea.Msg {
		commits, err := git.ListCommits(repoPath, branch)
		return commitsFetchedMsg{commits: commits, err: err}
	}
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
		m.ctx.LastRefresh = time.Now()
		return m, tea.Batch(m.fetchActiveWorktrees(), tickCmd(), uiTickCmd())
	case uiTickMsg:
		if m.ctx.Message != "" && !m.ctx.MessageExpiry.IsZero() && time.Now().After(m.ctx.MessageExpiry) {
			m.ctx.Message = ""
			m.ctx.MessageExpiry = time.Time{}
		}
		if m.ctx.AutoRefresh || (m.ctx.Message != "" && !m.ctx.MessageExpiry.IsZero()) {
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
		m.worktreeList, _ = m.worktreeList.Update(msg)
		if m.ctx.ActiveRepo >= 0 && m.ctx.ActiveRepo < len(m.ctx.Repos) {
			m.ctx.Repos[m.ctx.ActiveRepo].WorktreeCount = len(msg.Worktrees)
		}
		return m, m.fetchCommits()
	case commitsFetchedMsg:
		if msg.err == nil {
			m.sidePanel.SetCommits(msg.commits)
		}
		return m, nil
	case worktreeCreatedMsg:
		m.ctx.Loading = false
		if msg.err != nil {
			m.ctx.Message = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.ctx.Message = fmt.Sprintf("Worktree created at %s", msg.path)
		}
		m.ctx.MessageExpiry = time.Now().Add(5 * time.Second)
		m.view = viewNormal
		return m, tea.Batch(m.fetchActiveWorktrees(), uiTickCmd())
	case worktreeRemovedMsg:
		m.ctx.Loading = false
		if msg.err != nil {
			m.ctx.Message = fmt.Sprintf("Error: %v", msg.err)
		} else {
			m.ctx.Message = fmt.Sprintf("Worktree %q removed", msg.name)
		}
		m.ctx.MessageExpiry = time.Now().Add(5 * time.Second)
		m.view = viewNormal
		return m, tea.Batch(m.fetchActiveWorktrees(), uiTickCmd())
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		if m.view == viewRepoPicker {
			m.repoPicker.SetSize(msg.Width, msg.Height)
		}
		if m.view == viewCreateWorktree {
			m.worktreeCreate.SetSize(msg.Width, msg.Height)
		}
		if m.view == viewRemoveWorktree {
			m.worktreeRemove.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "?" {
			m.ctx.ShowHelp = !m.ctx.ShowHelp
			return m, nil
		}
		if m.ctx.ShowHelp {
			if msg.String() == "esc" {
				m.ctx.ShowHelp = false
			}
			return m, nil
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

	if m.view == viewCreateWorktree {
		switch msg := msg.(type) {
		case worktreecreate.WorktreeCreateRequestMsg:
			repoPath := m.ctx.Repos[m.ctx.ActiveRepo].Path
			name := msg.Name
			branch := msg.Branch
			m.worktreeCreate.Creating = true
			return m, func() tea.Msg {
				wtPath := filepath.Join(filepath.Dir(repoPath), name)
				err := git.AddWorktree(repoPath, name, branch)
				return worktreeCreatedMsg{path: wtPath, err: err}
			}
		case worktreecreate.WorktreeCreateCancelledMsg:
			m.view = viewNormal
			return m, nil
		default:
			var cmd tea.Cmd
			m.worktreeCreate, cmd = m.worktreeCreate.Update(msg)
			return m, cmd
		}
	}

	if m.view == viewRemoveWorktree {
		switch msg := msg.(type) {
		case worktreeremove.WorktreeRemoveRequestMsg:
			repoPath := m.ctx.Repos[m.ctx.ActiveRepo].Path
			wtPath := msg.WtPath
			branch := msg.Branch
			deleteBranch := msg.DeleteBranch
			wtName := filepath.Base(wtPath)
			m.worktreeRemove.Removing = true
			return m, func() tea.Msg {
				err := git.RemoveWorktree(repoPath, wtPath, branch, deleteBranch)
				return worktreeRemovedMsg{name: wtName, err: err}
			}
		case worktreeremove.WorktreeRemoveCancelledMsg:
			m.view = viewNormal
			return m, nil
		default:
			var cmd tea.Cmd
			m.worktreeRemove, cmd = m.worktreeRemove.Update(msg)
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
		case "n":
			if len(m.ctx.Repos) > 0 {
				m.worktreeCreate = worktreecreate.New(m.ctx.Width, m.ctx.Height)
				m.view = viewCreateWorktree
				return m, textinput.Blink
			}
		case "x":
			if wt, ok := m.worktreeList.SelectedWorktree(); ok {
				m.worktreeRemove = worktreeremove.New(filepath.Base(wt.Path), wt.Path, wt.Branch, m.ctx.Width, m.ctx.Height)
				m.view = viewRemoveWorktree
				return m, nil
			}
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
			return m, tea.Batch(cmd, m.fetchCommits())
		case "[":
			m.sidePanel.PrevCommit()
			return m, nil
		case "]":
			m.sidePanel.NextCommit()
			return m, nil
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.view == viewRepoPicker {
		return m.repoPicker.View()
	}

	if m.view == viewCreateWorktree {
		return m.worktreeCreate.View()
	}

	if m.view == viewRemoveWorktree {
		return m.worktreeRemove.View()
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
	} else if m.worktreeList.HasWorktrees() {
		panelWidth := m.ctx.Width / 3
		listWidth := m.ctx.Width - panelWidth
		if wt, ok := m.worktreeList.SelectedWorktree(); ok {
			m.sidePanel.SetWorktree(&wt)
		}
		list := m.worktreeList.View(listWidth, mid)
		panel := m.sidePanel.View(panelWidth, mid)
		content = lipgloss.JoinHorizontal(lipgloss.Top, list, panel)
	} else {
		content = m.worktreeList.View(m.ctx.Width, mid)
	}

	return top + "\n" + content + "\n" + foot
}
