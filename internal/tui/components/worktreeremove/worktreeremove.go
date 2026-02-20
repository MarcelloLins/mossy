package worktreeremove

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorktreeRemoveRequestMsg struct {
	WtPath       string
	Branch       string
	DeleteBranch bool
}

type WorktreeRemoveCancelledMsg struct{}

const modalWidth = 50

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8FBC8F")).
			Padding(0, 1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	activeButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFBD2E")).
				Bold(true).
				Padding(0, 2)

	inactiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2).
			Width(modalWidth)

	checkboxActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFBD2E")).
				Bold(true).
				Padding(0, 1)

	checkboxInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 1)
)

type Model struct {
	wtName       string
	wtPath       string
	branch       string
	deleteBranch bool
	focus        int // 0=toggle, 1=Remove, 2=Cancel
	width        int
	height       int
	Removing     bool
}

func New(wtName, wtPath, branch string, width, height int) Model {
	return Model{
		wtName: wtName,
		wtPath: wtPath,
		branch: branch,
		focus:  0,
		width:  width,
		height: height,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.Removing {
		return m, nil
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return WorktreeRemoveCancelledMsg{} }
		case "tab":
			m.focus = (m.focus + 1) % 3
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + 2) % 3
			return m, nil
		case "up":
			if m.focus > 0 {
				if m.focus == 2 {
					m.focus = 1
				} else {
					m.focus--
				}
			}
			return m, nil
		case "down":
			if m.focus < 1 {
				m.focus = 1
			}
			return m, nil
		case "left":
			if m.focus == 2 {
				m.focus = 1
				return m, nil
			}
		case "right":
			if m.focus == 1 {
				m.focus = 2
				return m, nil
			}
		case " ":
			if m.focus == 0 {
				m.deleteBranch = !m.deleteBranch
				return m, nil
			}
		case "enter":
			switch m.focus {
			case 0:
				m.deleteBranch = !m.deleteBranch
				return m, nil
			case 1:
				wtPath := m.wtPath
				branch := m.branch
				deleteBranch := m.deleteBranch
				return m, func() tea.Msg {
					return WorktreeRemoveRequestMsg{
						WtPath:       wtPath,
						Branch:       branch,
						DeleteBranch: deleteBranch,
					}
				}
			case 2:
				return m, func() tea.Msg { return WorktreeRemoveCancelledMsg{} }
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Remove Worktree"))
	b.WriteString("\n\n")

	if m.Removing {
		removingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E")).
			Bold(true).
			Padding(0, 1)
		b.WriteString(removingStyle.Render("⟳ Removing worktree…"))

		modal := modalStyle.Render(b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	}

	b.WriteString(labelStyle.Render("Worktree"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render(m.wtName))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Branch"))
	b.WriteString("\n")
	b.WriteString(valueStyle.Render(m.branch))
	b.WriteString("\n\n")

	check := "[ ]"
	if m.deleteBranch {
		check = "[✓]"
	}
	toggleLabel := check + " Delete branch"
	if m.focus == 0 {
		b.WriteString(checkboxActiveStyle.Render(toggleLabel))
	} else {
		b.WriteString(checkboxInactiveStyle.Render(toggleLabel))
	}
	b.WriteString("\n\n")

	removeBtn := "[Remove]"
	cancelBtn := "[Cancel]"
	if m.focus == 1 {
		removeBtn = activeButtonStyle.Render(removeBtn)
	} else {
		removeBtn = inactiveButtonStyle.Render(removeBtn)
	}
	if m.focus == 2 {
		cancelBtn = activeButtonStyle.Render(cancelBtn)
	} else {
		cancelBtn = inactiveButtonStyle.Render(cancelBtn)
	}
	b.WriteString(removeBtn + cancelBtn)

	modal := modalStyle.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}
