// Package main is the entry point for the ztigit CLI
package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/zsoftly/ztigit/internal/config"
	"github.com/zsoftly/ztigit/internal/mirror"
	"github.com/zsoftly/ztigit/internal/protect"
	"github.com/zsoftly/ztigit/internal/provider"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

var (
	version = "0.0.1"
	cfg     *config.Config
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:     "ztigit",
	Short:   "ZSoftly Tools for Git - Multi-platform Git hosting CLI",
	Long:    `ztigit is a cross-platform CLI tool for managing GitLab and GitHub repositories.`,
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		return nil
	},
}

// Mirror command
var mirrorCmd = &cobra.Command{
	Use:   "mirror <url-or-org>",
	Short: "Mirror repositories from groups/organizations",
	Long: `Clone or update repositories from GitLab groups or GitHub organizations.

Examples:
  # Auto-detect from URL
  ztigit mirror https://github.com/zsoftly
  ztigit mirror https://gitlab.com/my-group

  # Specify provider manually
  ztigit mirror zsoftly --provider github

  # Include older repos (default skips repos not updated in 12 months)
  ztigit mirror zsoftly -p github --max-age 24

Repositories are cloned to $HOME/<org>/ by default.
Skips archived repos and repos not updated within --max-age months.
Authentication: Expects GITHUB_TOKEN/GITLAB_TOKEN env vars for API access.
Git operations use your existing git credentials (HTTPS or SSH).`,
	Args: cobra.ExactArgs(1),
	RunE: runMirror,
}

var (
	mirrorProvider      string
	mirrorDir           string
	mirrorParallel      int
	mirrorVerbose       bool
	mirrorMaxAge        int
	mirrorSkipPreflight bool
)

func init() {
	mirrorCmd.Flags().StringVarP(&mirrorProvider, "provider", "p", "", "Provider type: gitlab or github (auto-detected from URL)")
	mirrorCmd.Flags().StringVarP(&mirrorDir, "dir", "d", "", "Base directory (default: $HOME/<org>)")
	mirrorCmd.Flags().IntVar(&mirrorParallel, "parallel", 4, "Number of parallel clone/pull operations")
	mirrorCmd.Flags().BoolVarP(&mirrorVerbose, "verbose", "v", false, "Verbose output")
	mirrorCmd.Flags().IntVar(&mirrorMaxAge, "max-age", 12, "Skip repos not updated in this many months (0 = no limit)")
	mirrorCmd.Flags().BoolVar(&mirrorSkipPreflight, "skip-preflight", false, "Skip git credential validation before cloning")
	rootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	target := args[0]
	var baseURL, orgName string
	var providerType provider.ProviderType

	// Parse target: URL or org name
	if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") {
		// Parse URL: https://github.com/zsoftly -> provider=github, org=zsoftly
		parsed, err := parseGitURL(target)
		if err != nil {
			return err
		}
		baseURL = parsed.baseURL
		orgName = parsed.orgName
		providerType = parsed.provider
	} else {
		// Just org name, provider must be specified or detected from config
		orgName = target
		if mirrorProvider != "" {
			providerType = provider.ProviderType(mirrorProvider)
		} else {
			return fmt.Errorf("provider required when not using URL. Use --provider github or --provider gitlab")
		}
		baseURL = cfg.GetBaseURL(string(providerType))
	}

	// Override provider if explicitly specified
	if mirrorProvider != "" {
		providerType = provider.ProviderType(mirrorProvider)
	}

	// Get token from environment or config (optional for public repos)
	token := cfg.GetToken(string(providerType))

	// Create provider
	var p provider.Provider
	var err error

	switch providerType {
	case provider.ProviderGitLab:
		p, err = provider.NewGitLabProvider(token, baseURL)
	case provider.ProviderGitHub:
		p, err = provider.NewGitHubProvider(token, baseURL)
	default:
		return fmt.Errorf("unknown provider: %s (use 'gitlab' or 'github')", providerType)
	}

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Test connection (skip auth test if no token)
	fmt.Printf("%s Connecting to %s\n", cyan("→"), bold(baseURL))
	if token != "" {
		if err := p.TestConnection(ctx); err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}
		user, _ := p.GetCurrentUser(ctx)
		fmt.Printf("%s Authenticated as %s\n\n", green("✓"), bold(user))
	} else {
		fmt.Printf("%s No token - public repos only\n\n", yellow("!"))
	}

	// Configure mirror options - default to $HOME/<org>/
	opts := mirror.Options{
		BaseDir:       mirrorDir,
		Parallel:      mirrorParallel,
		SkipArchived:  true,
		Verbose:       mirrorVerbose,
		MaxAgeMonths:  mirrorMaxAge,
		SkipPreflight: mirrorSkipPreflight,
	}

	if opts.BaseDir == "" {
		homeDir, _ := os.UserHomeDir()
		opts.BaseDir = filepath.Join(homeDir, orgName)
	}

	// Create mirror and run
	m := mirror.New(p, opts)

	fmt.Printf("%s Mirroring to %s\n\n", cyan("→"), bold(opts.BaseDir))

	results, err := m.MirrorGroups(ctx, []string{orgName})
	if err != nil {
		return err
	}

	mirror.PrintResults(results)
	return nil
}

