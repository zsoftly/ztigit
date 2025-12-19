// Package mirror handles repository mirroring operations
package mirror

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/zsoftly/ztigit/internal/provider"
)

// Colors for output
var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	faint  = color.New(color.Faint).SprintFunc()
)

// formatSize formats bytes into human-readable size
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// Result represents the result of a mirror operation
type Result struct {
	Repository provider.Repository
	Action     string // "cloned", "updated", "skipped", "failed"
	Error      error
	Duration   time.Duration
}

// Options configures the mirror operation
type Options struct {
	BaseDir       string
	Parallel      int
	SkipArchived  bool
	Verbose       bool
	MaxAgeMonths  int  // Skip repos not updated in this many months (0 = no limit)
	SkipPreflight bool // Skip credential validation before cloning
	SSH           bool // Use SSH URLs instead of HTTPS for git operations
}

// DefaultOptions returns the default mirror options
func DefaultOptions() Options {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = "." // Fallback to current directory
	}
	return Options{
		BaseDir:      filepath.Join(homeDir, "git-repos"),
		Parallel:     4,
		SkipArchived: true,
		Verbose:      false,
		MaxAgeMonths: 12,
	}
}

// Mirror mirrors repositories from a provider
type Mirror struct {
	provider provider.Provider
	options  Options
}

// New creates a new Mirror instance
func New(p provider.Provider, opts Options) *Mirror {
	if opts.Parallel < 1 {
		opts.Parallel = 1
	}
	return &Mirror{
		provider: p,
		options:  opts,
	}
}

// MirrorGroups mirrors all repositories from the specified groups
func (m *Mirror) MirrorGroups(ctx context.Context, groups []string) ([]Result, error) {
	var allRepos []provider.Repository

	for _, group := range groups {
		fmt.Printf("%s Fetching repos from %s...\n", cyan("→"), bold(group))
		repos, err := m.provider.ListGroupProjects(ctx, group)
		if err != nil {
			return nil, fmt.Errorf("failed to list projects for group %s: %w", group, err)
		}

		// Calculate total size
		var totalSize int64
		for _, r := range repos {
			totalSize += r.Size
		}

		fmt.Printf("%s Found %s repos %s\n\n", cyan("→"), bold(fmt.Sprintf("%d", len(repos))), faint("("+formatSize(totalSize)+")"))
		allRepos = append(allRepos, repos...)
	}

	// Preflight credential check
	if len(allRepos) > 0 && !m.options.SkipPreflight {
		fmt.Printf("%s Checking git credentials...\n", cyan("→"))
		result, err := m.Preflight(ctx, allRepos)
		if err != nil {
			return nil, err
		}
		// Use the method that works - override SSH if needed
		if result.Method == "ssh" && !m.options.SSH {
			m.options.SSH = true
			fmt.Printf("%s HTTPS unavailable, using SSH\n\n", green("✓"))
		} else {
			fmt.Printf("%s Git credentials OK (%s)\n\n", green("✓"), strings.ToUpper(result.Method))
		}
	}

	return m.mirrorRepos(ctx, allRepos)
}

// MirrorRepos mirrors the specified repositories
func (m *Mirror) mirrorRepos(ctx context.Context, repos []provider.Repository) ([]Result, error) {
	results := make([]Result, 0, len(repos))
	resultsChan := make(chan Result, len(repos))
	semaphore := make(chan struct{}, m.options.Parallel)

	// Calculate cutoff date for stale repos
	var cutoffDate time.Time
	if m.options.MaxAgeMonths > 0 {
		cutoffDate = time.Now().AddDate(0, -m.options.MaxAgeMonths, 0)
	}

	var wg sync.WaitGroup

	for _, repo := range repos {
		if m.options.SkipArchived && repo.Archived {
			resultsChan <- Result{
				Repository: repo,
				Action:     "skipped",
			}
			continue
		}

		// Skip stale repos (not updated within MaxAgeMonths)
		if m.options.MaxAgeMonths > 0 && !repo.LastUpdated.IsZero() && repo.LastUpdated.Before(cutoffDate) {
			resultsChan <- Result{
				Repository: repo,
				Action:     "stale",
			}
			continue
		}

		wg.Add(1)
		go func(r provider.Repository) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				resultsChan <- Result{
					Repository: r,
					Action:     "failed",
					Error:      ctx.Err(),
				}
				return
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			}

			start := time.Now()
			result := m.mirrorRepo(ctx, r)
			result.Duration = time.Since(start)
			resultsChan <- result
		}(repo)
	}

	// Wait for all goroutines and close channel
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	for result := range resultsChan {
		results = append(results, result)
	}

	return results, nil
}

