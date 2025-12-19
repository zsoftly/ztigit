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
  - Parallel processing (configurable)
  - Skip archived repositories
  - Skip stale repos (not updated in N months, default: 12)
  - Subgroup support (GitLab)
  - Organization and user repos (GitHub)
- **protect command**: Protect deployment environments by pattern
- **environments command**: List project environments with protection status
- **auth login command**: Configure authentication tokens
- **config command**: Display current configuration
- **Cross-platform builds**: Linux, macOS, Windows (amd64, arm64)
- **Configuration**: Environment variables, config file, CLI flags

### Technical

- Go 1.24 with cobra CLI framework
- Provider abstraction pattern for GitLab/GitHub
- System git binary for clone/pull operations
- GitHub Actions CI/CD pipeline

**Full Changelog**: https://github.com/zsoftly/ztigit/commits/0.0.1
