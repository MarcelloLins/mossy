package git

import (
	"os/exec"
	"strconv"
	"strings"
)

type Worktree struct {
	Path      string
	HEAD      string
	Branch    string
	Bare      bool
	Additions int
	Deletions int
}

func ListWorktrees(repoPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	all := parseWorktrees(string(out))
	if len(all) > 0 {
		all = all[1:]
	}
	defaultBranch := detectDefaultBranch(repoPath)
	for i := range all {
		if all[i].Branch != "" && all[i].Branch != defaultBranch && all[i].Branch != "(detached)" {
			a, d := diffStats(repoPath, defaultBranch, all[i].HEAD)
			all[i].Additions = a
			all[i].Deletions = d
		}
	}
	return all, nil
}

func detectDefaultBranch(repoPath string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(out))
		return strings.TrimPrefix(ref, "refs/remotes/origin/")
	}
	for _, name := range []string{"main", "master"} {
		check := exec.Command("git", "rev-parse", "--verify", "refs/heads/"+name)
		check.Dir = repoPath
		if err := check.Run(); err == nil {
			return name
		}
	}
	return "main"
}

func diffStats(repoPath, base, head string) (additions, deletions int) {
	cmd := exec.Command("git", "diff", "--numstat", base+"..."+head)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return 0, 0
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		if a, err := strconv.Atoi(fields[0]); err == nil {
			additions += a
		}
		if d, err := strconv.Atoi(fields[1]); err == nil {
			deletions += d
		}
	}
	return additions, deletions
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
