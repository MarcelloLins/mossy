package sidepanel

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/marcellolins/mossy/internal/git"
)

var (
	borderStyle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))
)

type Model struct {
	worktree *git.Worktree
}

func New() Model {
	return Model{}
}

func (m *Model) SetWorktree(wt *git.Worktree) {
	m.worktree = wt
}

func (m Model) View(width, height int) string {
	content := ""

	// Pad/truncate to exactly `height` lines
	lines := strings.Split(content, "\n")
	for len(lines) < height {
		lines = append(lines, "")
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	inner := strings.Join(lines, "\n")

	return borderStyle.
		Width(width - 1). // account for left border character
		Render(inner)
}
