# ztigit Examples - Unix (macOS, Linux, WSL)

## Installation

```bash
curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
```

## Setup

```bash
# Set token for API access
export GITHUB_TOKEN=ghp_xxxx
export GITLAB_TOKEN=glpat-xxxx

# Or save to keychain (persistent)
ztigit auth login -p github
ztigit auth login -p gitlab
```

## Mirror Repositories

```bash
# From URL (auto-detects provider)
ztigit mirror https://github.com/zsoftly
ztigit mirror https://gitlab.com/my-group

# By org name
ztigit mirror zsoftly --provider github
ztigit mirror my-group --provider gitlab

# Custom directory
ztigit mirror zsoftly -p github --dir ~/projects

# Use SSH instead of HTTPS
ztigit mirror zsoftly -p github --ssh

# Include older repos (default skips repos >12 months old)
ztigit mirror zsoftly -p github --max-age 24   # 24 months
ztigit mirror zsoftly -p github --max-age 0    # no limit

# Parallel operations
ztigit mirror zsoftly -p github --parallel 8

# Verbose output
ztigit mirror zsoftly -p github --verbose

# Skip credential check
ztigit mirror zsoftly -p github --skip-preflight
```

## Authentication

```bash
# Save token to system keychain
export GITHUB_TOKEN=ghp_xxxx
ztigit auth login -p github

# Pipe token (useful in scripts)
echo $GITHUB_TOKEN | ztigit auth login -p github

# GitLab self-hosted
export GITLAB_TOKEN=glpat-xxxx
ztigit auth login -p gitlab -u https://gitlab.mycompany.com

# View current config
ztigit config
```

## Environment Protection

```bash
# List environments
ztigit environments --project "org/repo" --provider github
ztigit environments -P "org/repo" -p gitlab

# Protect environments matching pattern
ztigit protect -P "org/repo" -p github --pattern "prod"
ztigit protect -P "org/repo" -p gitlab --pattern "staging"

# Preview changes (dry run)
ztigit protect -P "org/repo" -p github --pattern "prod" --dry-run

# Protect all environments
ztigit protect -P "org/repo" -p github --pattern "all"
```

## Scripting Examples

### Daily Backup (cron)

**Recommended:** Use `ztigit auth login` first to store token in system keychain. The token
will be loaded automatically - no environment variable needed.

```bash
#!/bin/bash
# Save as ~/scripts/backup-repos.sh
# Add to crontab: 0 2 * * * ~/scripts/backup-repos.sh
# Prerequisites: run 'ztigit auth login -p github' once to store token in keychain

ztigit mirror myorg -p github --dir ~/backups/github --max-age 0
```

**Alternative:** If keychain is unavailable (headless servers), use environment variable
with restrictive file permissions:

```bash
#!/bin/bash
# Token file must have 600 permissions: chmod 600 ~/.github_token
export GITHUB_TOKEN=$(cat ~/.github_token)
ztigit mirror myorg -p github --dir ~/backups/github --max-age 0
```

### Mirror Multiple Orgs

```bash
#!/bin/bash
# Token loaded from keychain (run 'ztigit auth login -p github' first)

for org in org1 org2 org3; do
    echo "Mirroring $org..."
    ztigit mirror $org -p github --dir ~/repos/$org
done
```

### GitHub Actions

```yaml
- name: Mirror repositories
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: ztigit mirror myorg -p github --dir ./repos
```

### GitLab CI

```yaml
mirror:
  script:
    - curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
    - ztigit mirror mygroup -p gitlab --dir ./repos
  variables:
    GITLAB_TOKEN: $CI_JOB_TOKEN
```
