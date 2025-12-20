# ztigit v0.0.3 Release Notes

## What's New

### One-Liner Installation

Install ztigit with a single command - no manual downloads or PATH configuration needed.

**Windows (PowerShell):**
```powershell
irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
```

**macOS, Linux, WSL:**
```bash
curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
```

### SSH Support

New `--ssh` flag to prefer SSH URLs over HTTPS:
```bash
ztigit mirror zsoftly --provider github --ssh
```

### Windows arm64 Support

Added Windows ARM64 binary for Surface Pro X and other ARM-based Windows devices.

---

## Features

### Mirror Command

Clone/update repos from GitLab groups or GitHub orgs:

```bash
# From URL (auto-detects provider)
ztigit mirror https://github.com/zsoftly
ztigit mirror https://gitlab.com/my-group

# By org name
ztigit mirror zsoftly --provider github
ztigit mirror my-group --provider gitlab

# Custom directory
ztigit mirror zsoftly -p github --dir ./repos

# Use SSH instead of HTTPS
ztigit mirror zsoftly -p github --ssh

# Include older repos (default skips repos >12 months old)
ztigit mirror zsoftly -p github --max-age 24   # 24 months
ztigit mirror zsoftly -p github --max-age 0    # no limit
```

### Preflight Credential Validation

Checks git credentials before starting:
- Tests HTTPS/SSH access with `git ls-remote`
- Fails fast with actionable fix suggestions
- Bypass with `--skip-preflight` if needed

### Environment Protection

```bash
# List environments
ztigit environments --project "org/repo" --provider github

# Protect environments matching pattern
ztigit protect --project "org/repo" --provider github --pattern "prod"
ztigit protect --project "org/repo" --provider gitlab --pattern "staging"

# Dry run
ztigit protect --project "org/repo" -p github --pattern "prod" --dry-run
```

### Authentication

```bash
# Save token to keychain (reads from env var or prompts)
export GITHUB_TOKEN=ghp_xxxx
ztigit auth login -p github

# Or pipe token
echo $GITHUB_TOKEN | ztigit auth login -p github

# View config
ztigit config
```

## Security

- **Keychain storage**: Tokens stored in system keychain (macOS Keychain, Linux libsecret, Windows Credential Manager)
- **No CLI token flag**: Tokens read from env vars or stdin only (prevents shell history exposure)
- **HTTPS enforcement**: Rejects HTTP URLs when token is present
- **Secure permissions**: Config directory 0700, config file 0600

## Platforms

| OS      | Architecture | Binary                       |
| ------- | ------------ | ---------------------------- |
| Linux   | amd64        | `ztigit-linux-amd64`         |
| Linux   | arm64        | `ztigit-linux-arm64`         |
| macOS   | amd64        | `ztigit-darwin-amd64`        |
| macOS   | arm64        | `ztigit-darwin-arm64`        |
| Windows | amd64        | `ztigit-windows-amd64.exe`   |
| Windows | arm64        | `ztigit-windows-arm64.exe`   |

## Installation

### Quick Install (Recommended)

**Windows:**
```powershell
irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
```

**macOS/Linux/WSL:**
```bash
curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
```

### Manual Install

See [docs/installation.md](docs/installation.md) for manual installation options.

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.2...0.0.3