// parsedURL holds parsed git hosting URL components
type parsedURL struct {
	baseURL  string
	orgName  string
	provider provider.ProviderType
}

// parseGitURL parses a URL like https://github.com/zsoftly
func parseGitURL(rawURL string) (*parsedURL, error) {
	// Remove trailing slash
	rawURL = strings.TrimSuffix(rawURL, "/")

	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Extract org/group from path
	path := strings.Trim(u.Path, "/")
	if path == "" {
		return nil, fmt.Errorf("URL must include organization/group (e.g., https://github.com/zsoftly)")
	}

	// Take first path segment as org name
	parts := strings.SplitN(path, "/", 2)
	orgName := parts[0]

	// Determine base URL and provider
	baseURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	providerType := provider.DetectProvider(baseURL)

	return &parsedURL{
		baseURL:  baseURL,
		orgName:  orgName,
		provider: providerType,
	}, nil
}

// Protect command
var protectCmd = &cobra.Command{
	Use:   "protect",
	Short: "Protect environments",
	Long:  `Protect deployment environments matching a pattern.`,
	RunE:  runProtect,
}

var (
	protectProject   string
	protectPattern   string
	protectURL       string
	protectProvider  string
	protectDryRun    bool
	protectAccessLvl int
	protectApprovals int
)

func init() {
	protectCmd.Flags().StringVarP(&protectProject, "project", "P", "", "Project path (e.g., group/project)")
	protectCmd.Flags().StringVar(&protectPattern, "pattern", "", "Environment name pattern (e.g., 'dev', 'prod', 'all')")
	protectCmd.Flags().StringVarP(&protectURL, "url", "u", "", "Git hosting URL")
	protectCmd.Flags().StringVarP(&protectProvider, "provider", "p", "", "Provider type: gitlab or github")
	protectCmd.Flags().BoolVar(&protectDryRun, "dry-run", false, "Show what would be protected without making changes")
	protectCmd.Flags().IntVar(&protectAccessLvl, "access-level", 30, "Access level required (30=developer, 40=maintainer, 60=admin)")
	protectCmd.Flags().IntVar(&protectApprovals, "approvals", 1, "Required approvals")
	protectCmd.MarkFlagRequired("project")
	protectCmd.MarkFlagRequired("pattern")
	rootCmd.AddCommand(protectCmd)
}

func runProtect(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine provider
	providerType := provider.ProviderType(protectProvider)
	if protectProvider == "" {
		providerType = provider.DetectProvider(protectURL)
	}

	// Get token and URL
	token := cfg.GetToken(string(providerType))
	baseURL := protectURL
	if baseURL == "" {
		baseURL = cfg.GetBaseURL(string(providerType))
	}

	if token == "" {
		return fmt.Errorf("no token configured for %s", providerType)
	}

	// Create provider
	var p provider.Provider
	var err error

	switch providerType {
	case provider.ProviderGitLab:
		p, err = provider.NewGitLabProvider(token, baseURL)
	case provider.ProviderGitHub:
		p, err = provider.NewGitHubProvider(token, baseURL)
	default:
		return fmt.Errorf("unknown provider: %s", providerType)
	}

	if err != nil {
		return err
	}

	// Test connection
	if err := p.TestConnection(ctx); err != nil {
		return err
	}

	// Configure protect options
	opts := protect.Options{
		AccessLevel:       protectAccessLvl,
		RequiredApprovals: protectApprovals,
		DryRun:            protectDryRun,
	}

	// Create protector and run
	pr := protect.New(p, opts)

	if protectDryRun {
		fmt.Println("[DRY-RUN] The following environments would be protected:")
		fmt.Println()
	}

	results, err := pr.ProtectEnvironments(ctx, protectProject, protectPattern)
	if err != nil {
		return err
	}

	protect.PrintResults(results, protectDryRun)
	return nil
}

// Environments command
var envsCmd = &cobra.Command{
	Use:   "environments",
	Short: "List environments for a project",
	Long:  `List all deployment environments and their protection status.`,
	RunE:  runEnvironments,
}

var (
	envsProject  string
	envsURL      string
	envsProvider string
)

func init() {
	envsCmd.Flags().StringVarP(&envsProject, "project", "P", "", "Project path (e.g., group/project)")
	envsCmd.Flags().StringVarP(&envsURL, "url", "u", "", "Git hosting URL")
	envsCmd.Flags().StringVarP(&envsProvider, "provider", "p", "", "Provider type: gitlab or github")
	envsCmd.MarkFlagRequired("project")
	rootCmd.AddCommand(envsCmd)
}

