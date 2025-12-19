# ztigit Makefile
# Cross-platform build targets for GitLab/GitHub CLI tool

BINARY_NAME=ztigit
VERSION?=0.1.0
BUILD_DIR=bin
CMD_DIR=cmd/ztigit

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build build-all clean test fmt tidy deps install

# Default target
all: build

# Build for current platform
build:
	@echo "Building $(BINARY_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Binary created: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
build-all: build-linux build-darwin build-windows
	@echo "All builds complete!"

build-linux:
	@echo "Building for Linux (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)

build-darwin:
	@echo "Building for macOS (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)

build-windows:
	@echo "Building for Windows (amd64)..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)

# Run tests
test:
	$(GOTEST) -v ./...

# Format code (Go + Markdown/JSON/YAML)
fmt:
	$(GOFMT) ./...
	@npx prettier --write "**/*.md" "**/*.json" "**/*.yaml" "**/*.yml" 2>/dev/null || true

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Download dependencies
deps:
	$(GOMOD) download

# Install locally
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installed successfully!"

# Install to user's local bin (no sudo)
install-user: build
	@echo "Installing $(BINARY_NAME) to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@echo "Installed to ~/bin/$(BINARY_NAME)"
	@echo "Make sure ~/bin is in your PATH"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete!"

# Help
help:
	@echo "ztigit Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build for current platform"
	@echo "  make build-all    Build for all platforms (Linux, macOS, Windows)"
	@echo "  make build-linux  Build for Linux (amd64, arm64)"
	@echo "  make build-darwin Build for macOS (amd64, arm64)"
	@echo "  make build-windows Build for Windows (amd64)"
	@echo "  make test         Run tests"
	@echo "  make fmt          Format code (Go + Markdown/JSON/YAML)"
	@echo "  make tidy         Tidy go.mod"
	@echo "  make deps         Download dependencies"
	@echo "  make install      Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user Install to ~/bin"
	@echo "  make clean        Remove build artifacts"
	@echo "  make help         Show this help"
