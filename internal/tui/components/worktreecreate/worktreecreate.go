package worktreecreate

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorktreeCreateRequestMsg struct {
	Name   string
	Branch string
}

type WorktreeCreateCancelledMsg struct{}

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
)

type Model struct {
	nameInput   textinput.Model
	branchInput textinput.Model
	focus       int
	width       int
	height      int
}

func New(width, height int) Model {
	ni := textinput.New()
	ni.Placeholder = "my-feature"
	ni.Focus()
	ni.CharLimit = 128
	ni.Width = modalWidth - 8

	bi := textinput.New()
	bi.Placeholder = "feature/my-feature"
	bi.CharLimit = 128
	bi.Width = modalWidth - 8

	return Model{
		nameInput:   ni,
		branchInput: bi,
		focus:       0,
		width:       width,
		height:      height,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) updateFocus() {
	m.nameInput.Blur()
	m.branchInput.Blur()

	switch m.focus {
	case 0:
		m.nameInput.Focus()
	case 1:
		m.branchInput.Focus()
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return WorktreeCreateCancelledMsg{} }
		case "tab":
			m.focus = (m.focus + 1) % 4
			m.updateFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + 3) % 4
			m.updateFocus()
			return m, nil
		case "up":
			if m.focus > 0 {
				if m.focus == 3 {
					m.focus = 2
				} else {
					m.focus--
				}
				m.updateFocus()
			}
			return m, nil
		case "down":
			if m.focus < 2 {
				m.focus++
				m.updateFocus()
			}
			return m, nil
		case "left":
			if m.focus == 3 {
				m.focus = 2
				return m, nil
			}
		case "right":
			if m.focus == 2 {
				m.focus = 3
				return m, nil
			}
		case "enter":
			switch m.focus {
			case 0:
				m.focus = 1
				m.updateFocus()
				return m, nil
			case 1, 2:
				name := m.nameInput.Value()
				branch := m.branchInput.Value()
				return m, func() tea.Msg {
					return WorktreeCreateRequestMsg{Name: name, Branch: branch}
				}
			case 3:
				return m, func() tea.Msg { return WorktreeCreateCancelledMsg{} }
			}
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case 0:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case 1:
		m.branchInput, cmd = m.branchInput.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New Worktree"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Worktree name"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(m.nameInput.View()))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Branch name"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(m.branchInput.View()))
	b.WriteString("\n\n")

	createBtn := "[Create]"
	cancelBtn := "[Cancel]"
	if m.focus == 2 {
		createBtn = activeButtonStyle.Render(createBtn)
	} else {
		createBtn = inactiveButtonStyle.Render(createBtn)
	}
	if m.focus == 3 {
		cancelBtn = activeButtonStyle.Render(cancelBtn)
	} else {
		cancelBtn = inactiveButtonStyle.Render(cancelBtn)
	}
	b.WriteString(createBtn + cancelBtn)

	modal := modalStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
}
