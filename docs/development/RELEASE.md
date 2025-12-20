# Release Process

Tag-based releases. Push version tag → automated builds → GitHub Release.

## Steps

```bash
# 1. Create release branch
git checkout main && git pull origin main
git checkout -b release/v<version>
git push -u origin release/v<version>

# 2. Tag (triggers build)
git tag <version>
git push origin <version>

# 3. Merge back
git checkout main
git merge release/v<version>
git push origin main
git branch -d release/v<version>
```

## Version Format

`<major>.<minor>.<patch>` (semver, no `v` prefix)

## Build Artifacts

| Platform      | Binary                     |
| ------------- | -------------------------- |
| Linux amd64   | `ztigit-linux-amd64`       |
| Linux arm64   | `ztigit-linux-arm64`       |
| macOS amd64   | `ztigit-darwin-amd64`      |
| macOS arm64   | `ztigit-darwin-arm64`      |
| Windows amd64 | `ztigit-windows-amd64.exe` |
| Windows arm64 | `ztigit-windows-arm64.exe` |

Also included:

- `install.sh` - Unix one-liner installer
- `install.ps1` - Windows one-liner installer
- Archives (`.tar.gz`, `.zip`)
- `checksums.txt`

## Troubleshooting

```bash
# Check workflow
gh run list
gh run view <run-id> --log

# Local test
go build ./cmd/ztigit && ./ztigit --version
```
