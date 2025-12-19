package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

// GitLabProvider implements the Provider interface for GitLab
type GitLabProvider struct {
	client  *gitlab.Client
	baseURL string
}

// NewGitLabProvider creates a new GitLab provider instance
func NewGitLabProvider(token, baseURL string) (*GitLabProvider, error) {
	var client *gitlab.Client
	var err error

	// Normalize base URL
	if baseURL == "" {
		baseURL = "https://gitlab.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Ensure we have the API path
	apiURL := baseURL
	if !strings.HasSuffix(apiURL, "/api/v4") {
		apiURL = baseURL + "/api/v4"
	}

	// Create client with base URL
	client, err = gitlab.NewClient(token, gitlab.WithBaseURL(apiURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &GitLabProvider{
		client:  client,
		baseURL: baseURL,
	}, nil
}

// Name returns the provider name
func (p *GitLabProvider) Name() string {
	return "gitlab"
}

// TestConnection tests the API connection and token validity
func (p *GitLabProvider) TestConnection(ctx context.Context) error {
	_, _, err := p.client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("GitLab connection test failed: %w", err)
	}
	return nil
}

// GetCurrentUser returns the authenticated user's username
func (p *GitLabProvider) GetCurrentUser(ctx context.Context) (string, error) {
	user, _, err := p.client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	return user.Username, nil
}

// ListGroupProjects lists all projects in a group including subgroups
func (p *GitLabProvider) ListGroupProjects(ctx context.Context, groupPath string) ([]Repository, error) {
	var repos []Repository

	// URL encode the group path
	encodedPath := url.PathEscape(groupPath)

	opts := &gitlab.ListGroupProjectsOptions{
		IncludeSubGroups: gitlab.Ptr(true),
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	for {
		projects, resp, err := p.client.Groups.ListGroupProjects(encodedPath, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list projects for group %s: %w", groupPath, err)
		}

		for _, project := range projects {
			var lastUpdated time.Time
			if project.LastActivityAt != nil {
				lastUpdated = *project.LastActivityAt
			}
			var size int64
			if project.Statistics != nil {
				size = project.Statistics.RepositorySize
			}
			repos = append(repos, Repository{
				ID:            int64(project.ID),
				Name:          project.Name,
				FullPath:      project.PathWithNamespace,
				CloneURL:      project.HTTPURLToRepo,
				SSHUrl:        project.SSHURLToRepo,
				DefaultBranch: project.DefaultBranch,
				Archived:      project.Archived,
				LastUpdated:   lastUpdated,
				Size:          size,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return repos, nil
}

// ListGroups lists all accessible groups
func (p *GitLabProvider) ListGroups(ctx context.Context) ([]Group, error) {
	var groups []Group

	opts := &gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	for {
		gitlabGroups, resp, err := p.client.Groups.ListGroups(opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list groups: %w", err)
		}

		for _, g := range gitlabGroups {
			groups = append(groups, Group{
				ID:       int64(g.ID),
				Name:     g.Name,
				FullPath: g.FullPath,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return groups, nil
}

// GetProject gets a single project by path
func (p *GitLabProvider) GetProject(ctx context.Context, projectPath string) (*Repository, error) {
	encodedPath := url.PathEscape(projectPath)

	project, _, err := p.client.Projects.GetProject(encodedPath, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}

	return &Repository{
		ID:            int64(project.ID),
		Name:          project.Name,
		FullPath:      project.PathWithNamespace,
		CloneURL:      project.HTTPURLToRepo,
		SSHUrl:        project.SSHURLToRepo,
		DefaultBranch: project.DefaultBranch,
		Archived:      project.Archived,
	}, nil
}

// ListEnvironments lists all environments for a project
func (p *GitLabProvider) ListEnvironments(ctx context.Context, projectPath string) ([]Environment, error) {
	var envs []Environment

	encodedPath := url.PathEscape(projectPath)

	opts := &gitlab.ListEnvironmentsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 100,
		},
	}

	for {
		gitlabEnvs, resp, err := p.client.Environments.ListEnvironments(encodedPath, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, fmt.Errorf("failed to list environments for %s: %w", projectPath, err)
		}

		for _, e := range gitlabEnvs {
			envs = append(envs, Environment{
				ID:    int64(e.ID),
				Name:  e.Name,
				State: e.State,
			})
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Check which environments are protected
	protectedEnvs, err := p.listProtectedEnvironments(ctx, projectPath)
	if err != nil {
		// Non-fatal: some projects may not have this feature
		return envs, nil
	}

	protectedSet := make(map[string]bool)
	for _, pe := range protectedEnvs {
		protectedSet[pe] = true
	}

	for i := range envs {
		envs[i].Protected = protectedSet[envs[i].Name]
	}

	return envs, nil
}

// listProtectedEnvironments returns names of protected environments
func (p *GitLabProvider) listProtectedEnvironments(ctx context.Context, projectPath string) ([]string, error) {
	encodedPath := url.PathEscape(projectPath)

	protectedEnvs, _, err := p.client.ProtectedEnvironments.ListProtectedEnvironments(encodedPath, nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	names := make([]string, len(protectedEnvs))
	for i, pe := range protectedEnvs {
		names[i] = pe.Name
	}

	return names, nil
}

// ProtectEnvironment protects an environment with the given rules
func (p *GitLabProvider) ProtectEnvironment(ctx context.Context, projectPath, envName string, rule ProtectionRule) error {
	encodedPath := url.PathEscape(projectPath)

	opts := &gitlab.ProtectRepositoryEnvironmentsOptions{
		Name: gitlab.Ptr(envName),
		DeployAccessLevels: &[]*gitlab.EnvironmentAccessOptions{
			{
				AccessLevel: gitlab.Ptr(gitlab.AccessLevelValue(rule.AccessLevel)),
			},
		},
	}

	if rule.RequiredApprovals > 0 {
		opts.ApprovalRules = &[]*gitlab.EnvironmentApprovalRuleOptions{
			{
				AccessLevel: gitlab.Ptr(gitlab.AccessLevelValue(rule.AccessLevel)),
			},
		}
	}

	_, _, err := p.client.ProtectedEnvironments.ProtectRepositoryEnvironments(encodedPath, opts, gitlab.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to protect environment %s: %w", envName, err)
	}

	return nil
}

// IsEnvironmentProtected checks if an environment is protected
func (p *GitLabProvider) IsEnvironmentProtected(ctx context.Context, projectPath, envName string) (bool, error) {
	encodedPath := url.PathEscape(projectPath)

	_, resp, err := p.client.ProtectedEnvironments.GetProtectedEnvironment(encodedPath, envName, gitlab.WithContext(ctx))
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check protection status for %s: %w", envName, err)
	}

	return true, nil
}
