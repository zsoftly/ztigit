# Commands Reference

## auth

Manage authentication tokens.

### auth login

Save authentication token for a provider.

Token is read from environment variable or stdin (never command line for security).

```bash
ztigit auth login --provider <gitlab|github> [--url <base_url>]
```

| Flag               | Required | Description                         |
| ------------------ | -------- | ----------------------------------- |
| `--provider`, `-p` | Yes      | Provider: `gitlab` or `github`      |
| `--url`, `-u`      | No       | Base URL (default: public instance) |

Examples:

```bash
# Using environment variable (recommended)
export GITLAB_TOKEN=glpat-xxxx
ztigit auth login -p gitlab

# Using stdin (pipe)
echo $GITHUB_TOKEN | ztigit auth login -p github

# Interactive prompt
ztigit auth login -p gitlab

# Self-hosted GitLab
export GITLAB_TOKEN=glpat-xxxx
ztigit auth login -p gitlab -u https://gitlab.company.com
```

**Security:** Tokens are stored in the system keychain (macOS Keychain, Linux secret-service,
Windows Credential Manager) when available, otherwise in config file with restricted permissions.

---

## config

Show current configuration.

```bash
ztigit config
```

Displays:

- Config file location
- GitLab URL and token (masked)
- GitHub URL and token (masked)
- Mirror settings

---

## mirror

Clone or update repositories from groups/organizations.

```bash
ztigit mirror <url-or-org> [options]
```

| Flag               | Required | Description                                                    |
| ------------------ | -------- | -------------------------------------------------------------- |
| `<url-or-org>`     | Yes      | URL or org/group name (positional)                             |
| `--provider`, `-p` | No       | Provider (required if not using URL)                           |
| `--dir`, `-d`      | No       | Base directory (default: `$HOME/<org>`)                        |
| `--max-age`        | No       | Skip repos not updated in N months (default: 12, 0 = no limit) |
| `--parallel`       | No       | Parallel operations (default: 4)                               |
| `--prefer-ssh`     | No       | Use SSH URLs instead of HTTPS for git operations               |
| `--skip-preflight` | No       | Skip git credential validation before cloning                  |
| `--verbose`, `-v`  | No       | Verbose output                                                 |

**Authentication:**

- API: Set `GITHUB_TOKEN` or `GITLAB_TOKEN` environment variable
- Git: Uses your existing git credentials (HTTPS credential helper or SSH keys)
- Default: HTTPS first, falls back to SSH if HTTPS fails
- Use `--prefer-ssh` to try SSH first (recommended if you have SSH keys configured)

**GitLab**: Groups including subgroups are supported.

**GitHub**: Both organizations and user accounts are supported.

Examples:

```bash
# Auto-detect from URL
ztigit mirror https://github.com/zsoftly
ztigit mirror https://gitlab.com/devops

# Specify provider manually
ztigit mirror zsoftly --provider github

# Include older repos (default skips repos not updated in 12 months)
ztigit mirror zsoftly -p github --max-age 24

# No age limit (clone all repos)
ztigit mirror zsoftly -p github --max-age 0

# Custom directory, verbose
ztigit mirror https://github.com/zsoftly -d ~/projects -v

# Use SSH instead of HTTPS
ztigit mirror https://github.com/zsoftly --prefer-ssh
```

Output:

```
Connecting to https://github.com...
Authenticated as: ditahk

Mirroring zsoftly to /home/ditahk/zsoftly...

[OK] Cloned: ztigit (1.2s)
[OK] Updated: ztiaws (0.8s)
[SKIP] Archived: old-repo
[SKIP] Stale: legacy-tool (last updated: 2023-01-15)

Summary:
  Cloned:  1
  Updated: 1
  Skipped: 1 (archived)
  Stale:   1 (not updated recently)
  Failed:  0
  Total:   4
```

---

## environments

List deployment environments for a project.

```bash
ztigit environments --project <path> [options]
```

| Flag               | Required | Description                       |
| ------------------ | -------- | --------------------------------- |
| `--project`, `-P`  | Yes      | Project path (e.g., `group/repo`) |
| `--provider`, `-p` | No       | Provider (auto-detected)          |
| `--url`, `-u`      | No       | Base URL                          |

Examples:

```bash
# GitLab project
ztigit environments -P "devops/deploy-tools"

# GitHub repo
ztigit environments -P "zsoftly/ztiaws" -p github
```

Output:

```
Environments:

  dev                                      [unprotected] available
  staging                                  [protected] available
  production                               [protected] available

Total: 3 environments
```

---

## protect

Protect environments matching a pattern.

```bash
ztigit protect --project <path> --pattern <pattern> [options]
```

| Flag               | Required | Description                                 |
| ------------------ | -------- | ------------------------------------------- |
| `--project`, `-P`  | Yes      | Project path                                |
| `--pattern`        | Yes      | Environment name pattern (prefix or `all`)  |
| `--provider`, `-p` | No       | Provider (required if `--url` not set)      |
| `--url`, `-u`      | No       | Base URL (required if `--provider` not set) |
| `--dry-run`        | No       | Show what would be protected                |
| `--access-level`   | No       | Required access level (default: 30)         |
| `--approvals`      | No       | Required approvals (default: 1)             |

**Note:** At least one of `--provider` or `--url` must be specified.

**GitHub Limitation:** The `--access-level` and `--approvals` flags only work with GitLab. GitHub
environment protection requires team or user IDs for reviewers, which this tool does not currently
support. For GitHub, environments will be created but protection rules must be configured via the
GitHub UI or API directly.

Access levels (GitLab only):

- `30` - Developer
- `40` - Maintainer
- `60` - Admin

Examples:

```bash
# Protect all environments starting with "prod"
ztigit protect -P "devops/deploy-tools" --pattern "prod"

# Protect all environments
ztigit protect -P "zsoftly/ztiaws" -p github --pattern "all"

# Dry run
ztigit protect -P "devops/deploy-tools" --pattern "dev" --dry-run

# Require maintainer access
ztigit protect -P "devops/deploy-tools" --pattern "prod" --access-level 40
```

Output:

```
[OK] Protected: prod-us-east-1
[OK] Protected: prod-eu-west-1
[SKIP] Already protected: prod-main

Summary:
  Protected: 2
  Skipped:   1 (already protected)
  Failed:    0
  Total:     3
```

---

## Global Flags

Available on all commands:

| Flag              | Description  |
| ----------------- | ------------ |
| `--help`, `-h`    | Show help    |
| `--version`, `-v` | Show version |
