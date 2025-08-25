# QuikGit Makefile

# Variables
APP_NAME := quikgit
VERSION := 1.0.0
BUILD_DIR := build
DIST_DIR := dist
CMD_DIR := cmd/quikgit

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"
CGO_ENABLED := 0

# Platforms for cross-compilation
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64 \
	windows/arm64

.PHONY: all build clean test install dev deps fmt lint check cross-compile package help

# Default target
all: build

# Build the application
build:
	@echo "Building $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Development build with debugging symbols
dev:
	@echo "Building $(APP_NAME) for development..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(APP_NAME)-dev ./$(CMD_DIR)
	@echo "Development build complete: $(BUILD_DIR)/$(APP_NAME)-dev"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, using go vet instead"; \
		go vet ./...; \
		echo "For better linting, install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Run comprehensive checks
check: fmt lint test
	@echo "All checks passed"

# Install the application
install: build
	@echo "Installing $(APP_NAME)..."
	cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME)
	@echo "$(APP_NAME) installed to $(GOPATH)/bin/$(APP_NAME)"

# Cross-compile for all platforms
cross-compile:
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		EXT=""; \
		if [ "$$OS" = "windows" ]; then EXT=".exe"; fi; \
		echo "Building for $$OS/$$ARCH..."; \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$$OS GOARCH=$$ARCH go build $(LDFLAGS) \
			-o $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$$OS-$$ARCH$$EXT ./$(CMD_DIR); \
	done
	@echo "Cross-compilation complete"

# Package releases
package: cross-compile
	@echo "Packaging releases..."
	@cd $(DIST_DIR) && for file in $(APP_NAME)-$(VERSION)-*; do \
		if [[ "$$file" == *"windows"* ]]; then \
			zip "$$file.zip" "$$file"; \
		else \
			tar -czf "$$file.tar.gz" "$$file"; \
		fi; \
	done
	@echo "Packaging complete"

# Generate release checksums
checksums: package
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.sha256
	@echo "Checksums generated"

# Create Homebrew formula
homebrew-formula:
	@echo "Creating Homebrew formula..."
	@mkdir -p $(DIST_DIR)/homebrew
	@sed 's/{{VERSION}}/$(VERSION)/g; s/{{SHA256}}/$(shell cd $(DIST_DIR) && sha256sum $(APP_NAME)-$(VERSION)-darwin-amd64.tar.gz | cut -d" " -f1)/g' \
		scripts/package-brew.rb.template > $(DIST_DIR)/homebrew/$(APP_NAME).rb
	@echo "Homebrew formula created: $(DIST_DIR)/homebrew/$(APP_NAME).rb"

# Create AUR PKGBUILD
aur-package:
	@echo "Creating AUR PKGBUILD..."
	@mkdir -p $(DIST_DIR)/aur
	@sed 's/{{VERSION}}/$(VERSION)/g; s/{{SHA256}}/$(shell cd $(DIST_DIR) && sha256sum $(APP_NAME)-$(VERSION)-linux-amd64.tar.gz | cut -d" " -f1)/g' \
		scripts/PKGBUILD.template > $(DIST_DIR)/aur/PKGBUILD
	@echo "AUR PKGBUILD created: $(DIST_DIR)/aur/PKGBUILD"

# Create Debian package
deb-package: cross-compile
	@echo "Creating Debian package..."
	@mkdir -p $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/{DEBIAN,usr/bin}
	@cp $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64 $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/usr/bin/$(APP_NAME)
	@chmod +x $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/usr/bin/$(APP_NAME)
	@echo "Package: $(APP_NAME)" > $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Section: utils" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Priority: optional" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Architecture: amd64" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Maintainer: QuikGit Team" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@echo "Description: GitHub repository manager TUI" >> $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64/DEBIAN/control
	@dpkg-deb --build $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64
	@echo "Debian package created: $(DIST_DIR)/deb/$(APP_NAME)_$(VERSION)_amd64.deb"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f coverage.out
	@echo "Clean complete"

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Run development version
run-dev: dev
	@echo "Running $(APP_NAME) in development mode..."
	./$(BUILD_DIR)/$(APP_NAME)-dev --debug

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "Docker image built: $(APP_NAME):$(VERSION)"

# Docker run
docker-run: docker-build
	@echo "Running $(APP_NAME) in Docker..."
	docker run -it --rm $(APP_NAME):$(VERSION)

# Release workflow
release: clean check cross-compile package checksums homebrew-formula aur-package
	@echo "Release $(VERSION) ready in $(DIST_DIR)/"

# Show help
help:
	@echo "QuikGit v$(VERSION) - Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          Build the application"
	@echo "  dev            Build development version with debugging"
	@echo "  deps           Install dependencies"
	@echo "  fmt            Format source code"
	@echo "  lint           Lint source code"
	@echo "  test           Run tests"
	@echo "  check          Run fmt, lint, and test"
	@echo "  install        Install the application"
	@echo "  cross-compile  Build for all platforms"
	@echo "  package        Create release packages"
	@echo "  checksums      Generate release checksums"
	@echo "  release        Complete release workflow"
	@echo "  clean          Clean build artifacts"
	@echo "  run            Build and run the application"
	@echo "  run-dev        Build and run development version"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run in Docker container"
	@echo "  help           Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  BUILD_DIR=$(BUILD_DIR)"
	@echo "  DIST_DIR=$(DIST_DIR)"