# ztigit Examples - Windows

## Installation

**PowerShell:**

```powershell
irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
```

## Setup

**PowerShell:**

```powershell
# Set token for API access (current session)
$env:GITHUB_TOKEN = "ghp_xxxx"
$env:GITLAB_TOKEN = "glpat-xxxx"

# Or save to Windows Credential Manager (persistent)
ztigit auth login -p github
ztigit auth login -p gitlab
```

**Command Prompt:**

```cmd
:: Set token for API access
set GITHUB_TOKEN=ghp_xxxx
set GITLAB_TOKEN=glpat-xxxx
```

## Mirror Repositories

**PowerShell:**

```powershell
# From URL (auto-detects provider)
ztigit mirror https://github.com/zsoftly
ztigit mirror https://gitlab.com/my-group

# By org name
ztigit mirror zsoftly --provider github
ztigit mirror my-group --provider gitlab

# Custom directory
ztigit mirror zsoftly -p github --dir C:\repos
ztigit mirror zsoftly -p github --dir $env:USERPROFILE\projects

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

**Command Prompt:**

```cmd
:: Mirror from URL
ztigit mirror https://github.com/zsoftly

:: Custom directory
ztigit mirror zsoftly -p github --dir C:\repos
```

## Authentication

**PowerShell:**

```powershell
# Save token to Windows Credential Manager
$env:GITHUB_TOKEN = "ghp_xxxx"
ztigit auth login -p github

# Pipe token
"ghp_xxxx" | ztigit auth login -p github

# From file (less secure - use Credential Manager when possible)
Get-Content ~/.github_token | ztigit auth login -p github

# GitLab self-hosted
$env:GITLAB_TOKEN = "glpat-xxxx"
ztigit auth login -p gitlab -u https://gitlab.mycompany.com

# View current config
ztigit config
```

## Environment Protection

```powershell
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

### Daily Backup (Task Scheduler)

**Recommended:** Use `ztigit auth login` first to store token in Windows Credential Manager. The
token will be loaded automatically - no environment variable needed.

Create `backup-repos.ps1`:

```powershell
# Save as C:\Scripts\backup-repos.ps1
# Create Task Scheduler task to run daily
# Prerequisites: run 'ztigit auth login -p github' once to store token

ztigit mirror myorg -p github --dir C:\backups\github --max-age 0
```

Task Scheduler command:

```
powershell.exe -ExecutionPolicy Bypass -File C:\Scripts\backup-repos.ps1
```

**Alternative:** If Credential Manager is unavailable, use environment variable. Ensure the token
file has appropriate permissions (not world-readable).

```powershell
$env:GITHUB_TOKEN = Get-Content "$env:USERPROFILE\.github_token"
ztigit mirror myorg -p github --dir C:\backups\github --max-age 0
```

### Mirror Multiple Orgs

```powershell
# Token loaded from Credential Manager (run 'ztigit auth login -p github' first)

@("org1", "org2", "org3") | ForEach-Object {
    Write-Host "Mirroring $_..."
    ztigit mirror $_ -p github --dir C:\repos\$_
}
```

### Azure DevOps Pipeline

```yaml
- task: PowerShell@2
  displayName: 'Mirror repositories'
  env:
    GITHUB_TOKEN: $(GITHUB_TOKEN)
  inputs:
    targetType: 'inline'
    script: |
      irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
      ztigit mirror myorg -p github --dir $(Build.SourcesDirectory)/repos
```

### GitHub Actions (Windows Runner)

```yaml
- name: Mirror repositories
  shell: pwsh
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
    ztigit mirror myorg -p github --dir ./repos
```
