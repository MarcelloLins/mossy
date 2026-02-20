package worktreelist

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/git"
	"github.com/marcellolins/mossy/internal/tui/context"
)

type WorktreesFetchedMsg struct {
	Worktrees []git.Worktree
	Err       error
}

func FetchWorktrees(repoPath string) tea.Cmd {
	return func() tea.Msg {
		wts, err := git.ListWorktrees(repoPath)
		return WorktreesFetchedMsg{Worktrees: wts, Err: err}
	}
}

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8FBC8F")).
			Padding(0, 2)

	branchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	hashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Align(lipgloss.Center)

	columnHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8FBC8F")).
				Bold(true)

	rowStyle = lipgloss.NewStyle().
			Padding(0, 2)
)

type Model struct {
	ctx       *context.ProgramContext
	worktrees []git.Worktree
	err       error
	loaded    bool
}

func New(ctx *context.ProgramContext) Model {
	return Model{ctx: ctx}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case WorktreesFetchedMsg:
		m.worktrees = msg.Worktrees
		m.err = msg.Err
		m.loaded = true
	}
	return m, nil
}

func (m Model) View(width, height int) string {
	if !m.loaded {
		return ""
	}

	if m.err != nil {
		msg := emptyStyle.Render(fmt.Sprintf("Error listing worktrees: %v", m.err))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
	}

	if len(m.worktrees) == 0 {
		msg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8FBC8F")).
			Align(lipgloss.Center).
			Render("No worktrees found\n\n" +
				lipgloss.NewStyle().
					Foreground(lipgloss.Color("245")).
					Render("This repository has no additional worktrees.\n"+
						"Use 'git worktree add' from the command line to create one."))
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, msg)
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(fmt.Sprintf("Worktrees (%d)", len(m.worktrees))))
	b.WriteString("\n\n")

	colHeader := rowStyle.Render(
		columnHeaderStyle.Render("  Branch") + "       " +
			columnHeaderStyle.Render("Commit") + "   " +
			columnHeaderStyle.Render("Location"))
	b.WriteString(colHeader)
	b.WriteString("\n")

	for _, wt := range m.worktrees {
		branch := branchStyle.Render(wt.Branch)
		hash := hashStyle.Render(wt.HEAD[:7])
		path := pathStyle.Render(wt.Path)
		b.WriteString(rowStyle.Render(fmt.Sprintf("  %s  %s  %s", branch, hash, path)))
		b.WriteString("\n")
	}

	s := b.String()
	contentHeight := lipgloss.Height(s)
	if pad := height - contentHeight; pad > 0 {
		s += strings.Repeat("\n", pad)
	}
	return s
}
