# ztigit 0.0.1 Release Notes

**Release Date:** December 18, 2025

## Overview

ztigit is a cross-platform CLI for GitLab and GitHub. Mirror repositories, protect environments,
manage authentication - with security best practices built in.

## Features

### Mirror Command

Clone/update repos from GitLab groups or GitHub orgs/users:

- Auto-detect provider from URL: `ztigit mirror https://github.com/zsoftly`
- Default clone location: `$HOME/<org>/`
- HTTPS-first, SSH-fallback for git operations
- Skip stale repos: `--max-age 12` (default, 0 = no limit)
- Parallel processing (default: 4, configurable)
- Colored output with repo sizes
- GitLab: Groups including subgroups
- GitHub: Organizations and user accounts

### Preflight Credential Validation

Checks git credentials before starting clone operations:

- Tests HTTPS/SSH access with `git ls-remote`
- Fails fast with actionable fix suggestions
- Bypass with `--skip-preflight` if needed

### Other Commands

- **protect**: Protect deployment environments by pattern
- **environments**: List project environments with protection status
- **auth login**: Configure authentication tokens (env vars or stdin)
- **config**: Display current configuration

## Security

- **Keychain storage**: Tokens stored in system keychain (macOS Keychain, Linux secret-service,
  Windows Credential Manager)
- **No CLI token flag**: Tokens read from env vars or stdin only (prevents shell history exposure)
- **HTTPS enforcement**: Rejects HTTP URLs when token is present
- **Secure permissions**: Config directory 0700, config file 0600

## Platforms

| OS      | Architecture |
| ------- | ------------ |
| Linux   | amd64, arm64 |
| macOS   | amd64, arm64 |
| Windows | amd64        |

## Installation

```bash
# Linux (amd64)
curl -L https://github.com/zsoftly/ztigit/releases/download/0.0.1/ztigit-linux-amd64 -o ztigit
chmod +x ztigit && sudo mv ztigit /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/zsoftly/ztigit/releases/download/0.0.1/ztigit-darwin-arm64 -o ztigit
chmod +x ztigit && sudo mv ztigit /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/zsoftly/ztigit/releases/download/0.0.1/ztigit-darwin-amd64 -o ztigit
chmod +x ztigit && sudo mv ztigit /usr/local/bin/
```

## Quick Start

```bash
# Set token (env var)
export GITHUB_TOKEN=ghp_xxxx

# Mirror all repos from an org
ztigit mirror https://github.com/zsoftly

# Or save token to keychain
ztigit auth login -p github
```

**Full Changelog**: https://github.com/zsoftly/ztigit/commits/0.0.1
