package context

import "time"

// PaneState tracks which pane is currently visible in the right split.
type PaneState int

const (
	PaneNone     PaneState = iota // no side panel (full-screen mossy)
	PaneMossy                     // mossy-registered worktree pane visible
	PaneExternal                  // external/dormant pane visible
)

type Repository struct {
	Name          string
	Path          string
	WorktreeCount int
}

type ProgramContext struct {
	Width             int
	Height            int
	Repos             []Repository
	ActiveRepo        int
	Message           string
	MessageExpiry     time.Time
	Loading           bool
	AutoRefresh       bool
	LastRefresh       time.Time
	PausedRemaining   int
	ShowHelp          bool
	TmuxPanes         map[string]string // worktree path → tmux pane ID
	TmuxVisiblePane   string            // mossy pane ID when visible
	TmuxDisplacedPane string            // external pane ID being tracked
	TmuxPaneState     PaneState         // current visibility state
}
