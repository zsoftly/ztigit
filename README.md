# ztigit

Cross-platform CLI for GitLab and GitHub. Mirror repositories, protect environments, manage
authentication.

## Install

**Windows:**

```powershell
irm https://github.com/zsoftly/ztigit/releases/latest/download/install.ps1 | iex
```

**macOS / Linux / WSL:**

```bash
curl -fsSL https://github.com/zsoftly/ztigit/releases/latest/download/install.sh | bash
```

## Quick Start

```bash
# Set token
export GITHUB_TOKEN=ghp_xxxx      # Unix
$env:GITHUB_TOKEN = "ghp_xxxx"    # PowerShell

# Mirror repositories
ztigit mirror https://github.com/zsoftly
ztigit mirror zsoftly --provider github --dir ./repos
```

## Commands

| Command        | Description                           |
| -------------- | ------------------------------------- |
| `mirror`       | Clone/update repositories from groups |
| `auth login`   | Save authentication token             |
| `config`       | Show current configuration            |
| `environments` | List project environments             |
| `protect`      | Protect environments matching pattern |

## Examples

See platform-specific examples:

- **[Unix Examples](examples/unix.md)** - macOS, Linux, WSL
- **[Windows Examples](examples/windows.md)** - PowerShell, Command Prompt

## Configuration

Tokens are loaded from (in order):

1. **System keychain** - macOS Keychain, Linux libsecret, Windows Credential Manager
2. **Environment variables** - `GITHUB_TOKEN`, `GITLAB_TOKEN`
3. **Config file** - `~/.config/ztigit/ztigit.yaml`

```bash
# Save token to keychain
ztigit auth login -p github
```

## Documentation

- [Installation](docs/installation.md)
- [Configuration](docs/configuration.md)
- [Commands Reference](docs/commands.md)

## License

MIT
