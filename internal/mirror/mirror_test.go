package mirror

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zsoftly/ztigit/internal/provider"
)

// mockProvider is a mock implementation of the provider.Provider interface for testing.
type mockProvider struct {
	repos []provider.Repository
}

func (m *mockProvider) Name() string                                                              { return "mock" }
func (m *mockProvider) TestConnection(ctx context.Context) error                                  { return nil }
func (m *mockProvider) GetCurrentUser(ctx context.Context) (string, error)                        { return "mockuser", nil }
func (m *mockProvider) ListGroupProjects(ctx context.Context, groupPath string) ([]provider.Repository, error) {
	return m.repos, nil
}
func (m *mockProvider) ListGroups(ctx context.Context) ([]provider.Group, error)                 { return nil, nil }
func (m *mockProvider) GetProject(ctx context.Context, projectPath string) (*provider.Repository, error) {
	return nil, nil
}
func (m *mockProvider) ListEnvironments(ctx context.Context, projectPath string) ([]provider.Environment, error) {
	return nil, nil
}
func (m *mockProvider) ProtectEnvironment(ctx context.Context, projectPath, envName string, rule provider.ProtectionRule) error {
	return nil
}
func (m *mockProvider) IsEnvironmentProtected(ctx context.Context, projectPath, envName string) (bool, error) {
	return false, nil
}

func TestMirrorRepo_GitLabSubgroup(t *testing.T) {
	// 1. Setup
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// This repo mimics a GitLab project in a nested subgroup
	repo := provider.Repository{
		Name:     "my-project",
		FullPath: "my-group/my-subgroup/my-project",
		// Use a public, lightweight repo for the clone test
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{
		repos: []provider.Repository{repo},
	}

	opts := Options{
		BaseDir: tempDir,
		// Use a low number for tests
		Parallel: 1,
	}

	mirror := New(mockProvider, opts)

	// 2. Execute
	result := mirror.mirrorRepo(context.Background(), repo)

	// 3. Assert
	if result.Action != "cloned" {
		t.Errorf("Expected action 'cloned', but got '%s'", result.Action)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}

	// Verify the directory structure
	expectedRepoPath := filepath.Join(tempDir, repo.FullPath)
	if !isGitRepo(expectedRepoPath) {
		t.Errorf("Repository not found at expected path: %s", expectedRepoPath)
	}

	// Verify that the old, incorrect path does NOT exist
	incorrectRepoPath := filepath.Join(tempDir, repo.Name)
	if isGitRepo(incorrectRepoPath) {
		t.Errorf("Repository found at incorrect, flattened path: %s", incorrectRepoPath)
	}
}
