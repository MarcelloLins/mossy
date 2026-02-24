package keys

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	Left           key.Binding
	Right          key.Binding
	AddRepo        key.Binding
	DeleteRepo     key.Binding
	NewWorktree    key.Binding
	RemoveWorktree key.Binding
	PrevCommit     key.Binding
	NextCommit     key.Binding
	Refresh        key.Binding
	AutoRefresh    key.Binding
	TmuxPane       key.Binding
	Help           key.Binding
	Quit           key.Binding
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev repo"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next repo"),
	),
	AddRepo: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add repo"),
	),
	DeleteRepo: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "remove repo"),
	),
	NewWorktree: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new worktree"),
	),
	RemoveWorktree: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "remove worktree"),
	),
	PrevCommit: key.NewBinding(
		key.WithKeys("["),
		key.WithHelp("[", "prev commit"),
	),
	NextCommit: key.NewBinding(
		key.WithKeys("]"),
		key.WithHelp("]", "next commit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	AutoRefresh: key.NewBinding(
		key.WithKeys("R"),
		key.WithHelp("R", "toggle auto-refresh"),
	),
	TmuxPane: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle tmux pane"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.AddRepo, k.DeleteRepo, k.NewWorktree, k.RemoveWorktree},
		{k.PrevCommit, k.NextCommit, k.Refresh, k.AutoRefresh},
		{k.TmuxPane, k.Help, k.Quit},
	}
}
