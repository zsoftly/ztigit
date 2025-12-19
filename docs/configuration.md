# Configuration

## Token Priority

Tokens are loaded in order (first found wins):

1. CLI flag (`--token`)
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
  token: glpat-xxxxxxxxxxxx
  base_url: https://gitlab.com

github:
  token: ghp_xxxxxxxxxxxx
  base_url: https://github.com

mirror:
  base_dir: ~/git-repos
  parallel: 4
  skip_archived: true

debug: false
```

### Create Config via CLI

```bash
# GitLab
ztigit auth login --provider gitlab --token glpat-xxxx --url https://gitlab.example.com

# GitHub
ztigit auth login --provider github --token ghp_xxxx
```

Config file is created with `0600` permissions (owner read/write only).

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
ztigit auth login --provider gitlab --token glpat-xxxx --url https://gitlab.company.com

# GitHub Enterprise
ztigit auth login --provider github --token ghp_xxxx --url https://github.company.com
```

## View Current Config

```bash
ztigit config
```

Output shows configured providers with masked tokens.
