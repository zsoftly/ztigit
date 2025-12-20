# ztigit v0.0.4 Release Notes

## What's New

### Git Installation Check

ztigit now validates that git is installed before running the mirror command. If git is not found, it provides platform-specific installation instructions:

**Windows:**
```
✗ Git is not installed

  Install git using one of:

    • winget (recommended):
      winget install Git.Git

    • Chocolatey:
      choco install git

    • Manual download:
      https://git-scm.com/download/win

  After installing, restart your terminal.
```

**macOS:**
```
✗ Git is not installed

  Install git using one of:

    • Xcode Command Line Tools (recommended):
      xcode-select --install

    • Homebrew:
      brew install git

    • Manual download:
      https://git-scm.com/download/mac
```

**Linux:**
```
✗ Git is not installed

  Install git using your package manager:

    • Debian/Ubuntu:
      sudo apt install git

    • Fedora:
      sudo dnf install git

    • Arch:
      sudo pacman -S git

    • Alpine:
      sudo apk add git
```

---

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

## Platforms

| OS      | Architecture | Binary                       |
| ------- | ------------ | ---------------------------- |
| Linux   | amd64        | `ztigit-linux-amd64`         |
| Linux   | arm64        | `ztigit-linux-arm64`         |
| macOS   | amd64        | `ztigit-darwin-amd64`        |
| macOS   | arm64        | `ztigit-darwin-arm64`        |
| Windows | amd64        | `ztigit-windows-amd64.exe`   |
| Windows | arm64        | `ztigit-windows-arm64.exe`   |

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.3...0.0.4
