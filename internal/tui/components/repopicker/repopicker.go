package repopicker

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RepoSelectedMsg struct {
	Name string
	Path string
}

type RepoPickerCancelledMsg struct{}

type dirEntry struct {
	name      string
	path      string
	isGitRepo bool
	isAdded   bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#8FBC8F")).
			Padding(0, 1)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	dirStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	gitTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8FBC8F"))

	addedTagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Padding(0, 2)

	errStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))
)

type Model struct {
	currentDir string
	entries    []dirEntry
	cursor     int
	offset     int
	width      int
	height     int
	err        error
	existing   map[string]struct{}
}

func New(startDir string, width, height int, existingPaths []string) Model {
	existing := make(map[string]struct{}, len(existingPaths))
	for _, p := range existingPaths {
		existing[p] = struct{}{}
	}
	m := Model{
		currentDir: startDir,
		width:      width,
		height:     height,
		existing:   existing,
	}
	m.readDir()
	return m
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) readDir() {
	m.entries = nil
	m.cursor = 0
	m.offset = 0

	if m.currentDir != "/" {
		m.entries = append(m.entries, dirEntry{
			name: "..",
			path: filepath.Dir(m.currentDir),
		})
	}

	dirEntries, err := os.ReadDir(m.currentDir)
	if err != nil {
		m.err = err
		return
	}
	m.err = nil

	var dirs []dirEntry
	for _, e := range dirEntries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		p := filepath.Join(m.currentDir, e.Name())
		_, gitErr := os.Stat(filepath.Join(p, ".git"))
		isGit := gitErr == nil
		_, added := m.existing[p]
		dirs = append(dirs, dirEntry{
			name:      e.Name(),
			path:      p,
			isGitRepo: isGit,
			isAdded:   isGit && added,
		})
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].name < dirs[j].name
	})

	m.entries = append(m.entries, dirs...)
}

func (m Model) visibleRows() int {
	rows := m.height - 6
	if rows < 1 {
		rows = 1
	}
	return rows
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return RepoPickerCancelledMsg{} }
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset = m.cursor
				}
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				if m.cursor >= m.offset+m.visibleRows() {
					m.offset = m.cursor - m.visibleRows() + 1
				}
			}
		case "enter":
			if m.cursor < 0 || m.cursor >= len(m.entries) {
				break
			}
			entry := m.entries[m.cursor]
			if entry.name == ".." || !entry.isGitRepo {
				m.currentDir = entry.path
				m.readDir()
			} else if entry.isAdded {
				// Already added — do nothing
			} else {
				name := filepath.Base(entry.path)
				path := entry.path
				return m, func() tea.Msg {
					return RepoSelectedMsg{Name: name, Path: path}
				}
			}
		}
	}
	return m, nil
}

func (m Model) displayPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return m.currentDir
	}
	if strings.HasPrefix(m.currentDir, home) {
		return "~" + m.currentDir[len(home):]
	}
	return m.currentDir
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("📂 " + m.displayPath()))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errStyle.Render(fmt.Sprintf("  Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	visible := m.visibleRows()
	end := m.offset + visible
	if end > len(m.entries) {
		end = len(m.entries)
	}

	for i := m.offset; i < end; i++ {
		entry := m.entries[i]

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s 📁 %s", cursor, entry.name)
		if entry.isAdded {
			line += "  " + addedTagStyle.Render("✓ added")
		} else if entry.isGitRepo {
			line += "  " + gitTagStyle.Render("✓ git")
		}

		if i == m.cursor {
			b.WriteString(cursorStyle.Render(line))
		} else {
			b.WriteString(dirStyle.Render(line))
		}
		b.WriteString("\n")
	}

	for i := end - m.offset; i < visible; i++ {
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: open/select • esc: cancel"))

	return b.String()
}