// mirrorRepo clones or updates a single repository
func (m *Mirror) mirrorRepo(ctx context.Context, repo provider.Repository) Result {
	// Clone directly into BaseDir/<repo-name>
	repoDir := filepath.Join(m.options.BaseDir, repo.Name)

	// Format size string
	sizeStr := ""
	if repo.Size > 0 {
		sizeStr = " " + faint("("+formatSize(repo.Size)+")")
	}

	// Check if repository already exists
	if isGitRepo(repoDir) {
		fmt.Printf("  %s %s%s\n", cyan("↻"), repo.Name, sizeStr)
		err := m.updateRepo(ctx, repoDir)
		if err != nil {
			return Result{
				Repository: repo,
				Action:     "failed",
				Error:      fmt.Errorf("update failed: %w", err),
			}
		}
		return Result{
			Repository: repo,
			Action:     "updated",
		}
	}

	// Clone the repository - order depends on SSH option
	fmt.Printf("  %s %s%s\n", cyan("↓"), repo.Name, sizeStr)

	var primaryURL, fallbackURL string
	var primaryMethod, fallbackMethod string

	if m.options.SSH {
		primaryURL, fallbackURL = repo.SSHUrl, repo.CloneURL
		primaryMethod, fallbackMethod = "SSH", "HTTPS"
	} else {
		primaryURL, fallbackURL = repo.CloneURL, repo.SSHUrl
		primaryMethod, fallbackMethod = "HTTPS", "SSH"
	}

	err := m.cloneRepo(ctx, primaryURL, repoDir)
	if err != nil {
		// Try fallback if primary fails
		if fallbackURL != "" {
			fmt.Printf("    %s %s failed, trying %s...\n", yellow("!"), primaryMethod, fallbackMethod)
			fallbackErr := m.cloneRepo(ctx, fallbackURL, repoDir)
			if fallbackErr != nil {
				return Result{
					Repository: repo,
					Action:     "failed",
					Error:      fmt.Errorf("clone failed (%s: %v, %s: %v)", primaryMethod, err, fallbackMethod, fallbackErr),
				}
			}
			return Result{
				Repository: repo,
				Action:     "cloned",
			}
		}
		return Result{
			Repository: repo,
			Action:     "failed",
			Error:      fmt.Errorf("clone failed: %w", err),
		}
	}

	return Result{
		Repository: repo,
		Action:     "cloned",
	}
}

// cloneRepo clones a repository to the specified directory
func (m *Mirror) cloneRepo(ctx context.Context, url, dir string) error {
	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", url, dir)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if m.options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

// updateRepo updates an existing repository
func (m *Mirror) updateRepo(ctx context.Context, dir string) error {
	// Fetch all remotes
	fetchCmd := exec.CommandContext(ctx, "git", "-C", dir, "fetch", "--all")
	fetchCmd.Stdout = nil
	fetchCmd.Stderr = nil

	if m.options.Verbose {
		fetchCmd.Stdout = os.Stdout
		fetchCmd.Stderr = os.Stderr
	}

	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Get the default branch from git
	branch, err := m.getDefaultBranch(ctx, dir)
	if err != nil {
		return err
	}

	// Check if there are local changes
	statusCmd := exec.CommandContext(ctx, "git", "-C", dir, "status", "--porcelain")
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git status failed: %w", err)
	}

	if len(statusOutput) > 0 {
		// Stash local changes
		stashCmd := exec.CommandContext(ctx, "git", "-C", dir, "stash", "push", "-m", "ztigit auto-stash")
		stashCmd.Stdout = nil
		stashCmd.Stderr = nil
		_ = stashCmd.Run() // Ignore errors, might not have anything to stash
	}

	// Switch to default branch
	if err := m.checkoutBranch(ctx, dir, branch); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", branch, err)
	}

	// Pull latest changes
	pullCmd := exec.CommandContext(ctx, "git", "-C", dir, "pull", "origin", branch)
	pullCmd.Stdout = nil
	pullCmd.Stderr = nil

	if m.options.Verbose {
		pullCmd.Stdout = os.Stdout
		pullCmd.Stderr = os.Stderr
	}

	if err := pullCmd.Run(); err != nil {
		// Try reset to origin if pull fails
		resetCmd := exec.CommandContext(ctx, "git", "-C", dir, "reset", "--hard", "origin/"+branch)
		resetCmd.Stdout = nil
		resetCmd.Stderr = nil
		if resetErr := resetCmd.Run(); resetErr != nil {
			return fmt.Errorf("git pull and reset failed: %w", err)
		}
	}

	return nil
}

