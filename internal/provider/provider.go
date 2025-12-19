// Package provider defines the interface for Git hosting providers (GitLab, GitHub)
package provider

import (
	"context"
	"strings"
	"time"
)

// Repository represents a git repository from any provider
type Repository struct {
	ID            int64
	Name          string
	FullPath      string // e.g., "group/subgroup/repo" or "org/repo"
	CloneURL      string // HTTPS clone URL
	SSHUrl        string // SSH clone URL
	DefaultBranch string
	Archived      bool
	LastUpdated   time.Time // Last activity/push date
	Size          int64     // Size in bytes
}

// Group represents a group/organization from any provider
type Group struct {
	ID       int64
	Name     string
	FullPath string
}

// Environment represents a deployment environment (GitLab-specific, but abstracted)
type Environment struct {
	ID        int64
	Name      string
	State     string // available, stopped, etc.
	Protected bool
}

// ProtectionRule defines environment protection settings
type ProtectionRule struct {
	AccessLevel       int // 30=developer, 40=maintainer, 60=admin
	RequiredApprovals int
}

// Provider is the interface that all git hosting providers must implement
type Provider interface {
	// Name returns the provider name (e.g., "gitlab", "github")
	Name() string

	// TestConnection tests the API connection and token validity
	TestConnection(ctx context.Context) error

	// GetCurrentUser returns the authenticated user's username
	GetCurrentUser(ctx context.Context) (string, error)

	// ListGroupProjects lists all projects/repos in a group/org (including subgroups)
	ListGroupProjects(ctx context.Context, groupPath string) ([]Repository, error)

	// ListGroups lists all accessible groups/orgs
	ListGroups(ctx context.Context) ([]Group, error)

	// GetProject gets a single project by path
	GetProject(ctx context.Context, projectPath string) (*Repository, error)

	// Environment operations (may not be supported by all providers)
	ListEnvironments(ctx context.Context, projectPath string) ([]Environment, error)
	ProtectEnvironment(ctx context.Context, projectPath, envName string, rule ProtectionRule) error
	IsEnvironmentProtected(ctx context.Context, projectPath, envName string) (bool, error)
}

// ProviderType represents the type of git hosting provider
type ProviderType string

const (
	ProviderGitLab ProviderType = "gitlab"
	ProviderGitHub ProviderType = "github"
)

// DetectProvider attempts to detect the provider type from a URL
func DetectProvider(url string) ProviderType {
	if url == "" {
		return ProviderGitHub // default to GitHub
	}

	// Check for common patterns
	if strings.Contains(url, "gitlab") {
		return ProviderGitLab
	}
	if strings.Contains(url, "github") {
		return ProviderGitHub
	}

	// Default to GitLab for self-hosted instances
	return ProviderGitLab
}
