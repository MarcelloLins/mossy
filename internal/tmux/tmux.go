package tmux

import (
	"os"
	"os/exec"
	"strings"
)

const windowName = "mossy-worktrees"

// InsideTmux returns true if the current process is running inside tmux.
func InsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

func mossyWindowExists() bool {
	cmd := exec.Command("tmux", "list-windows", "-F", "#{window_name}")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == windowName {
			return true
		}
	}
	return false
}

// CreateWindow creates a pane with a shell in the given directory, parked
// inside a shared hidden window called "mossy-worktrees".
// Returns the pane ID.
func CreateWindow(dir string) (string, error) {
	cmd := exec.Command("tmux", "new-window", "-d", "-c", dir, "-P", "-F", "#{pane_id}")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	paneID := strings.TrimSpace(string(out))
	// Respawn to guarantee the shell starts in the correct directory,
	// even if the user's shell config overrides -c.
	exec.Command("tmux", "respawn-pane", "-k", "-c", dir, "-t", paneID).Run()

	if mossyWindowExists() {
		// Move the pane into the existing mossy-worktrees window.
		exec.Command("tmux", "join-pane", "-d", "-s", paneID, "-t", windowName).Run()
	} else {
		// This is the first pane — just rename the window.
		exec.Command("tmux", "rename-window", "-t", paneID, windowName).Run()
	}
	return paneID, nil
}

// JoinPane moves the specified pane into the current window as a horizontal
// split on the right. The pane is not focused.
func JoinPane(paneID string) error {
	cmd := exec.Command("tmux", "join-pane", "-h", "-d", "-s", paneID)
	return cmd.Run()
}

// BreakPane moves the specified pane back into the shared mossy-worktrees
// window, or creates it if it doesn't exist.
func BreakPane(paneID string) error {
	if mossyWindowExists() {
		return exec.Command("tmux", "join-pane", "-d", "-s", paneID, "-t", windowName).Run()
	}
	return exec.Command("tmux", "break-pane", "-d", "-s", paneID, "-n", windowName).Run()
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