// getDefaultBranch gets the default branch from git
func (m *Mirror) getDefaultBranch(ctx context.Context, dir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--abbrev-ref", "origin/HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get default branch: %w", err)
	}
	// Output is "origin/main" - strip the "origin/" prefix
	ref := strings.TrimSpace(string(output))
	return strings.TrimPrefix(ref, "origin/"), nil
}

// checkoutBranch switches to a branch, creating it from remote if needed
func (m *Mirror) checkoutBranch(ctx context.Context, dir, branch string) error {
	// First try simple checkout (branch exists locally)
	checkoutCmd := exec.CommandContext(ctx, "git", "-C", dir, "checkout", branch)
	checkoutCmd.Stdout = nil
	checkoutCmd.Stderr = nil
	if err := checkoutCmd.Run(); err == nil {
		return nil
	}

	// Branch doesn't exist locally, create from remote
	createCmd := exec.CommandContext(ctx, "git", "-C", dir, "checkout", "-b", branch, "origin/"+branch)
	createCmd.Stdout = nil
	createCmd.Stderr = nil
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("branch %s not found locally or on remote", branch)
	}

	return nil
}

// isGitRepo checks if a directory is a git repository
func isGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// PrintResults prints the mirror results to stdout
func PrintResults(results []Result) {
	var cloned, updated, skipped, stale, failed int

	fmt.Println()
	for _, r := range results {
		switch r.Action {
		case "cloned":
			cloned++
			fmt.Printf("  %s %s %s\n", green("✓"), r.Repository.Name, faint(r.Duration.Round(time.Millisecond).String()))
		case "updated":
			updated++
			fmt.Printf("  %s %s %s\n", green("✓"), r.Repository.Name, faint(r.Duration.Round(time.Millisecond).String()))
		case "skipped":
			skipped++
			fmt.Printf("  %s %s %s\n", yellow("○"), r.Repository.Name, faint("(archived)"))
		case "stale":
			stale++
			fmt.Printf("  %s %s %s\n", yellow("○"), r.Repository.Name, faint("(stale: "+r.Repository.LastUpdated.Format("2006-01-02")+")"))
		case "failed":
			failed++
			fmt.Printf("  %s %s %s\n", red("✗"), r.Repository.Name, faint(r.Error.Error()))
		}
	}

	fmt.Println()
	fmt.Printf("%s\n", bold("Summary"))
	if cloned > 0 {
		fmt.Printf("  %s Cloned:  %d\n", green("✓"), cloned)
	}
	if updated > 0 {
		fmt.Printf("  %s Updated: %d\n", green("✓"), updated)
	}
	if skipped > 0 {
		fmt.Printf("  %s Skipped: %d (archived)\n", yellow("○"), skipped)
	}
	if stale > 0 {
		fmt.Printf("  %s Stale:   %d\n", yellow("○"), stale)
	}
	if failed > 0 {
		fmt.Printf("  %s Failed:  %d\n", red("✗"), failed)
	}
	fmt.Printf("  Total:   %d\n", len(results))
}
