package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcellolins/mossy/internal/tui/components/repopicker"
	"github.com/marcellolins/mossy/internal/tui/components/tabs"
	"github.com/marcellolins/mossy/internal/tui/context"
)

type viewState int

const (
	viewNormal viewState = iota
	viewRepoPicker
)

type Model struct {
	ctx        *context.ProgramContext
	tabs       tabs.Model
	repoPicker repopicker.Model
	view       viewState
}

func New() Model {
	ctx := &context.ProgramContext{
		ActiveRepo: -1,
	}
	return Model{
		ctx:  ctx,
		tabs: tabs.New(ctx),
		view: viewNormal,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.SetWindowTitle("mossy")
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ctx.Width = msg.Width
		m.ctx.Height = msg.Height
		if m.view == viewRepoPicker {
			m.repoPicker.SetSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if m.view == viewRepoPicker {
		switch msg := msg.(type) {
		case repopicker.RepoSelectedMsg:
			m.ctx.Repos = append(m.ctx.Repos, context.Repository{
				Name: msg.Name,
				Path: msg.Path,
			})
			m.ctx.ActiveRepo = len(m.ctx.Repos) - 1
			m.view = viewNormal
			return m, nil
		case repopicker.RepoPickerCancelledMsg:
			m.view = viewNormal
			return m, nil
		default:
			var cmd tea.Cmd
			m.repoPicker, cmd = m.repoPicker.Update(msg)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "a":
			home, err := os.UserHomeDir()
			if err != nil {
				home = "/"
			}
			var paths []string
			for _, r := range m.ctx.Repos {
				paths = append(paths, r.Path)
			}
			m.repoPicker = repopicker.New(home, m.ctx.Width, m.ctx.Height, paths)
			m.view = viewRepoPicker
			return m, nil
		case "h", "left":
			if len(m.ctx.Repos) > 0 && m.ctx.ActiveRepo > 0 {
				m.ctx.ActiveRepo--
			}
		case "l", "right":
			if len(m.ctx.Repos) > 0 && m.ctx.ActiveRepo < len(m.ctx.Repos)-1 {
				m.ctx.ActiveRepo++
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.view == viewRepoPicker {
		return m.repoPicker.View()
	}
	return m.tabs.View() + "\n"
}
