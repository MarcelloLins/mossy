package context

import "time"

type Repository struct {
	Name          string
	Path          string
	WorktreeCount int
}

type ProgramContext struct {
	Width           int
	Height          int
	Repos           []Repository
	ActiveRepo      int
	Message         string
	MessageExpiry   time.Time
	Loading         bool
	AutoRefresh     bool
	LastRefresh     time.Time
	PausedRemaining int
	ShowHelp        bool
}
