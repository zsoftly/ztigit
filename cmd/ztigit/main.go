// Package main is the entry point for the ztigit CLI
package main

import (
	"bufio"
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
	version = "dev" // overridden at build time via ldflags
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

  # Multiple groups (comma-separated)
  ztigit mirror group1,group2,group3 -p gitlab

  # Multiple groups (space-separated with --groups flag)
  ztigit mirror --groups "group1 group2 group3" -p gitlab

  # Include older repos (default skips repos not updated in 12 months)
  ztigit mirror zsoftly -p github --max-age 24

Repositories are cloned to $HOME/<org>/ by default.
Skips archived repos and repos not updated within --max-age months.
Authentication: Expects GITHUB_TOKEN/GITLAB_TOKEN env vars for API access.
Git operations use your existing git credentials (HTTPS or SSH).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMirror,
}

var (
	mirrorProvider      string
	mirrorDir           string
	mirrorParallel      int
	mirrorVerbose       bool
	mirrorMaxAge        int
	mirrorSkipPreflight bool
	mirrorSSH           bool
	mirrorGroups        string
)

func init() {
	mirrorCmd.Flags().StringVarP(&mirrorProvider, "provider", "p", "", "Provider type: gitlab or github (auto-detected from URL)")
	mirrorCmd.Flags().StringVarP(&mirrorDir, "dir", "d", "", "Base directory (default: $HOME/<org>)")
	mirrorCmd.Flags().IntVar(&mirrorParallel, "parallel", 4, "Number of parallel clone/pull operations")
	mirrorCmd.Flags().BoolVarP(&mirrorVerbose, "verbose", "v", false, "Verbose output")
	mirrorCmd.Flags().IntVar(&mirrorMaxAge, "max-age", 12, "Skip repos not updated in this many months (0 = no limit)")
	mirrorCmd.Flags().BoolVar(&mirrorSkipPreflight, "skip-preflight", false, "Skip git credential validation before cloning")
	mirrorCmd.Flags().BoolVar(&mirrorSSH, "ssh", false, "Use SSH URLs instead of HTTPS for git operations")
	mirrorCmd.Flags().StringVar(&mirrorGroups, "groups", "", "Space-separated list of groups to mirror (e.g., \"group1 group2 group3\")")
	rootCmd.AddCommand(mirrorCmd)
}

func runMirror(cmd *cobra.Command, args []string) error {
	// Check git is installed before doing anything else
	if err := mirror.CheckGitInstalled(); err != nil {
		return err
	}

	ctx := context.Background()

	// Determine groups to mirror
	var groups []string
	var baseURL string
	var providerType provider.ProviderType

	// Case 1: --groups flag provided (space-separated)
	if mirrorGroups != "" {
		groups = strings.Fields(mirrorGroups)

		// Provider must be specified
		if mirrorProvider == "" {
			return fmt.Errorf("--provider required when using --groups flag")
		}
		providerType = provider.ProviderType(mirrorProvider)
		baseURL = cfg.GetBaseURL(string(providerType))
	} else if len(args) == 0 {
		// Case 2: No arguments and no --groups flag
		return fmt.Errorf("either provide a URL/org or use --groups flag")
	} else {
		// Case 3: Argument provided
		target := args[0]

		// Check if comma-separated groups
		if strings.Contains(target, ",") && !strings.HasPrefix(target, "http") {
			// Comma-separated groups: "group1,group2,group3"
			groups = strings.Split(target, ",")
			for i := range groups {
				groups[i] = strings.TrimSpace(groups[i])
			}

			// Provider must be specified
			if mirrorProvider == "" {
				return fmt.Errorf("--provider required when using comma-separated groups")
			}
			providerType = provider.ProviderType(mirrorProvider)
			baseURL = cfg.GetBaseURL(string(providerType))
		} else if strings.HasPrefix(target, "https://") || strings.HasPrefix(target, "http://") {
			// Parse URL: https://github.com/zsoftly -> provider=github, org=zsoftly
			parsed, err := parseGitURL(target)
			if err != nil {
				return err
			}
			baseURL = parsed.baseURL
			groups = []string{parsed.orgName}
			providerType = parsed.provider
		} else {
			// Single org name
			groups = []string{target}
			if mirrorProvider != "" {
				providerType = provider.ProviderType(mirrorProvider)
			} else {
				return fmt.Errorf("provider required when not using URL. Use --provider github or --provider gitlab")
			}
			baseURL = cfg.GetBaseURL(string(providerType))
		}
	}

	// Override provider if explicitly specified
	if mirrorProvider != "" {
		providerType = provider.ProviderType(mirrorProvider)
	}

	// Validate provider type (used in directory paths, must be safe)
	if err := validateProviderType(providerType); err != nil {
		return err
	}

	// Get token from environment or config (optional for public repos)
	token := cfg.GetToken(string(providerType))

	// Security: reject HTTP URLs when token is present
	if err := validateURLSecurity(baseURL, token); err != nil {
		return err
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

	// Configure mirror options
	opts := mirror.Options{
		BaseDir:       mirrorDir,
		Parallel:      mirrorParallel,
		SkipArchived:  true,
		Verbose:       mirrorVerbose,
		MaxAgeMonths:  mirrorMaxAge,
		SkipPreflight: mirrorSkipPreflight,
		SSH:           mirrorSSH,
	}

	// Determine base directory
	if opts.BaseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil || homeDir == "" {
			homeDir = "." // Fallback to current directory
		}

		// For multiple groups, use a common parent directory
		if len(groups) > 1 {
			// Use provider-specific directory: $HOME/gitlab-repos or $HOME/github-repos
			opts.BaseDir = filepath.Join(homeDir, fmt.Sprintf("%s-repos", providerType))
		} else {
			// Single group: use $HOME/<group-name>
			opts.BaseDir = filepath.Join(homeDir, groups[0])
		}
	}

	// Create mirror and run
	m := mirror.New(p, opts)

	fmt.Printf("%s Mirroring %d group(s) to %s\n\n", cyan("→"), len(groups), bold(opts.BaseDir))

	results, err := m.MirrorGroups(ctx, groups)
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

	// Validate that at least one of --provider or --url is specified
	if protectProvider == "" && protectURL == "" {
		return fmt.Errorf("at least one of --provider or --url must be specified")
	}

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
	Long: `Save an authentication token for a provider.

Token is read from environment variable or stdin (never command line for security).

Examples:
  # Using environment variable (recommended)
  export GITHUB_TOKEN=ghp_xxxx
  ztigit auth login -p github

  # Using stdin
  echo $GITHUB_TOKEN | ztigit auth login -p github

  # Interactive (paste token, press Enter)
  ztigit auth login -p gitlab`,
	RunE: runAuthLogin,
}

