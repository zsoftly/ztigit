# Configuration

## Token Storage (Security)

Tokens are stored securely using the system keychain when available:

- **macOS**: Keychain Access
- **Linux**: secret-service (GNOME Keyring, KWallet)
- **Windows**: Credential Manager

If keychain is unavailable (e.g., headless servers), tokens fall back to config file with `0600`
permissions.

## Token Priority

Tokens are loaded in order (first found wins):

1. System keychain (most secure)
2. Environment variable (`GITLAB_TOKEN`, `GITHUB_TOKEN`)
3. Config file (`~/.config/ztigit/ztigit.yaml`)

## Environment Variables

| Variable       | Description                                     |
| -------------- | ----------------------------------------------- |
| `GITLAB_TOKEN` | GitLab personal access token                    |
| `GITLAB_URL`   | GitLab base URL (default: `https://gitlab.com`) |
| `GITHUB_TOKEN` | GitHub personal access token                    |
| `GITHUB_URL`   | GitHub base URL (default: `https://github.com`) |

## Config File

Location: `~/.config/ztigit/ztigit.yaml`

```yaml
default_provider: gitlab

gitlab:
  base_url: https://gitlab.com
  # token stored in system keychain (not in file)

github:
  base_url: https://github.com
  # token stored in system keychain (not in file)

mirror:
  base_dir: ~/git-repos
  parallel: 4
  skip_archived: true

debug: false
```

### Create Config via CLI

```bash
# GitLab (token read from GITLAB_TOKEN env var or stdin)
export GITLAB_TOKEN=glpat-xxxx
ztigit auth login -p gitlab

# GitLab self-hosted
ztigit auth login -p gitlab -u https://gitlab.example.com

# GitHub
export GITHUB_TOKEN=ghp_xxxx
ztigit auth login -p github

# Interactive (will prompt for token)
ztigit auth login -p gitlab
```

Config directory is created with `0700` permissions, config file with `0600` (owner access only).

## Token Scopes

### GitLab

Required scopes:

- `read_api` - list groups/projects
- `read_repository` - clone repositories
- `api` - protect environments (optional)

Create at: `Settings > Access Tokens > Personal Access Tokens`

### GitHub

Required scopes:

- `repo` - full repository access
- `read:org` - list organizations

Create at: `Settings > Developer settings > Personal access tokens`

## Provider Detection

When `--provider` is not specified:

1. If URL contains `gitlab` → GitLab
2. If URL contains `github` → GitHub
3. Default → GitLab

## Self-Hosted Instances

```bash
# GitLab self-hosted
export GITLAB_TOKEN=glpat-xxxx
ztigit auth login -p gitlab -u https://gitlab.company.com

# GitHub Enterprise
export GITHUB_TOKEN=ghp_xxxx
ztigit auth login -p github -u https://github.company.com
```

## View Current Config

```bash
ztigit config
```

Output shows configured providers with masked tokens.
