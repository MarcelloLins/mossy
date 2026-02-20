package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/tui/context"
)

var (
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8FBC8F")).
			Padding(1, 2, 0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Padding(0, 2)

	addTabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	addTabKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Underline(true)

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true).
			Padding(0, 1)

	borderColor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	overflowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	countStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

type Model struct {
	ctx          *context.ProgramContext
	scrollOffset int
}

func New(ctx *context.ProgramContext) Model {
	return Model{ctx: ctx}
}

func (m Model) tabLabel(i int) string {
	repo := m.ctx.Repos[i]
	if repo.WorktreeCount > 0 {
		return fmt.Sprintf("%s %s", repo.Name, countStyle.Render(fmt.Sprintf("(%d)", repo.WorktreeCount)))
	}
	return repo.Name
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

// ScrollToActive adjusts the scroll offset to ensure the active tab is visible.
func (m *Model) ScrollToActive() {
	if len(m.ctx.Repos) == 0 {
		m.scrollOffset = 0
		return
	}
	if m.ctx.ActiveRepo < m.scrollOffset {
		m.scrollOffset = m.ctx.ActiveRepo
	}

	// Scroll forward until active tab fits in the available width.
	logo := logoStyle.Render("🌿 mossy")
	logoWidth := lipgloss.Width(logo)
	sep := separatorStyle.Render("│")
	sepWidth := lipgloss.Width(sep)
	addTab := lipgloss.NewStyle().Padding(0, 2).Render(addTabStyle.Render("+ ") + addTabKeyStyle.Render("A") + addTabStyle.Render("dd"))
	addWidth := lipgloss.Width(addTab)
	leftIndicator := overflowStyle.Render("◄")
	rightIndicator := overflowStyle.Render("►")

	for m.scrollOffset <= m.ctx.ActiveRepo {
		budget := m.ctx.Width - logoWidth - addWidth - sepWidth

		if m.scrollOffset > 0 {
			budget -= lipgloss.Width(leftIndicator)
		}

		used := 0
		fits := false
		for i := m.scrollOffset; i < len(m.ctx.Repos); i++ {
			var tab string
			if i == m.ctx.ActiveRepo {
				tab = activeTabStyle.Render(m.tabLabel(i))
			} else {
				tab = inactiveTabStyle.Render(m.tabLabel(i))
			}
			w := lipgloss.Width(tab) + sepWidth
			remaining := budget - used - w

			// Reserve space for ► if there are more tabs after this one.
			if i < len(m.ctx.Repos)-1 && remaining < lipgloss.Width(rightIndicator) && i >= m.ctx.ActiveRepo {
				// Active tab doesn't fit with ► indicator — need to scroll more.
				break
			}

			used += w
			if i == m.ctx.ActiveRepo {
				fits = true
				break
			}
			if used > budget {
				break
			}
		}

		if fits {
			break
		}
		m.scrollOffset++
	}
}

func (m Model) View() string {
	logo := logoStyle.Render("🌿 mossy")
	logoWidth := lipgloss.Width(logo)
	sep := separatorStyle.Render("│")
	sepWidth := lipgloss.Width(sep)

	if len(m.ctx.Repos) == 0 {
		tabContent := emptyStyle.Render("No repositories added. Press 'a' to add your first repository.")
		tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, tabContent)
		bar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs,
			lipgloss.PlaceHorizontal(m.ctx.Width-lipgloss.Width(tabs), lipgloss.Right, logo),
		)
		border := borderColor.Render(strings.Repeat("─", m.ctx.Width))
		return bar + "\n" + border
	}

	addTab := lipgloss.NewStyle().Padding(0, 2).Render(addTabStyle.Render("+ ") + addTabKeyStyle.Render("A") + addTabStyle.Render("dd"))
	addWidth := lipgloss.Width(addTab)
	leftIndicator := overflowStyle.Render("◄")
	rightIndicator := overflowStyle.Render("►")

	budget := m.ctx.Width - logoWidth - addWidth - sepWidth

	hasLeft := m.scrollOffset > 0
	if hasLeft {
		budget -= lipgloss.Width(leftIndicator)
	}

	var tabParts []string
	if hasLeft {
		tabParts = append(tabParts, leftIndicator)
	}

	used := 0
	lastVisible := m.scrollOffset
	for i := m.scrollOffset; i < len(m.ctx.Repos); i++ {
		var tab string
		if i == m.ctx.ActiveRepo {
			tab = activeTabStyle.Render(m.tabLabel(i))
		} else {
			tab = inactiveTabStyle.Render(m.tabLabel(i))
		}
		w := lipgloss.Width(tab) + sepWidth

		// Reserve space for ► if this isn't the last repo.
		needed := w
		if i < len(m.ctx.Repos)-1 {
			needed += lipgloss.Width(rightIndicator)
		}
		if used+needed > budget && i > m.scrollOffset {
			break
		}

		tabParts = append(tabParts, tab, sep)
		used += w
		lastVisible = i
	}

	hasRight := lastVisible < len(m.ctx.Repos)-1
	if hasRight {
		tabParts = append(tabParts, rightIndicator)
	}

	tabParts = append(tabParts, addTab)

	tabs := lipgloss.JoinHorizontal(lipgloss.Bottom, tabParts...)
	bar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs,
		lipgloss.PlaceHorizontal(m.ctx.Width-lipgloss.Width(tabs), lipgloss.Right, logo),
	)
	border := borderColor.Render(strings.Repeat("─", m.ctx.Width))

	return bar + "\n" + border
}
