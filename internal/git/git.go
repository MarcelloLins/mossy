package git

import (
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	HEAD   string
	Branch string
	Bare   bool
}

func ListWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseWorktrees(string(out)), nil
}

func parseWorktrees(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree
	inEntry := false

	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			if inEntry {
				worktrees = append(worktrees, current)
				current = Worktree{}
				inEntry = false
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			current.Path = strings.TrimPrefix(line, "worktree ")
			inEntry = true
		} else if strings.HasPrefix(line, "HEAD ") {
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(branch, "refs/heads/")
		} else if line == "bare" {
			current.Bare = true
		} else if line == "detached" {
			current.Branch = "(detached)"
		}
	}

	if inEntry {
		worktrees = append(worktrees, current)
	}

	return worktrees
}
