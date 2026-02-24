package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Sessions maps worktree paths to tmux pane IDs.
type Sessions struct {
	Panes map[string]string `json:"panes"`
}

func sessionsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "mossy", "sessions.json"), nil
}

func LoadSessions() (Sessions, error) {
	p, err := sessionsPath()
	if err != nil {
		return Sessions{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return Sessions{Panes: make(map[string]string)}, nil
		}
		return Sessions{}, err
	}
	var s Sessions
	if err := json.Unmarshal(data, &s); err != nil {
		return Sessions{Panes: make(map[string]string)}, nil
	}
	if s.Panes == nil {
		s.Panes = make(map[string]string)
	}
	return s, nil
}

func SaveSessions(s Sessions) error {
	p, err := sessionsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
