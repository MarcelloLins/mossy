package sidepanel

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/git"
)

var (
	borderStyle = lipgloss.NewStyle().
			BorderLeft(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	navStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	navActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E")).
			Bold(true)

	navDimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8FBC8F")).
			Bold(true)

	hashStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E"))

	subjectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	bodyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFBD2E")).
			Bold(true)

	pushedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8FBC8F"))

	localStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	addStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	delStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	emptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

type Model struct {
	worktree *git.Worktree
	commits  []git.Commit
	Cursor   int
}

func New() Model {
	return Model{}
}

func (m *Model) SetWorktree(wt *git.Worktree) {
	m.worktree = wt
}

func (m *Model) SetCommits(commits []git.Commit) {
	m.commits = commits
	m.Cursor = 0
}

func (m *Model) NextCommit() {
	if m.Cursor < len(m.commits)-1 {
		m.Cursor++
	}
}

func (m *Model) PrevCommit() {
	if m.Cursor > 0 {
		m.Cursor--
	}
}

func (m Model) renderNav(width int) string {
	if len(m.commits) == 0 {
		return navStyle.Width(width).Render("")
	}

	var left, right string
	if m.Cursor > 0 {
		left = navActiveStyle.Render("◄ [")
	} else {
		left = navDimStyle.Render("◄ [")
	}
	if m.Cursor < len(m.commits)-1 {
		right = navActiveStyle.Render("] ►")
	} else {
		right = navDimStyle.Render("] ►")
	}

	counter := metaStyle.Render(fmt.Sprintf("%d / %d", m.Cursor+1, len(m.commits)))
	nav := left + " " + counter + " " + right

	return navStyle.Width(width).Render(nav)
}

func (m Model) View(width, height int) string {
	contentWidth := width - 1 - 2 // border char + padding
	if contentWidth < 0 {
		contentWidth = 0
	}

	nav := m.renderNav(contentWidth)
	navHeight := lipgloss.Height(nav)

	bodyHeight := height - navHeight
	if bodyHeight < 0 {
		bodyHeight = 0
	}

	var lines []string

	if len(m.commits) == 0 {
		lines = append(lines, "")
		lines = append(lines, emptyStyle.Render("No commits ahead of default branch"))
	} else {
		c := m.commits[m.Cursor]

		// Status (pushed / local)
		lines = append(lines, "")
		if c.Pushed {
			lines = append(lines, pushedStyle.Render("⬆ pushed"))
		} else {
			lines = append(lines, localStyle.Render("● local only"))
		}

		// Tags
		if len(c.Tags) > 0 {
			var tagParts []string
			for _, t := range c.Tags {
				tagParts = append(tagParts, tagStyle.Render("🏷 "+t))
			}
			lines = append(lines, strings.Join(tagParts, " "))
		}

		// Commit hash
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Commit"))
		lines = append(lines, hashStyle.Render(c.Hash))

		// Subject
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Subject"))
		wrapped := lipgloss.NewStyle().Width(contentWidth).Render(subjectStyle.Render(c.Subject))
		lines = append(lines, wrapped)

		// Author + Date
		lines = append(lines, "")
		lines = append(lines, labelStyle.Render("Author"))
		lines = append(lines, metaStyle.Render(c.Author+" · "+c.Date))

		// Diff stat
		if c.Additions > 0 || c.Deletions > 0 {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("Changes"))
			stat := addStyle.Render(fmt.Sprintf("+%d", c.Additions)) +
				metaStyle.Render(" / ") +
				delStyle.Render(fmt.Sprintf("-%d", c.Deletions)) +
				metaStyle.Render(fmt.Sprintf(" in %d file(s)", len(c.Files)))
			lines = append(lines, stat)
		}

		// Body
		if c.Body != "" {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("Message"))
			bodyWrapped := lipgloss.NewStyle().Width(contentWidth).Render(bodyStyle.Render(c.Body))
			lines = append(lines, bodyWrapped)
		}

		// Files
		if len(c.Files) > 0 {
			lines = append(lines, "")
			lines = append(lines, labelStyle.Render("Files"))
			for _, f := range c.Files {
				lines = append(lines, fileStyle.Render("  "+f))
			}
		}
	}

	// Pad/truncate body to bodyHeight
	for len(lines) < bodyHeight {
		lines = append(lines, "")
	}
	if len(lines) > bodyHeight {
		lines = lines[:bodyHeight]
	}

	inner := nav + "\n" + lipgloss.NewStyle().Padding(0, 1).Render(strings.Join(lines, "\n"))

	// Final pad/truncate to exactly height
	allLines := strings.Split(inner, "\n")
	for len(allLines) < height {
		allLines = append(allLines, "")
	}
	if len(allLines) > height {
		allLines = allLines[:height]
	}

	return borderStyle.
		Width(width - 1).
		Render(strings.Join(allLines, "\n"))
}
