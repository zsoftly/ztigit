# Changelog

All notable changes to ztigit will be documented in this file.

Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), using
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.5] - 2026-01-23

### Added

- **Mirror multiple groups**: Mirror more than one GitLab group/GitHub org in a single command
  - Comma-separated: `ztigit mirror group1,group2,group3 -p gitlab`
  - Space-separated: `ztigit mirror --groups "group1 group2 group3" -p gitlab`
- **GitLab subgroup nesting preserved**: Full namespace hierarchy is kept in the local directory
  structure (e.g., `group/subgroup/project`)

### Changed

- **Default mirror directory for multiple groups**: Uses provider-specific directory
  (`$HOME/gitlab-repos` / `$HOME/github-repos`) instead of `$HOME/<group>`
- **Path validation**: Validate repository paths before cloning/updating and guard against overly
  long paths (notably on Windows)

### Fixed

- **Mirror input handling**: Improved error handling for empty/invalid group inputs

### Tests

- **Mirror test coverage**: Added comprehensive mirror tests to reduce regressions

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.4...0.0.5

---

## [0.0.4] - 2025-12-20

### Added

- **Git installation check**: Validates git is installed before running mirror command
  - Fails fast with platform-specific installation instructions
  - Windows: winget, Chocolatey, or manual download
  - macOS: xcode-select, Homebrew, or manual download
  - Linux: apt, dnf, pacman, or apk depending on distro

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.3...0.0.4

---

## [0.0.3] - 2025-12-20

### Added

- **One-liner installers**: Install with single command
  - Windows: `irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex`
  - Unix: `curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash`
- **Windows arm64 support**: Added binary for ARM-based Windows devices
- **Platform examples**: Separate example files for Unix and Windows
  - `examples/unix.md` - macOS, Linux, WSL examples
  - `examples/windows.md` - PowerShell, Command Prompt examples
- **Script linting CI**: Shellcheck and PSScriptAnalyzer run on every push

### Changed

- **Efficient CI pipeline**: Path-based job triggering
  - Go code changes → test + security jobs only
  - Script changes → lint-scripts job only
  - Docs/examples changes → no CI (skipped)
- **Simplified README**: Concise, references examples folder

### Fixed

- **install.sh**: Quote `$TMP_FILE` in trap for paths with spaces
- **install.ps1**: Reject 32-bit Windows (not supported)
- **Security docs**: Recommend keychain over plain text token files

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.2...0.0.3

---

## [0.0.2] - 2025-12-19

### Changed

- **Auto-detect git credentials**: Preflight automatically switches to SSH if HTTPS is unavailable
- **Renamed `--prefer-ssh` to `--ssh`**: Simpler flag name to skip HTTPS test and use SSH directly

### Fixed

- **Installation docs**: Added macOS Intel (darwin-amd64) to download instructions
- **Clarified docs**: Updated authentication documentation to accurately describe auto-detection
  behavior

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.1...0.0.2

---

## [0.0.1] - 2025-12-18

### Added

- **Multi-platform support**: GitLab and GitHub from single CLI
- **mirror command**: Clone/update repositories from groups/orgs
  - Auto-detect provider from URL (e.g., `https://github.com/zsoftly`)
  - Clone to `$HOME/<org>/` by default
  - Auto-detects working git credentials (HTTPS or SSH), or use `--ssh` to force SSH
  - Parallel processing (configurable, default: 4)
  - Skip archived repositories
  - Skip stale repos (not updated in N months, default: 12)
  - Subgroup support (GitLab)
  - Organization and user repos (GitHub)
  - Display repo sizes during operations
  - Colored output (git-style)
  - Smart update: stashes changes, switches to default branch, pulls latest
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
