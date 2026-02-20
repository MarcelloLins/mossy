package worktreelist

import (
	"fmt"
	"path/filepath"
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
	nameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	branchStyle = lipgloss.NewStyle().
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
			Padding(0, 2).
			PaddingTop(1)

	dividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))
)

var (
	selectedRowStyle = lipgloss.NewStyle().
		Padding(0, 2).
		PaddingTop(1).
		Background(lipgloss.Color("236"))
)

type Model struct {
	ctx       *context.ProgramContext
	worktrees []git.Worktree
	cursor    int
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
		m.cursor = 0
		m.loaded = true
	case tea.KeyMsg:
		if len(m.worktrees) == 0 {
			break
		}
		switch msg.String() {
		case "j", "down":
			if m.cursor < len(m.worktrees)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		}
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

	// Column layout: name(grow) + commit(fixed) + branch(fixed)
	const (
		commitWidth = 9
		padWidth    = 4 // outer padding from rowStyle (2 each side)
	)
	// Branch column gets ~40% of total width, capped at 50
	branchColWidth := (width - padWidth) * 2 / 5
	if branchColWidth > 50 {
		branchColWidth = 50
	}
	if branchColWidth < 10 {
		branchColWidth = 10
	}
	nameWidth := width - padWidth - commitWidth - branchColWidth

	cellStyle := lipgloss.NewStyle()

	colHeader := rowStyle.Render(
		columnHeaderStyle.Width(nameWidth).MaxWidth(nameWidth).Render("\uf413 Worktree") +
			columnHeaderStyle.Width(commitWidth).MaxWidth(commitWidth).Render("Commit") +
			columnHeaderStyle.Width(branchColWidth).MaxWidth(branchColWidth).Render("\uf418 Branch"))
	b.WriteString(colHeader)
	b.WriteString("\n")

	divider := dividerStyle.Render(strings.Repeat("─", width-padWidth))

	for i, wt := range m.worktrees {
		selected := i == m.cursor
		bg := lipgloss.Color("236")

		nStyle := nameStyle
		hStyle := hashStyle
		bStyle := branchStyle
		cStyle := cellStyle
		rStyle := rowStyle
		if selected {
			nStyle = nStyle.Background(bg)
			hStyle = hStyle.Background(bg)
			bStyle = bStyle.Background(bg)
			cStyle = cStyle.Background(bg)
			rStyle = selectedRowStyle
		}

		wtName := filepath.Base(wt.Path)
		line := cStyle.Width(nameWidth).MaxWidth(nameWidth).Render(nStyle.Render(wtName)) +
			cStyle.Width(commitWidth).MaxWidth(commitWidth).Render(hStyle.Render(wt.HEAD[:7])) +
			cStyle.Width(branchColWidth).MaxWidth(branchColWidth).Render(bStyle.Render(wt.Branch))

		b.WriteString(rStyle.Render(line))
		b.WriteString("\n")
		if i < len(m.worktrees)-1 {
			b.WriteString(rowStyle.Render(divider))
			b.WriteString("\n")
		}
	}

	s := b.String()
	contentHeight := lipgloss.Height(s)
	if pad := height - contentHeight; pad > 0 {
		s += strings.Repeat("\n", pad)
	}
	return s
}
