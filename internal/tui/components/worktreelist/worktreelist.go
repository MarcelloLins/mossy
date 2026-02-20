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

	// Column layout: branch(grow) + commit(fixed) + location(fixed)
	const (
		commitWidth = 9
		padWidth    = 4 // outer padding from rowStyle (2 each side)
	)
	// Location column gets ~40% of total width, capped at 50
	locationWidth := (width - padWidth) * 2 / 5
	if locationWidth > 50 {
		locationWidth = 50
	}
	if locationWidth < 10 {
		locationWidth = 10
	}
	branchWidth := width - padWidth - commitWidth - locationWidth

	cellStyle := lipgloss.NewStyle()

	colHeader := rowStyle.Render(
		columnHeaderStyle.Width(branchWidth).MaxWidth(branchWidth).Render("\uf418 Branch") +
			columnHeaderStyle.Width(commitWidth).MaxWidth(commitWidth).Render("Commit") +
			columnHeaderStyle.Width(locationWidth).MaxWidth(locationWidth).Render("Location"))
	b.WriteString(colHeader)
	b.WriteString("\n")

	divider := dividerStyle.Render(strings.Repeat("─", width-padWidth))

	for i, wt := range m.worktrees {
		selected := i == m.cursor
		bg := lipgloss.Color("236")

		bStyle := branchStyle
		hStyle := hashStyle
		pStyle := pathStyle
		cStyle := cellStyle
		rStyle := rowStyle
		if selected {
			bStyle = bStyle.Background(bg)
			hStyle = hStyle.Background(bg)
			pStyle = pStyle.Background(bg)
			cStyle = cStyle.Background(bg)
			rStyle = selectedRowStyle
		}

		line := cStyle.Width(branchWidth).MaxWidth(branchWidth).Render(bStyle.Render(wt.Branch)) +
			cStyle.Width(commitWidth).MaxWidth(commitWidth).Render(hStyle.Render(wt.HEAD[:7])) +
			cStyle.Width(locationWidth).MaxWidth(locationWidth).Render(pStyle.Render(wt.Path))

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
