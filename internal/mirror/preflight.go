package mirror

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/zsoftly/ztigit/internal/provider"
)

// PreflightResult contains the result of credential preflight check
type PreflightResult struct {
	HTTPSWorks bool
	SSHWorks   bool
	Method     string // "https", "ssh", or ""
	Error      error
}

// credentialMethod represents a git credential method to test
type credentialMethod struct {
	name string // "ssh" or "https"
	url  string
}

// Preflight checks git credentials before starting clone operations
// Returns the preferred method (https or ssh) that works, or an error if neither works
func (m *Mirror) Preflight(ctx context.Context, repos []provider.Repository) (*PreflightResult, error) {
	if len(repos) == 0 {
		if m.options.SSH {
			return &PreflightResult{Method: "ssh"}, nil
		}
		return &PreflightResult{Method: "https"}, nil
	}

	// Use first repo for testing
	testRepo := repos[0]

	// Build ordered list of methods to test
	var methods []credentialMethod
	if m.options.SSH {
		methods = []credentialMethod{
			{name: "ssh", url: testRepo.SSHUrl},
			{name: "https", url: testRepo.CloneURL},
		}
	} else {
		methods = []credentialMethod{
			{name: "https", url: testRepo.CloneURL},
			{name: "ssh", url: testRepo.SSHUrl},
		}
	}

	// Test each method in order
	result := &PreflightResult{}
	for _, method := range methods {
		if method.url == "" {
			continue
		}
		fmt.Printf("  %s Testing %s credentials...\n", cyan("→"), strings.ToUpper(method.name))
		if m.testCredentials(ctx, method.url) {
			result.Method = method.name
			if method.name == "ssh" {
				result.SSHWorks = true
			} else {
				result.HTTPSWorks = true
			}
			return result, nil
		}
	}

	// Neither works - build helpful error message
	host := extractHost(testRepo.CloneURL)
	providerName := detectProviderName(host)

	errMsg := fmt.Sprintf(`%s Git credentials not configured

  Neither HTTPS nor SSH authentication is working for %s.

  To fix, try one of:

    %s Configure SSH (recommended):
      1. Generate key:  ssh-keygen -t ed25519
      2. Add to agent: ssh-add ~/.ssh/id_ed25519
      3. Copy public key to %s

    %s Configure HTTPS with token:
      git config --global url."https://oauth2:$%s_TOKEN@%s/".insteadOf "https://%s/"

    %s Configure credential helper:
      git config --global credential.helper store
      git clone %s  # Enter credentials once
`,
		red("✗"),
		host,
		bold("•"),
		providerName,
		bold("•"),
		strings.ToUpper(providerName),
		host,
		host,
		bold("•"),
		testRepo.CloneURL,
	)

	return nil, errors.New(errMsg)
}

// testCredentials tests if git can access a URL using ls-remote
func (m *Mirror) testCredentials(ctx context.Context, url string) bool {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Use git ls-remote to test credentials without cloning
	// --exit-code returns non-zero if no refs found (but auth succeeded)
	// We just care about whether auth works, not if refs exist
	cmd := exec.CommandContext(timeoutCtx, "git", "ls-remote", "--quiet", url)

	// Suppress output
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Set GIT_TERMINAL_PROMPT=0 to prevent git from prompting for credentials
	cmd.Env = append(cmd.Environ(), "GIT_TERMINAL_PROMPT=0")

	err := cmd.Run()
	return err == nil
}

// extractHost extracts the host from a git URL
func extractHost(gitURL string) string {
	if strings.HasPrefix(gitURL, "git@") {
		// SSH format: git@github.com:org/repo.git
		parts := strings.SplitN(gitURL, ":", 2)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "git@")
		}
	}

	// HTTPS format: https://github.com/org/repo.git
	parsed, err := url.Parse(gitURL)
	if err == nil && parsed.Host != "" {
		return parsed.Host
	}

	return "the remote server"
}

// detectProviderName returns the provider name for help messages
func detectProviderName(host string) string {
	switch {
	case strings.Contains(host, "github"):
		return "GitHub"
	case strings.Contains(host, "gitlab"):
		return "GitLab"
	case strings.Contains(host, "bitbucket"):
		return "Bitbucket"
	default:
		return host
	}
}
