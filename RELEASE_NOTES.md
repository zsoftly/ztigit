# ztigit v0.0.5 Release Notes

## What's New

### Mirror multiple groups

Mirror more than one GitLab group / GitHub org in a single command:

```bash
# Comma-separated
ztigit mirror group1,group2,group3 -p gitlab

# Space-separated
ztigit mirror --groups "group1 group2 group3" -p gitlab
```

When mirroring multiple groups, the default directory becomes provider-specific:

- GitLab: `$HOME/gitlab-repos/...`
- GitHub: `$HOME/github-repos/...`

### Preserve GitLab subgroup nesting

GitLab subgroup paths are now preserved in the local directory structure. For example,
`my-group/my-subgroup/my-project` mirrors into `BaseDir/my-group/my-subgroup/my-project`.

### Better path validation + clearer errors

Mirroring now validates repository paths before cloning/updating and surfaces clearer errors for
invalid inputs (including guarding against overly long paths on Windows).

### More mirror tests

Added a comprehensive test suite around mirror behavior to reduce regressions.

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

| OS      | Architecture | Binary                     |
| ------- | ------------ | -------------------------- |
| Linux   | amd64        | `ztigit-linux-amd64`       |
| Linux   | arm64        | `ztigit-linux-arm64`       |
| macOS   | amd64        | `ztigit-darwin-amd64`      |
| macOS   | arm64        | `ztigit-darwin-arm64`      |
| Windows | amd64        | `ztigit-windows-amd64.exe` |
| Windows | arm64        | `ztigit-windows-arm64.exe` |

**Full Changelog**: https://github.com/zsoftly/ztigit/compare/0.0.4...0.0.5
