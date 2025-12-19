package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
)

// GitHubProvider implements the Provider interface for GitHub
type GitHubProvider struct {
	client  *github.Client
	baseURL string
}

// NewGitHubProvider creates a new GitHub provider instance
// Token is optional - unauthenticated access works for public repos (with rate limits)
func NewGitHubProvider(token, baseURL string) (*GitHubProvider, error) {
	var client *github.Client

	if token != "" {
		client = github.NewClient(nil).WithAuthToken(token)
	} else {
		client = github.NewClient(nil) // Unauthenticated - works for public repos
	}

	// Handle GitHub Enterprise
	if baseURL != "" && !strings.Contains(baseURL, "github.com") {
		baseURL = strings.TrimSuffix(baseURL, "/")
		apiURL := baseURL + "/api/v3/"
		uploadURL := baseURL + "/api/uploads/"

		parsedAPI, err := url.Parse(apiURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		parsedUpload, err := url.Parse(uploadURL)
		if err != nil {
			return nil, fmt.Errorf("invalid upload URL: %w", err)
		}

		client.BaseURL = parsedAPI
		client.UploadURL = parsedUpload
	}

	if baseURL == "" {
		baseURL = "https://github.com"
	}

	return &GitHubProvider{
		client:  client,
		baseURL: baseURL,
	}, nil
}

// Name returns the provider name
func (p *GitHubProvider) Name() string {
	return "github"
}

// TestConnection tests the API connection and token validity
func (p *GitHubProvider) TestConnection(ctx context.Context) error {
	_, _, err := p.client.Users.Get(ctx, "")
	if err != nil {
		return fmt.Errorf("GitHub connection test failed: %w", err)
	}
	return nil
}

// GetCurrentUser returns the authenticated user's username
func (p *GitHubProvider) GetCurrentUser(ctx context.Context) (string, error) {
	user, _, err := p.client.Users.Get(ctx, "")
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return user.GetLogin(), nil
}

// ListGroupProjects lists all repositories in an organization or user account
func (p *GitHubProvider) ListGroupProjects(ctx context.Context, ownerName string) ([]Repository, error) {
	var repos []Repository

	// Try organization first, fall back to user repos
	orgRepos, err := p.listOrgRepos(ctx, ownerName)
	if err == nil {
		return orgRepos, nil
	}

	// If org fails, try as user
	userRepos, userErr := p.listUserRepos(ctx, ownerName)
	if userErr != nil {
		return nil, fmt.Errorf("failed to list repositories for %s (org error: %v, user error: %w)", ownerName, err, userErr)
	}

	repos = append(repos, userRepos...)
	return repos, nil
}

// listOrgRepos lists repositories for an organization
func (p *GitHubProvider) listOrgRepos(ctx context.Context, orgName string) ([]Repository, error) {
	var repos []Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		ghRepos, resp, err := p.client.Repositories.ListByOrg(ctx, orgName, opts)
		if err != nil {
			return nil, err
		}

		for _, repo := range ghRepos {
			var lastUpdated time.Time
			if repo.PushedAt != nil {
				lastUpdated = repo.PushedAt.Time
			}
			repos = append(repos, Repository{
				ID:            repo.GetID(),
				Name:          repo.GetName(),
				FullPath:      repo.GetFullName(),
				CloneURL:      repo.GetCloneURL(),
				SSHUrl:        repo.GetSSHURL(),
				DefaultBranch: repo.GetDefaultBranch(),
				Archived:      repo.GetArchived(),
				LastUpdated:   lastUpdated,
				Size:          int64(repo.GetSize()) * 1024, // GitHub returns KB, convert to bytes
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

// listUserRepos lists repositories for a user
func (p *GitHubProvider) listUserRepos(ctx context.Context, username string) ([]Repository, error) {
	var repos []Repository

	opts := &github.RepositoryListByUserOptions{
		Type: "owner", // Only repos owned by user, not forks
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		ghRepos, resp, err := p.client.Repositories.ListByUser(ctx, username, opts)
		if err != nil {
			return nil, err
		}

		for _, repo := range ghRepos {
			var lastUpdated time.Time
			if repo.PushedAt != nil {
				lastUpdated = repo.PushedAt.Time
			}
			repos = append(repos, Repository{
				ID:            repo.GetID(),
				Name:          repo.GetName(),
				FullPath:      repo.GetFullName(),
				CloneURL:      repo.GetCloneURL(),
				SSHUrl:        repo.GetSSHURL(),
				DefaultBranch: repo.GetDefaultBranch(),
				Archived:      repo.GetArchived(),
				LastUpdated:   lastUpdated,
				Size:          int64(repo.GetSize()) * 1024, // GitHub returns KB, convert to bytes
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

// ListGroups lists all accessible organizations
func (p *GitHubProvider) ListGroups(ctx context.Context) ([]Group, error) {
	var groups []Group

	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		orgs, resp, err := p.client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list organizations: %w", err)
		}

		for _, org := range orgs {
			groups = append(groups, Group{
				ID:       org.GetID(),
				Name:     org.GetLogin(),
				FullPath: org.GetLogin(),
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return groups, nil
}

// GetProject gets a single repository by path (owner/repo)
func (p *GitHubProvider) GetProject(ctx context.Context, projectPath string) (*Repository, error) {
	parts := strings.SplitN(projectPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid project path: %s (expected owner/repo)", projectPath)
	}

	owner, repoName := parts[0], parts[1]

	repo, _, err := p.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s: %w", projectPath, err)
	}

	return &Repository{
		ID:            repo.GetID(),
		Name:          repo.GetName(),
		FullPath:      repo.GetFullName(),
		CloneURL:      repo.GetCloneURL(),
		SSHUrl:        repo.GetSSHURL(),
		DefaultBranch: repo.GetDefaultBranch(),
		Archived:      repo.GetArchived(),
	}, nil
}

// ListEnvironments lists all environments for a repository
func (p *GitHubProvider) ListEnvironments(ctx context.Context, projectPath string) ([]Environment, error) {
	parts := strings.SplitN(projectPath, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid project path: %s (expected owner/repo)", projectPath)
	}

	owner, repoName := parts[0], parts[1]

	var envs []Environment
	opts := &github.EnvironmentListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		envResponse, resp, err := p.client.Repositories.ListEnvironments(ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list environments for %s: %w", projectPath, err)
		}

		for _, e := range envResponse.Environments {
			protected := false
			if e.ProtectionRules != nil && len(e.ProtectionRules) > 0 {
				protected = true
			}

			envs = append(envs, Environment{
				ID:        e.GetID(),
				Name:      e.GetName(),
				State:     "available", // GitHub doesn't have the same state concept
				Protected: protected,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return envs, nil
}

// ProtectEnvironment protects an environment with the given rules
// Note: GitHub's environment protection works differently than GitLab
func (p *GitHubProvider) ProtectEnvironment(ctx context.Context, projectPath, envName string, rule ProtectionRule) error {
	parts := strings.SplitN(projectPath, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid project path: %s (expected owner/repo)", projectPath)
	}

	owner, repoName := parts[0], parts[1]

	// Create or update environment with protection
	createEnv := &github.CreateUpdateEnvironment{
		// GitHub uses reviewers for approval, different from GitLab's access levels
		// For now, we'll just create the environment without specific reviewers
		// as that requires team/user IDs
	}

	_, _, err := p.client.Repositories.CreateUpdateEnvironment(ctx, owner, repoName, envName, createEnv)
	if err != nil {
		return fmt.Errorf("failed to protect environment %s: %w", envName, err)
	}

	return nil
}

// IsEnvironmentProtected checks if an environment has protection rules
func (p *GitHubProvider) IsEnvironmentProtected(ctx context.Context, projectPath, envName string) (bool, error) {
	parts := strings.SplitN(projectPath, "/", 2)
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid project path: %s (expected owner/repo)", projectPath)
	}

	owner, repoName := parts[0], parts[1]

	env, resp, err := p.client.Repositories.GetEnvironment(ctx, owner, repoName, envName)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check protection status for %s: %w", envName, err)
	}

	return env.ProtectionRules != nil && len(env.ProtectionRules) > 0, nil
}
