package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var aiAgentPattern = regexp.MustCompile(`(?im)Co-authored-by:\s+(Copilot|Goose|Claude|Cursor|Amp)\b`)

type Worktree struct {
	Path      string
	HEAD      string
	Branch    string
	Bare      bool
	Additions int
	Deletions int
}

type Commit struct {
	Hash      string
	Subject   string
	Body      string
	Author    string
	Date      string
	Pushed    bool
	Tags      []string
	Files     []string
	Additions int
	Deletions int
	AIAgents  []string
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

func AddWorktree(repoPath, name, branch string) error {
	wtPath := filepath.Join(filepath.Dir(repoPath), name)
	cmd := exec.Command("git", "worktree", "add", wtPath, "-b", branch)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return parseWorktreeError(string(out), name, branch)
	}
	return nil
}

func RemoveWorktree(repoPath, wtPath, branch string, deleteBranch bool) error {
	cmd := exec.Command("git", "worktree", "remove", wtPath)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	if deleteBranch && branch != "" && branch != "(detached)" {
		cmd = exec.Command("git", "branch", "-D", branch)
		cmd.Dir = repoPath
		out, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("worktree removed but branch deletion failed: %s", strings.TrimSpace(string(out)))
		}
	}
	return nil
}

func ListCommits(repoPath, branch string) ([]Commit, error) {
	defaultBranch := detectDefaultBranch(repoPath)
	revRange := defaultBranch + ".." + branch
	// Use %x00 at the end as record separator; %x01 separates fields within.
	// Format: hash\x01subject\x01author\x01date\x01body\x00
	cmd := exec.Command("git", "log", revRange,
		"--format=%H%x01%s%x01%an%x01%ar%x01%b%x00", "--")
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(string(out))
	if text == "" {
		return nil, nil
	}
	records := strings.Split(text, "\x00")
	var commits []Commit
	for _, rec := range records {
		rec = strings.TrimSpace(rec)
		if rec == "" {
			continue
		}
		fields := strings.SplitN(rec, "\x01", 5)
		if len(fields) < 4 {
			continue
		}
		c := Commit{
			Hash:    fields[0],
			Subject: fields[1],
			Author:  fields[2],
			Date:    fields[3],
		}
		if len(fields) == 5 {
			c.Body = strings.TrimSpace(fields[4])
		}
		commits = append(commits, c)
	}
	enrichCommits(repoPath, branch, commits)
	return commits, nil
}

func enrichCommits(repoPath, branch string, commits []Commit) {
	// Determine which commits are pushed by finding the remote tip.
	remoteTip := ""
	cmd := exec.Command("git", "rev-parse", "--verify", "origin/"+branch)
	cmd.Dir = repoPath
	if out, err := cmd.Output(); err == nil {
		remoteTip = strings.TrimSpace(string(out))
	}

	// If we have a remote tip, all commits that are ancestors of it are pushed.
	pushedSet := make(map[string]bool)
	if remoteTip != "" {
		// Get the set of commits in default..branch that are also in default..origin/branch
		defaultBranch := detectDefaultBranch(repoPath)
		cmd := exec.Command("git", "log", defaultBranch+"..origin/"+branch, "--format=%H")
		cmd.Dir = repoPath
		if out, err := cmd.Output(); err == nil {
			for _, h := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if h != "" {
					pushedSet[h] = true
				}
			}
		}
	}

	for i := range commits {
		commits[i].Pushed = pushedSet[commits[i].Hash]

		// Tags
		cmd := exec.Command("git", "tag", "--points-at", commits[i].Hash)
		cmd.Dir = repoPath
		if out, err := cmd.Output(); err == nil {
			for _, t := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if t != "" {
					commits[i].Tags = append(commits[i].Tags, t)
				}
			}
		}

		// Diff stat and file list
		cmd = exec.Command("git", "diff-tree", "--no-commit-id", "--numstat", "-r", commits[i].Hash)
		cmd.Dir = repoPath
		if out, err := cmd.Output(); err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if line == "" {
					continue
				}
				fields := strings.Fields(line)
				if len(fields) < 3 {
					continue
				}
				if a, err := strconv.Atoi(fields[0]); err == nil {
					commits[i].Additions += a
				}
				if d, err := strconv.Atoi(fields[1]); err == nil {
					commits[i].Deletions += d
				}
				commits[i].Files = append(commits[i].Files, fields[2])
			}
		}

		// AI co-authors (check both subject and body — some commits
		// have the trailer squashed into the subject line)
		full := commits[i].Subject + "\n" + commits[i].Body
		matches := aiAgentPattern.FindAllStringSubmatch(full, -1)
		seen := make(map[string]bool)
		for _, m := range matches {
			name := m[1]
			if !seen[name] {
				seen[name] = true
				commits[i].AIAgents = append(commits[i].AIAgents, name)
			}
		}
	}
}

func parseWorktreeError(output, name, branch string) error {
	// Extract only the fatal line — git prefixes output with
	// "Preparing worktree ..." which can contain misleading keywords.
	fatal := ""
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "fatal:") {
			fatal = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "fatal:"))
			break
		}
	}
	if fatal == "" {
		fatal = strings.TrimSpace(output)
	}
	switch {
	case strings.Contains(fatal, "branch named") && strings.Contains(fatal, "already exists"):
		return fmt.Errorf("a branch named %q already exists", branch)
	case strings.Contains(fatal, "already exists"):
		return fmt.Errorf("a worktree named %q already exists", name)
	case strings.Contains(fatal, "not a valid branch name"):
		return fmt.Errorf("%q is not a valid branch name", branch)
	case strings.Contains(fatal, "is a missing but locked"):
		return fmt.Errorf("worktree %q is locked; unlock it first", name)
	default:
		return fmt.Errorf("%s", fatal)
	}
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

// RebaseOnto fetches the latest default branch from origin and rebases the
// worktree's branch onto it. If the rebase encounters conflicts it is
// automatically aborted and an error is returned.
func RebaseOnto(repoPath, wtPath string) error {
	defaultBranch := detectDefaultBranch(repoPath)

	fetch := exec.Command("git", "fetch", "origin", defaultBranch)
	fetch.Dir = wtPath
	if out, err := fetch.CombinedOutput(); err != nil {
		return fmt.Errorf("fetch failed: %s", strings.TrimSpace(string(out)))
	}

	rebase := exec.Command("git", "rebase", "origin/"+defaultBranch)
	rebase.Dir = wtPath
	if out, err := rebase.CombinedOutput(); err != nil {
		abort := exec.Command("git", "rebase", "--abort")
		abort.Dir = wtPath
		abort.Run()
		return fmt.Errorf("conflicts detected — resolve in terminal (%s)", strings.TrimSpace(string(out)))
	}
	return nil
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
