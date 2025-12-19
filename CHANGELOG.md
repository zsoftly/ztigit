# Changelog

All notable changes to ztigit will be documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), using
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.1] - 2025-12-18

### Added

- **Multi-platform support**: GitLab and GitHub from single CLI
- **mirror command**: Clone/update repositories from groups/orgs
  - Auto-detect provider from URL (e.g., `https://github.com/zsoftly`)
  - Clone to `$HOME/<org>/` by default
  - HTTPS-first, SSH-fallback for git operations
  - Parallel processing (configurable, default: 4)
  - Skip archived repositories
  - Skip stale repos (not updated in N months, default: 12)
  - Subgroup support (GitLab)
  - Organization and user repos (GitHub)
  - Display repo sizes during operations
  - Colored output (git-style)
- **Preflight credential validation**: Checks git credentials before cloning
  - Tests HTTPS/SSH access with `git ls-remote`
  - Fails fast with actionable fix suggestions
  - Bypass with `--skip-preflight` flag
- **protect command**: Protect deployment environments by pattern
- **environments command**: List project environments with protection status
- **auth login command**: Configure authentication tokens
  - Tokens read from env vars or stdin (never CLI flags)
  - Interactive prompt when no token provided
- **config command**: Display current configuration
- **Cross-platform builds**: Linux, macOS, Windows (amd64, arm64)
- **Configuration**: Environment variables, config file, CLI flags

### Security

- **Keychain storage**: Tokens stored in system keychain when available
  - macOS Keychain, Linux secret-service, Windows Credential Manager
  - Falls back to config file with 0600 permissions
- **No token in CLI flags**: Prevents exposure in shell history
- **HTTPS enforcement**: Rejects HTTP URLs when token is present
- **Secure config directory**: Created with 0700 permissions
- **Token display**: Shows `***configured***` instead of partial token

### Known Limitations

- **GitHub environment protection**: The `protect` command creates environments on GitHub but cannot
  configure protection rules (reviewers, wait timers). GitHub's API requires team or user IDs for
  reviewers, which this tool does not currently support. Configure protection rules via GitHub UI.

### Technical

- Go 1.24 with cobra CLI framework
- Provider abstraction pattern for GitLab/GitHub
- System git binary for clone/pull operations
- GitHub Actions CI/CD pipeline
- go-keyring for cross-platform keychain support

**Full Changelog**: https://github.com/zsoftly/ztigit/commits/0.0.1
