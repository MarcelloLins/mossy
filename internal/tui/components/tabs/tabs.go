package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/tui/context"
)

var (
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8FBC8F")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)

	addTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(0, 2)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(0, 1)

	borderColor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

type Model struct {
	ctx *context.ProgramContext
}

func New(ctx *context.ProgramContext) Model {
	return Model{ctx: ctx}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	logo := logoStyle.Render("🌿 mossy")

	sep := separatorStyle.Render("│")

	var parts []string
	parts = append(parts, logo, sep)

	if len(m.ctx.Repos) == 0 {
		parts = append(parts, emptyStyle.Render("No repositories added. Press 'a' to add your first repository."))
	} else {
		for i, repo := range m.ctx.Repos {
			if i == m.ctx.ActiveRepo {
				parts = append(parts, activeTabStyle.Render(repo.Name))
			} else {
				parts = append(parts, inactiveTabStyle.Render(repo.Name))
			}
			parts = append(parts, sep)
		}
		parts = append(parts, addTabStyle.Render("+ Add"))
	}

	bar := lipgloss.JoinHorizontal(lipgloss.Center, parts...)
	border := borderColor.Render(strings.Repeat("─", m.ctx.Width))

	return bar + "\n" + border
}
