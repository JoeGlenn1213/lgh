# LGH Makefile
# Build, test, and release automation

VERSION ?= 1.0.0
BUILD_DATE := $(shell date +%Y-%m-%d)
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.GitCommit=$(GIT_COMMIT)

BINARY_NAME := lgh
DIST_DIR := dist
PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64


.PHONY: all build clean test install uninstall release checksums

# Default target
all: build

# Build for current platform
build:
	@echo "Building LGH v$(VERSION)..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY_NAME) ./cmd/lgh/
	@echo "✓ Built: $(DIST_DIR)/$(BINARY_NAME)"

# Build for all platforms
release: clean
	@echo "Building LGH v$(VERSION) for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output=$(DIST_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ "$$GOOS" = "windows" ]; then output=$$output.exe; fi; \
		echo "Building $$output..."; \
		CGO_ENABLED=0 GOOS=$$GOOS GOARCH=$$GOARCH go build -ldflags="$(LDFLAGS)" -o $$output ./cmd/lgh/; \
	done
	@echo ""
	@echo "✓ Release builds complete:"
	@ls -la $(DIST_DIR)/
	@$(MAKE) checksums


# Generate SHA256 checksums for release binaries
checksums:
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && shasum -a 256 lgh-* > checksums.txt
	@echo "✓ Checksums saved to $(DIST_DIR)/checksums.txt"
	@cat $(DIST_DIR)/checksums.txt

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -cover

# Run short tests (no integration)
test-short:
	@echo "Running short tests..."
	go test ./... -v -short

# Install to /usr/local/bin
install: build
	@echo "Installing LGH to /usr/local/bin..."
	@sudo cp $(DIST_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed: /usr/local/bin/$(BINARY_NAME)"
	@$(BINARY_NAME) --version

# Uninstall from /usr/local/bin
uninstall:
	@echo "Uninstalling LGH..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Uninstalled: /usr/local/bin/$(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(DIST_DIR)
	@go clean
	@echo "✓ Clean complete"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Format complete"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./...

# Security check
security:
	@echo "Running security checks..."
	@go vet ./...
	@echo "✓ Security check complete"

# Show help
help:
	@echo "LGH Makefile Commands:"
	@echo ""
	@echo "  make build       Build for current platform"
	@echo "  make release     Build for all platforms + generate checksums"
	@echo "  make checksums   Generate SHA256 checksums"
	@echo "  make test        Run all tests"
	@echo "  make test-short  Run short tests (skip integration)"
	@echo "  make install     Install to /usr/local/bin"
	@echo "  make uninstall   Remove from /usr/local/bin"
	@echo "  make clean       Clean build artifacts"
	@echo "  make fmt         Format code"
	@echo "  make security    Run security checks"
	@echo "  make help        Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"