var (
	authLoginProvider string
	authLoginURL      string
)

func init() {
	authLoginCmd.Flags().StringVarP(&authLoginProvider, "provider", "p", "", "Provider type: gitlab or github")
	authLoginCmd.Flags().StringVarP(&authLoginURL, "url", "u", "", "Base URL for the provider")
	authLoginCmd.MarkFlagRequired("provider")

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

	// Get token from environment variable or stdin (never from command line flag)
	token := getTokenFromEnvOrStdin(providerType)
	if token == "" {
		return fmt.Errorf("no token provided. Set %s_TOKEN environment variable or pipe token via stdin",
			strings.ToUpper(string(providerType)))
	}

	// Security: reject HTTP URLs when token is present
	if err := validateURLSecurity(baseURL, token); err != nil {
		return err
	}

	// Test the token
	var p provider.Provider
	var err error

	switch providerType {
	case provider.ProviderGitLab:
		p, err = provider.NewGitLabProvider(token, baseURL)
	case provider.ProviderGitHub:
		p, err = provider.NewGitHubProvider(token, baseURL)
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
		cfg.GitLab.Token = token
		cfg.GitLab.BaseURL = baseURL
	case provider.ProviderGitHub:
		cfg.GitHub.Token = token
		cfg.GitHub.BaseURL = baseURL
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if config.IsKeyringAvailable() {
		fmt.Printf("\n%s Token stored in system keychain (secure)\n", green("✓"))
	} else {
		fmt.Printf("\n%s Token saved to %s\n", green("✓"), config.GetConfigFile())
		fmt.Printf("  %s Consider using a system with keychain support for better security\n", yellow("!"))
	}
	return nil
}

// validateURLSecurity checks that HTTPS is used when token is present
func validateURLSecurity(baseURL, token string) error {
	if token != "" && strings.HasPrefix(strings.ToLower(baseURL), "http://") {
		return fmt.Errorf("refusing to use HTTP with authentication token (would expose token in plaintext). Use HTTPS instead")
	}
	return nil
}

// validateProviderType ensures the provider type is safe for use in filesystem paths
func validateProviderType(pt provider.ProviderType) error {
	// Provider type must be one of the known types
	switch pt {
	case provider.ProviderGitLab, provider.ProviderGitHub:
		return nil
	default:
		return fmt.Errorf("invalid provider type: %q (must be 'gitlab' or 'github')", pt)
	}
}

// getTokenFromEnvOrStdin reads token from environment variable or stdin
func getTokenFromEnvOrStdin(providerType provider.ProviderType) string {
	// Try environment variable first
	var envVars []string
	switch providerType {
	case provider.ProviderGitLab:
		envVars = []string{"GITLAB_TOKEN", "ZTIGIT_GITLAB_TOKEN"}
	case provider.ProviderGitHub:
		envVars = []string{"GITHUB_TOKEN", "ZTIGIT_GITHUB_TOKEN"}
	}

	for _, env := range envVars {
		if token := os.Getenv(env); token != "" {
			return token
		}
	}

	// Check if stdin has data (piped input)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Data is being piped
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text())
		}
	}

	// Interactive: prompt for token
	fmt.Print("Enter token: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}

	return ""
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
		fmt.Printf("  Token: %s\n", green("***configured***"))
	} else {
		fmt.Println("  Token: (not set)")
	}

	fmt.Println()
	fmt.Println("GitHub:")
	fmt.Printf("  URL:   %s\n", cfg.GitHub.BaseURL)
	if cfg.GitHub.Token != "" {
		fmt.Printf("  Token: %s\n", green("***configured***"))
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
