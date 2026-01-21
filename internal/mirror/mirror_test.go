package mirror

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/zsoftly/ztigit/internal/provider"
)

// mockProvider is a mock implementation of the provider.Provider interface for testing.
type mockProvider struct {
	repos []provider.Repository
}

func (m *mockProvider) Name() string                                       { return "mock" }
func (m *mockProvider) TestConnection(ctx context.Context) error           { return nil }
func (m *mockProvider) GetCurrentUser(ctx context.Context) (string, error) { return "mockuser", nil }
func (m *mockProvider) ListGroupProjects(ctx context.Context, groupPath string) ([]provider.Repository, error) {
	return m.repos, nil
}
func (m *mockProvider) ListGroups(ctx context.Context) ([]provider.Group, error) { return nil, nil }
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

func TestMirrorRepo_NestedPath(t *testing.T) {
	// 1. Setup
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// This repo mimics a nested subgroup structure (GitLab-style)
	// Note: Using GitHub fixture for testing since we're testing path handling, not provider-specific behavior
	repo := provider.Repository{
		Name:     "my-project",
		FullPath: "my-group/my-subgroup/my-project",
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{
		repos: []provider.Repository{repo},
	}

	opts := Options{
		BaseDir: tempDir,
		// Use Parallel=1 for deterministic test behavior and easier debugging
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

func TestValidatePath_WindowsReservedNames(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid path", "org/project", false},
		{"reserved CON", "org/CON/project", true},
		{"reserved PRN", "org/PRN", true},
		{"reserved AUX", "AUX/project", true},
		{"reserved NUL", "org/NUL.txt", true},
		{"reserved COM1", "org/COM1", true},
		{"reserved LPT1", "org/LPT1.log", true},
		{"trailing dot", "org/project.", true},
		{"trailing space", "org/project ", true},
		{"valid CON prefix", "org/CONFIG", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidateFullPathLength(t *testing.T) {
	// Test Windows MAX_PATH enforcement
	if runtime.GOOS == "windows" {
		// Path at limit (259 chars) should pass
		pathAtLimit := "C:\\" + strings.Repeat("a", 256)
		err := validateFullPathLength(pathAtLimit)
		if err != nil {
			t.Errorf("Expected no error for path at Windows MAX_PATH limit, got: %v", err)
		}

		// Path exceeding limit (260+ chars) should fail
		pathExceedingLimit := "C:\\" + strings.Repeat("a", 257)
		err = validateFullPathLength(pathExceedingLimit)
		if err == nil {
			t.Error("Expected error for path exceeding Windows MAX_PATH limit, got nil")
		}
	}

	// Test Unix path limits
	if runtime.GOOS != "windows" {
		// Very long path should eventually fail
		veryLongPath := "/" + strings.Repeat("a", 5000)
		err := validateFullPathLength(veryLongPath)
		if err == nil {
			t.Error("Expected error for extremely long path on Unix, got nil")
		}
	}
}

func TestMirrorRepo_DeepNesting(t *testing.T) {
	// Test 3+ level nesting
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := provider.Repository{
		Name:     "deep-project",
		FullPath: "org/group1/group2/group3/deep-project",
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{repos: []provider.Repository{repo}}
	opts := Options{BaseDir: tempDir, Parallel: 1}
	mirror := New(mockProvider, opts)

	result := mirror.mirrorRepo(context.Background(), repo)

	if result.Action != "cloned" {
		t.Errorf("Expected action 'cloned', but got '%s'", result.Action)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}

	// Verify deep directory structure
	expectedRepoPath := filepath.Join(tempDir, repo.FullPath)
	if !isGitRepo(expectedRepoPath) {
		t.Errorf("Repository not found at expected deep path: %s", expectedRepoPath)
	}
}

func TestMirrorRepo_RootLevel(t *testing.T) {
	// Test root-level repo (no subgroups)
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := provider.Repository{
		Name:     "simple-project",
		FullPath: "simple-project",
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{repos: []provider.Repository{repo}}
	opts := Options{BaseDir: tempDir, Parallel: 1}
	mirror := New(mockProvider, opts)

	result := mirror.mirrorRepo(context.Background(), repo)

	if result.Action != "cloned" {
		t.Errorf("Expected action 'cloned', but got '%s'", result.Action)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}

	// Verify root-level structure
	expectedRepoPath := filepath.Join(tempDir, repo.FullPath)
	if !isGitRepo(expectedRepoPath) {
		t.Errorf("Repository not found at expected root path: %s", expectedRepoPath)
	}
}

func TestMirrorRepo_SpecialCharacters(t *testing.T) {
	// Test path with hyphens, underscores, dots (all valid)
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := provider.Repository{
		Name:     "my-project_v2.0",
		FullPath: "my-org/my-group_test/my-project_v2.0",
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{repos: []provider.Repository{repo}}
	opts := Options{BaseDir: tempDir, Parallel: 1}
	mirror := New(mockProvider, opts)

	result := mirror.mirrorRepo(context.Background(), repo)

	if result.Action != "cloned" {
		t.Errorf("Expected action 'cloned', but got '%s'", result.Action)
	}
	if result.Error != nil {
		t.Errorf("Expected no error, but got: %v", result.Error)
	}

	expectedRepoPath := filepath.Join(tempDir, repo.FullPath)
	if !isGitRepo(expectedRepoPath) {
		t.Errorf("Repository not found at expected path: %s", expectedRepoPath)
	}
}

func TestValidatePath_InvalidCharacters(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid simple path", "org/project", false},
		{"valid nested path", "org/group1/group2/project", false},
		{"valid with hyphens", "my-org/my-project", false},
		{"valid with underscores", "my_org/my_project", false},
		{"valid with dots", "my.org/my.project", false},
		{"empty path", "", true},
		{"path traversal", "org/../other", true},
		{"null character", "org/\x00project", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath_LongPaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test 1: Path at conservative relative limit should pass initial validation
	longPath := "org/" + strings.Repeat("a", 195) // Total ~199 chars
	repo := provider.Repository{
		Name:     "project",
		FullPath: longPath,
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{repos: []provider.Repository{repo}}
	opts := Options{BaseDir: tempDir, Parallel: 1}
	mirror := New(mockProvider, opts)

	result := mirror.mirrorRepo(context.Background(), repo)

	// Verify the clone either succeeded or failed with a clear error
	if result.Action != "cloned" && result.Action != "failed" {
		t.Errorf("Expected action 'cloned' or 'failed', got '%s'", result.Action)
	}

	// If it failed, it should be due to path length (on Windows) or git error, not a panic
	if result.Action == "failed" && result.Error == nil {
		t.Error("Failed action should have an error message")
	}

	// Test 2: Extremely long relative path should fail validation
	// Use a length that exceeds limits on both Windows (199) and Unix (3000)
	var veryLongPath string
	if runtime.GOOS == "windows" {
		veryLongPath = "org/" + strings.Repeat("x", 300) // Exceeds 199
	} else {
		veryLongPath = "org/" + strings.Repeat("x", 3500) // Exceeds 3000
	}
	err = validatePath(veryLongPath)
	if err == nil {
		t.Error("Expected error for extremely long relative path, got nil")
	}

	// Test 3: Validate full path length checking
	absPath := filepath.Join(tempDir, longPath)
	err = validateFullPathLength(absPath)
	// Error depends on OS and actual path length
	if runtime.GOOS == "windows" && len(absPath) > 259 {
		if err == nil {
			t.Errorf("Expected error for Windows path exceeding MAX_PATH, got nil")
		}
	}
}

func TestMirrorRepo_InvalidPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ztigit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	repo := provider.Repository{
		Name:     "bad-project",
		FullPath: "org/../../../etc/passwd", // Path traversal attempt
		CloneURL: "https://github.com/git-fixtures/basic.git",
		SSHUrl:   "git@github.com:git-fixtures/basic.git",
	}

	mockProvider := &mockProvider{repos: []provider.Repository{repo}}
	opts := Options{BaseDir: tempDir, Parallel: 1}
	mirror := New(mockProvider, opts)

	result := mirror.mirrorRepo(context.Background(), repo)

	// Should fail due to invalid path
	if result.Action != "failed" {
		t.Errorf("Expected action 'failed', but got '%s'", result.Action)
	}
	if result.Error == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}