func runEnvironments(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Determine provider
	providerType := provider.ProviderType(envsProvider)
	if envsProvider == "" {
		providerType = provider.DetectProvider(envsURL)
	}

	// Get token and URL
	token := cfg.GetToken(string(providerType))
	baseURL := envsURL
	if baseURL == "" {
		baseURL = cfg.GetBaseURL(string(providerType))
	}

	if token == "" {
		return fmt.Errorf("no token configured for %s", providerType)
	}

	// Create provider
	var p provider.Provider
	var err error

	switch providerType {
	case provider.ProviderGitLab:
		p, err = provider.NewGitLabProvider(token, baseURL)
	case provider.ProviderGitHub:
		p, err = provider.NewGitHubProvider(token, baseURL)
	default:
		return fmt.Errorf("unknown provider: %s", providerType)
	}

	if err != nil {
		return err
	}

	// Test connection
	if err := p.TestConnection(ctx); err != nil {
		return err
	}

	// List environments
	pr := protect.New(p, protect.DefaultOptions())
	envs, err := pr.ListEnvironments(ctx, envsProject)
	if err != nil {
		return err
	}

	protect.PrintEnvironments(envs)
	return nil
}

// Auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  `Configure authentication tokens for GitLab and GitHub.`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure authentication token",
	Long:  `Save an authentication token for a provider.`,
	RunE:  runAuthLogin,
}

var (
	authLoginProvider string
	authLoginToken    string
	authLoginURL      string
)

func init() {
	authLoginCmd.Flags().StringVarP(&authLoginProvider, "provider", "p", "", "Provider type: gitlab or github")
	authLoginCmd.Flags().StringVarP(&authLoginToken, "token", "t", "", "Authentication token")
	authLoginCmd.Flags().StringVarP(&authLoginURL, "url", "u", "", "Base URL for the provider")
	authLoginCmd.MarkFlagRequired("provider")
	authLoginCmd.MarkFlagRequired("token")

	authCmd.AddCommand(authLoginCmd)
	rootCmd.AddCommand(authCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Validate provider
	providerType := provider.ProviderType(authLoginProvider)
	if providerType != provider.ProviderGitLab && providerType != provider.ProviderGitHub {
		return fmt.Errorf("invalid provider: %s (must be 'gitlab' or 'github')", authLoginProvider)
	}

	// Determine base URL
	baseURL := authLoginURL
	if baseURL == "" {
		if providerType == provider.ProviderGitLab {
			baseURL = "https://gitlab.com"
		} else {
			baseURL = "https://github.com"
		}
	}

	// Test the token
	var p provider.Provider
	var err error

	switch providerType {
	case provider.ProviderGitLab:
		p, err = provider.NewGitLabProvider(authLoginToken, baseURL)
	case provider.ProviderGitHub:
		p, err = provider.NewGitHubProvider(authLoginToken, baseURL)
	}

	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	fmt.Printf("Testing connection to %s...\n", baseURL)
	if err := p.TestConnection(ctx); err != nil {
		return fmt.Errorf("token validation failed: %w", err)
	}

	user, _ := p.GetCurrentUser(ctx)
	fmt.Printf("Authenticated as: %s\n", user)

	// Save to config
	switch providerType {
	case provider.ProviderGitLab:
		cfg.GitLab.Token = authLoginToken
		cfg.GitLab.BaseURL = baseURL
	case provider.ProviderGitHub:
		cfg.GitHub.Token = authLoginToken
		cfg.GitHub.BaseURL = baseURL
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n[OK] Token saved to %s\n", config.GetConfigFile())
	return nil
}

// Config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show configuration",
	Long:  `Display the current configuration.`,
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	fmt.Printf("Configuration file: %s\n\n", config.GetConfigFile())

	fmt.Println("GitLab:")
	fmt.Printf("  URL:   %s\n", cfg.GitLab.BaseURL)
	if cfg.GitLab.Token != "" {
		fmt.Printf("  Token: %s...%s\n", cfg.GitLab.Token[:4], cfg.GitLab.Token[len(cfg.GitLab.Token)-4:])
	} else {
		fmt.Println("  Token: (not set)")
	}

	fmt.Println()
	fmt.Println("GitHub:")
	fmt.Printf("  URL:   %s\n", cfg.GitHub.BaseURL)
	if cfg.GitHub.Token != "" {
		fmt.Printf("  Token: %s...%s\n", cfg.GitHub.Token[:4], cfg.GitHub.Token[len(cfg.GitHub.Token)-4:])
	} else {
		fmt.Println("  Token: (not set)")
	}

	fmt.Println()
	fmt.Println("Mirror:")
	fmt.Printf("  Base directory: %s\n", cfg.Mirror.BaseDir)
	fmt.Printf("  Parallel:       %d\n", cfg.Mirror.Parallel)
	fmt.Printf("  Skip archived:  %t\n", cfg.Mirror.SkipArchived)

	return nil
}
