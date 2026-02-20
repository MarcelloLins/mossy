package footer

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/tui/context"
	"github.com/marcellolins/mossy/internal/tui/keys"
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

	syncStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Background(lipgloss.Color("236")).
			Padding(0, 2)
)

type Model struct {
	ctx  *context.ProgramContext
	help help.Model
}

func New(ctx *context.ProgramContext) Model {
	h := help.New()
	h.ShowAll = true
	h.Styles.FullKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#8FBC8F")).
		Bold(true)
	h.Styles.FullDesc = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))
	h.Styles.FullSeparator = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
	return Model{ctx: ctx, help: h}
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
	var autoRefresh string
	if m.ctx.AutoRefresh {
		base := activeViewStyle
		keyStyle := base.Underline(true)
		countdown := ""
		if !m.ctx.LastRefresh.IsZero() {
			elapsed := int(time.Since(m.ctx.LastRefresh).Seconds())
			remaining := 30 - elapsed
			if remaining < 0 {
				remaining = 0
			}
			countdown = base.UnsetPadding().Render(fmt.Sprintf(" (%ds)", remaining))
		}
		autoRefresh = base.Render("⟳ Auto") + keyStyle.UnsetPadding().Render("R") + base.UnsetPadding().Render("efresh") + countdown
	} else {
		base := inactiveViewStyle
		keyStyle := base.Underline(true)
		countdown := ""
		if m.ctx.PausedRemaining > 0 {
			countdown = base.UnsetPadding().Render(fmt.Sprintf(" (%ds)", m.ctx.PausedRemaining))
		}
		autoRefresh = base.Render("⟳ Auto") + keyStyle.UnsetPadding().Render("R") + base.UnsetPadding().Render("efresh") + countdown
	}
	issues := inactiveViewStyle.Render("⋮ WIP2")
	left := bell + sep + autoRefresh + sep + issues

	// Center: message area
	var mid string
	if m.ctx.Message != "" {
		mid = messageStyle.Render(m.ctx.Message)
	} else if m.ctx.Loading {
		mid = syncStyle.Render("⟳ syncing")
	}

	// Right: donate, help
	donate := rightSectionStyle.Render("♡ donate")
	helpBadge := rightSectionStyle.Render("? help")
	right := donate + sep + helpBadge

	leftWidth := lipgloss.Width(left)
	midWidth := lipgloss.Width(mid)
	rightWidth := lipgloss.Width(right)
	totalGap := m.ctx.Width - leftWidth - midWidth - rightWidth
	if totalGap < 0 {
		totalGap = 0
	}
	leftGap := totalGap / 2
	rightGap := totalGap - leftGap

	bar := left +
		barStyle.Render(strings.Repeat(" ", leftGap)) +
		mid +
		barStyle.Render(strings.Repeat(" ", rightGap)) +
		right

	if m.ctx.ShowHelp {
		m.help.Width = m.ctx.Width
		helpView := m.help.View(keys.Keys)
		return bar + "\n" + helpView
	}

	return bar
}
