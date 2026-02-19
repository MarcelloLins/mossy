package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/tui/context"
)

var (
	barStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236"))

	activeViewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E")).
			Background(lipgloss.Color("236")).
			Bold(true).
			Padding(0, 1)

	inactiveViewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	bellStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	rightSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	sepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("236"))

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E")).
			Background(lipgloss.Color("236")).
			Bold(true).
			Padding(0, 2)
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
	sep := sepStyle.Render(" │ ")

	// Left: view switcher
	bell := bellStyle.Render("󰂚")
	prs := activeViewStyle.Render("⎔ WIP1")
	issues := inactiveViewStyle.Render("⋮ WIP2")
	left := bell + sep + prs + sep + issues

	// Center: message area
	var mid string
	if m.ctx.Message != "" {
		mid = messageStyle.Render(m.ctx.Message)
	}

	// Right: donate, help
	donate := rightSectionStyle.Render("♡ donate")
	help := rightSectionStyle.Render("? help")
	right := donate + sep + help

	leftWidth := lipgloss.Width(left)
	midWidth := lipgloss.Width(mid)
	rightWidth := lipgloss.Width(right)
	totalGap := m.ctx.Width - leftWidth - midWidth - rightWidth
	if totalGap < 0 {
		totalGap = 0
	}
	leftGap := totalGap / 2
	rightGap := totalGap - leftGap

	return left +
		barStyle.Render(strings.Repeat(" ", leftGap)) +
		mid +
		barStyle.Render(strings.Repeat(" ", rightGap)) +
		right
}
