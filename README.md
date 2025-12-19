# ztigit

Cross-platform CLI for GitLab and GitHub. Mirror repositories, protect environments, manage
authentication.

## Why

- **One tool** for both GitLab and GitHub
- **Cross-platform** - Linux, macOS, Windows
- **Single binary** - no runtime dependencies
- **Parallel operations** - fast repository mirroring

## Install

### Download Binary

```bash
# Linux (amd64)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-linux-amd64 -o ztigit
chmod +x ztigit
sudo mv ztigit /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-darwin-arm64 -o ztigit
chmod +x ztigit
sudo mv ztigit /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-windows-amd64.exe -OutFile ztigit.exe
```

### Build from Source

```bash
git clone https://github.com/zsoftly/ztigit.git
cd ztigit
go build -o ztigit ./cmd/ztigit
```

## Quick Start

```bash
# Set token (API access for listing repos)
export GITHUB_TOKEN=ghp_xxxx
export GITLAB_TOKEN=glpat-xxxx

# Mirror repositories (auto-detect from URL)
ztigit mirror https://github.com/zsoftly

# Or specify provider manually
ztigit mirror zsoftly --provider github

# List environments
ztigit environments --project "zsoftly/ztiaws" --provider github

# Protect environments
ztigit protect --project "zsoftly/ztiaws" --provider github --pattern "prod"
```

## Commands

| Command               | Description                           |
| --------------------- | ------------------------------------- |
| `ztigit auth login`   | Save authentication token             |
| `ztigit config`       | Show current configuration            |
| `ztigit mirror`       | Clone/update repositories from groups |
| `ztigit environments` | List project environments             |
| `ztigit protect`      | Protect environments matching pattern |

## Configuration

Tokens are loaded from (in order):

1. **System keychain** (most secure) - macOS Keychain, Linux secret-service, Windows Credential
   Manager
2. **Environment variables**: `GITLAB_TOKEN`, `GITHUB_TOKEN`
3. **Config file**: `~/.config/ztigit/ztigit.yaml`

```bash
# Save token to keychain
ztigit auth login -p github   # reads from GITHUB_TOKEN env var or prompts
```

See [docs/configuration.md](docs/configuration.md) for details.

## Documentation

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Commands Reference](docs/commands.md)

## Requirements

- Git (for clone/pull operations)
- GitLab/GitHub personal access token with appropriate scopes

## License

MIT
