# Installation

## Binary Download

Pre-built binaries available for:

| OS      | Architecture          | Binary                     |
| ------- | --------------------- | -------------------------- |
| Linux   | amd64                 | `ztigit-linux-amd64`       |
| Linux   | arm64                 | `ztigit-linux-arm64`       |
| macOS   | amd64 (Intel)         | `ztigit-darwin-amd64`      |
| macOS   | arm64 (Apple Silicon) | `ztigit-darwin-arm64`      |
| Windows | amd64                 | `ztigit-windows-amd64.exe` |
| Windows | arm64                 | `ztigit-windows-arm64.exe` |

### Linux/macOS

```bash
# Linux (amd64)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-linux-amd64 -o ztigit

# Linux (arm64)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-linux-arm64 -o ztigit

# macOS (Intel)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-darwin-amd64 -o ztigit

# macOS (Apple Silicon)
curl -L https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-darwin-arm64 -o ztigit

# Make executable and install
chmod +x ztigit
sudo mv ztigit /usr/local/bin/

# Verify
ztigit --version
```

### Windows

```powershell
# Download
Invoke-WebRequest -Uri https://github.com/zsoftly/ztigit/releases/latest/download/ztigit-windows-amd64.exe -OutFile ztigit.exe

# Add to PATH or move to a directory in PATH
Move-Item ztigit.exe C:\Windows\System32\

# Verify
ztigit --version
```

## Build from Source

Requirements:

- Go 1.24+
- Git

```bash
# Clone
git clone https://github.com/zsoftly/ztigit.git
cd ztigit

# Build
go build -o ztigit ./cmd/ztigit

# Or use make
make build

# Install to /usr/local/bin
sudo make install

# Or install to ~/bin
make install-user
```

### Cross-Compile

```bash
# All platforms
make build-all

# Specific platform
GOOS=darwin GOARCH=arm64 go build -o ztigit-darwin-arm64 ./cmd/ztigit
```

## Verify Installation

```bash
ztigit --version
ztigit --help
```

## Dependencies

Runtime:

- `git` - required for clone/pull operations
- **Linux only**: `libsecret` for keychain support (optional, falls back to config file)

  ```bash
  # Debian/Ubuntu
  sudo apt install libsecret-1-0

  # Fedora/RHEL
  sudo dnf install libsecret
  ```

Build:

- Go 1.24+
