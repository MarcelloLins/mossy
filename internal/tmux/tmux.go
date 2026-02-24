package tmux

import (
	"os"
	"os/exec"
	"strings"
)

const sessionName = "mossy-worktrees"

// InsideTmux returns true if the current process is running inside tmux.
func InsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func sessionExists() bool {
	return exec.Command("tmux", "has-session", "-t", sessionName).Run() == nil
}

// CreateWindow creates a pane with a shell in the given directory, parked
// inside a separate tmux session called "mossy-worktrees" so it stays
// invisible in the user's window list.
// Returns the pane ID.
func CreateWindow(dir string) (string, error) {
	var cmd *exec.Cmd
	if sessionExists() {
		cmd = exec.Command("tmux", "new-window", "-d", "-t", sessionName+":", "-c", dir, "-P", "-F", "#{pane_id}")
	} else {
		cmd = exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-c", dir, "-P", "-F", "#{pane_id}")
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	paneID := strings.TrimSpace(string(out))
	// Respawn to guarantee the shell starts in the correct directory,
	// even if the user's shell config overrides -c.
	exec.Command("tmux", "respawn-pane", "-k", "-c", dir, "-t", paneID).Run()
	return paneID, nil
}

// JoinPane moves the specified pane into the current window as a horizontal
// split on the right. The pane is not focused.
func JoinPane(paneID string) error {
	cmd := exec.Command("tmux", "join-pane", "-h", "-d", "-s", paneID)
	return cmd.Run()
}

// BreakPane moves the specified pane back into the mossy-worktrees session.
// If the session was destroyed (all panes were joined out), it is recreated.
func BreakPane(paneID string) error {
	if sessionExists() {
		return exec.Command("tmux", "join-pane", "-d", "-s", paneID, "-t", sessionName+":").Run()
	}
	// Session was destroyed — recreate it with a dummy pane, move ours in, kill the dummy.
	out, err := exec.Command("tmux", "new-session", "-d", "-s", sessionName, "-P", "-F", "#{pane_id}").Output()
	if err != nil {
		return err
	}
	dummyPane := strings.TrimSpace(string(out))
	if err := exec.Command("tmux", "join-pane", "-d", "-s", paneID, "-t", sessionName+":").Run(); err != nil {
		exec.Command("tmux", "kill-session", "-t", sessionName).Run()
		return err
	}
	exec.Command("tmux", "kill-pane", "-t", dummyPane).Run()
	return nil
}

// SwapPane swaps two panes in-place without changing the layout.
// No resize events are triggered.
func SwapPane(paneA, paneB string) error {
	cmd := exec.Command("tmux", "swap-pane", "-d", "-s", paneA, "-t", paneB)
	return cmd.Run()
}

// KillPane kills a tmux pane by its ID.
func KillPane(paneID string) error {
	cmd := exec.Command("tmux", "kill-pane", "-t", paneID)
	return cmd.Run()
}

// PaneExists returns true if the given pane ID exists.
func PaneExists(paneID string) bool {
	cmd := exec.Command("tmux", "display-message", "-t", paneID, "-p", "#{pane_id}")
	return cmd.Run() == nil
}
