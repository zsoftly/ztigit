// Package protect handles environment protection operations
package protect

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/zsoftly/ztigit/internal/provider"
)

// Result represents the result of a protection operation
type Result struct {
	Environment provider.Environment
	Action      string // "protected", "skipped", "failed"
	Error       error
}

// Options configures the protect operation
type Options struct {
	AccessLevel       int // 30=developer, 40=maintainer, 60=admin
	RequiredApprovals int
	DryRun            bool
}

// DefaultOptions returns the default protect options
func DefaultOptions() Options {
	return Options{
		AccessLevel:       30, // Developer
		RequiredApprovals: 1,
		DryRun:            false,
	}
}

// Protector handles environment protection operations
type Protector struct {
	provider provider.Provider
	options  Options
}

// New creates a new Protector instance
func New(p provider.Provider, opts Options) *Protector {
	return &Protector{
		provider: p,
		options:  opts,
	}
}

// ProtectEnvironments protects environments matching the pattern
func (p *Protector) ProtectEnvironments(ctx context.Context, projectPath, pattern string) ([]Result, error) {
	// List all environments
	envs, err := p.provider.ListEnvironments(ctx, projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	// Filter environments by pattern
	filtered := filterEnvironments(envs, pattern)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no environments found matching pattern: %s", pattern)
	}

	results := make([]Result, 0, len(filtered))

	for _, env := range filtered {
		result := p.protectEnv(ctx, projectPath, env)
		results = append(results, result)

		// Small delay to avoid API rate limiting
		if !p.options.DryRun && result.Action == "protected" {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return results, nil
}

// protectEnv protects a single environment
func (p *Protector) protectEnv(ctx context.Context, projectPath string, env provider.Environment) Result {
	// Check if already protected
	if env.Protected {
		return Result{
			Environment: env,
			Action:      "skipped",
		}
	}

	if p.options.DryRun {
		return Result{
			Environment: env,
			Action:      "protected",
		}
	}

	// Protect the environment
	rule := provider.ProtectionRule{
		AccessLevel:       p.options.AccessLevel,
		RequiredApprovals: p.options.RequiredApprovals,
	}

	err := p.provider.ProtectEnvironment(ctx, projectPath, env.Name, rule)
	if err != nil {
		return Result{
			Environment: env,
			Action:      "failed",
			Error:       err,
		}
	}

	return Result{
		Environment: env,
		Action:      "protected",
	}
}

// ListEnvironments lists all environments for a project with their protection status
func (p *Protector) ListEnvironments(ctx context.Context, projectPath string) ([]provider.Environment, error) {
	return p.provider.ListEnvironments(ctx, projectPath)
}

// filterEnvironments filters environments by pattern
func filterEnvironments(envs []provider.Environment, pattern string) []provider.Environment {
	if pattern == "all" || pattern == "*" {
		return envs
	}

	// Treat pattern as prefix if no regex metacharacters
	var filtered []provider.Environment

	// Try regex first
	re, err := regexp.Compile("^" + pattern)
	if err != nil {
		// Fall back to prefix matching
		for _, env := range envs {
			if len(env.Name) >= len(pattern) && env.Name[:len(pattern)] == pattern {
				filtered = append(filtered, env)
			}
		}
		return filtered
	}

	for _, env := range envs {
		if re.MatchString(env.Name) {
			filtered = append(filtered, env)
		}
	}

	return filtered
}

// PrintResults prints the protection results to stdout
func PrintResults(results []Result, dryRun bool) {
	var protected, skipped, failed int

	prefix := ""
	if dryRun {
		prefix = "[DRY-RUN] "
	}

	for _, r := range results {
		switch r.Action {
		case "protected":
			protected++
			fmt.Printf("%s[OK] Protected: %s\n", prefix, r.Environment.Name)
		case "skipped":
			skipped++
			fmt.Printf("%s[SKIP] Already protected: %s\n", prefix, r.Environment.Name)
		case "failed":
			failed++
			fmt.Printf("%s[FAIL] Failed: %s - %v\n", prefix, r.Environment.Name, r.Error)
		}
	}

	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  Protected: %d\n", protected)
	fmt.Printf("  Skipped:   %d (already protected)\n", skipped)
	fmt.Printf("  Failed:    %d\n", failed)
	fmt.Printf("  Total:     %d\n", len(results))
}

// PrintEnvironments prints a list of environments
func PrintEnvironments(envs []provider.Environment) {
	fmt.Println("Environments:")
	fmt.Println()

	for _, env := range envs {
		status := "unprotected"
		if env.Protected {
			status = "protected"
		}
		fmt.Printf("  %-40s [%s] %s\n", env.Name, status, env.State)
	}

	fmt.Println()
	fmt.Printf("Total: %d environments\n", len(envs))
}
