#!/bin/bash

# QuikGit Build Script
# This script builds QuikGit for multiple platforms and packages releases

set -euo pipefail

# Configuration
APP_NAME="quikgit"
VERSION="${VERSION:-1.0.0}"
BUILD_DIR="build"
DIST_DIR="dist"
CMD_DIR="cmd/quikgit"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v git &> /dev/null; then
        log_error "Git is not installed"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Optional tools
    if command -v golangci-lint &> /dev/null; then
        log_info "golangci-lint found"
    else
        log_warning "golangci-lint not found. Linting will use go vet instead"
    fi
    
    if command -v docker &> /dev/null; then
        log_info "Docker found"
    else
        log_warning "Docker not found. Docker builds will be skipped"
    fi
    
    log_success "Dependencies check complete"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    rm -f coverage.out
    log_success "Clean complete"
}

# Install Go dependencies
install_deps() {
    log_info "Installing Go dependencies..."
    go mod download
    go mod tidy
    log_success "Dependencies installed"
}

# Format code
format_code() {
    log_info "Formatting code..."
    go fmt ./...
    log_success "Code formatted"
}

# Lint code
lint_code() {
    log_info "Linting code..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
    else
        go vet ./...
    fi
    log_success "Linting complete"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    go test -v -race -coverprofile=coverage.out ./...
    
    # Generate coverage report if tests pass
    if [ -f coverage.out ]; then
        COVERAGE=$(go tool cover -func=coverage.out | tail -n 1 | awk '{print $3}')
        log_info "Test coverage: $COVERAGE"
    fi
    
    log_success "Tests complete"
}

# Build for single platform
build_platform() {
    local os=$1
    local arch=$2
    local output_dir=$3
    
    local ext=""
    if [ "$os" = "windows" ]; then
        ext=".exe"
    fi
    
    local output_file="$output_dir/${APP_NAME}-${VERSION}-${os}-${arch}${ext}"
    
    log_info "Building for $os/$arch..."
    
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.version=$VERSION -s -w" \
        -o "$output_file" \
        "./$CMD_DIR"
    
    log_success "Built $os/$arch: $output_file"
}

# Cross-compile for all platforms using GoReleaser
cross_compile() {
    log_info "Starting cross-compilation with GoReleaser..."
    
    # Check if GoReleaser is installed
    if ! command -v goreleaser &> /dev/null; then
        log_error "GoReleaser is not installed. Please install it first:"
        log_info "  macOS: brew install goreleaser"
        log_info "  Linux: See https://goreleaser.com/install"
        exit 1
    fi
    
    # Use GoReleaser for cross-compilation (snapshot mode for local builds)
    if goreleaser build --snapshot --clean; then
        log_success "GoReleaser cross-compilation complete"
    else
        log_error "GoReleaser build failed"
        exit 1
    fi
}

# Package releases using GoReleaser
package_releases() {
    log_info "Packaging releases with GoReleaser..."
    
    # Check if GoReleaser is installed
    if ! command -v goreleaser &> /dev/null; then
        log_error "GoReleaser is not installed. Please install it first:"
        log_info "  macOS: brew install goreleaser"
        log_info "  Linux: See https://goreleaser.com/install"
        exit 1
    fi
    
    # Check if Docker is available
    local skip_docker=""
    if ! command -v docker &> /dev/null || ! docker info &>/dev/null; then
        log_warning "Docker not available. Skipping Docker builds..."
        skip_docker="--skip=docker"
    fi
    
    # Use GoReleaser for packaging (snapshot mode for local packaging)
    if goreleaser release --snapshot --clean --skip=publish $skip_docker; then
        log_success "GoReleaser packaging complete"
    else
        log_error "GoReleaser packaging failed"
        exit 1
    fi
}

# Generate checksums (now handled by GoReleaser)
generate_checksums() {
    log_info "Checksums are now generated automatically by GoReleaser"
    log_success "Checksums generated"
}

# GoReleaser full release (publishes to GitHub)
goreleaser_release() {
    log_info "Starting GoReleaser full release..."
    
    # Check if GoReleaser is installed
    if ! command -v goreleaser &> /dev/null; then
        log_error "GoReleaser is not installed. Please install it first:"
        log_info "  macOS: brew install goreleaser"
        log_info "  Linux: See https://goreleaser.com/install"
        exit 1
    fi
    
    # Check for GITHUB_TOKEN
    if [ -z "$GITHUB_TOKEN" ]; then
        log_warning "GITHUB_TOKEN not set. GoReleaser will skip GitHub release creation."
        log_info "To set up GitHub token:"
        log_info "  1. Go to GitHub → Settings → Developer settings → Personal access tokens"
        log_info "  2. Create token with 'repo' and 'write:packages' permissions"
        log_info "  3. Export it: export GITHUB_TOKEN=\"your_token_here\""
    fi
    
    # Check for optional Linux packaging credentials
    if [ -z "$AUR_SSH_PRIVATE_KEY" ]; then
        log_info "AUR_SSH_PRIVATE_KEY not set. AUR publishing will be skipped."
        log_info "To enable AUR publishing:"
        log_info "  1. Set up AUR account and SSH key"
        log_info "  2. Export key: export AUR_SSH_PRIVATE_KEY=\"\$(cat ~/.ssh/aur_rsa)\""
    else
        log_info "AUR publishing enabled"
    fi
    
    # Check for APT repository setup
    if [ -n "$APT_REPO_URL" ] && [ -n "$APT_REPO_KEY" ]; then
        log_info "APT repository publishing enabled: $APT_REPO_URL"
    else
        log_info "APT repository publishing disabled (APT_REPO_URL/APT_REPO_KEY not set)"
    fi
    
    # Check if Docker is available
    local skip_docker=""
    if ! command -v docker &> /dev/null || ! docker info &>/dev/null; then
        log_warning "Docker not available. Skipping Docker builds..."
        skip_docker="--skip=docker"
    fi
    
    # Check if we're on a tagged commit
    if ! git describe --tags --exact-match HEAD &>/dev/null; then
        log_warning "No tag found on current commit. Creating snapshot release..."
        if goreleaser release --snapshot --clean --skip=publish $skip_docker; then
            log_success "GoReleaser snapshot release complete (local only)"
            log_info "To create a full release:"
            log_info "  1. Create and push a tag: git tag v1.0.0 && git push origin v1.0.0"
            log_info "  2. Run: goreleaser release"
        else
            log_error "GoReleaser snapshot release failed"
            exit 1
        fi
    else
        # Full release with tag
        if goreleaser release $skip_docker; then
            log_success "GoReleaser full release complete"
            log_info "Release published to GitHub with all artifacts"
        else
            log_error "GoReleaser release failed"
            exit 1
        fi
    fi
}

# Create Homebrew formula
create_homebrew_formula() {
    log_info "Creating Homebrew formula..."
    
    mkdir -p "$DIST_DIR/homebrew"
    
    # Get SHA256 for macOS AMD64 build
    local macos_file="$DIST_DIR/${APP_NAME}-${VERSION}-darwin-amd64.tar.gz"
    if [ -f "$macos_file" ]; then
        local sha256=$(sha256sum "$macos_file" | cut -d' ' -f1)
        
        cat > "$DIST_DIR/homebrew/${APP_NAME}.rb" << EOF
class Quikgit < Formula
  desc "GitHub repository manager TUI"
  homepage "https://github.com/lvcasx1/quikgit"
  url "https://github.com/lvcasx1/quikgit/releases/download/v${VERSION}/${APP_NAME}-${VERSION}-darwin-amd64.tar.gz"
  sha256 "$sha256"
  license "MIT"

  depends_on "git"

  def install
    bin.install "quikgit"
  end

  test do
    system "#{bin}/quikgit", "--version"
  end
end
EOF
        
        log_success "Homebrew formula created"
    else
        log_warning "macOS build not found, skipping Homebrew formula"
    fi
}

# Create AUR PKGBUILD
create_aur_package() {
    log_info "Creating AUR PKGBUILD..."
    
    mkdir -p "$DIST_DIR/aur"
    
    # Get SHA256 for Linux AMD64 build
    local linux_file="$DIST_DIR/${APP_NAME}-${VERSION}-linux-amd64.tar.gz"
    if [ -f "$linux_file" ]; then
        local sha256=$(sha256sum "$linux_file" | cut -d' ' -f1)
        
        cat > "$DIST_DIR/aur/PKGBUILD" << EOF
# Maintainer: QuikGit Team
pkgname=quikgit
pkgver=${VERSION}
pkgrel=1
pkgdesc="GitHub repository manager TUI"
arch=('x86_64')
url="https://github.com/lvcasx1/quikgit"
license=('MIT')
depends=('glibc' 'git')
source=("https://github.com/lvcasx1/quikgit/releases/download/v\$pkgver/quikgit-\$pkgver-linux-amd64.tar.gz")
sha256sums=('$sha256')

package() {
    install -Dm755 "quikgit-\$pkgver-linux-amd64" "\$pkgdir/usr/bin/quikgit"
}
EOF
        
        log_success "AUR PKGBUILD created"
    else
        log_warning "Linux build not found, skipping AUR package"
    fi
}

# Build Docker image
build_docker() {
    if ! command -v docker &> /dev/null; then
        log_warning "Docker not found, skipping Docker build"
        return
    fi
    
    log_info "Building Docker image..."
    
    cat > Dockerfile << EOF
FROM alpine:latest

RUN apk add --no-cache ca-certificates git

WORKDIR /root/

COPY dist/quikgit-${VERSION}-linux-amd64 /usr/local/bin/quikgit

RUN chmod +x /usr/local/bin/quikgit

ENTRYPOINT ["quikgit"]
EOF
    
    docker build -t "${APP_NAME}:${VERSION}" .
    docker tag "${APP_NAME}:${VERSION}" "${APP_NAME}:latest"
    
    # Clean up Dockerfile
    rm Dockerfile
    
    log_success "Docker image built: ${APP_NAME}:${VERSION}"
}

# Generate release notes
generate_release_notes() {
    log_info "Generating release notes..."
    
    cat > "$DIST_DIR/RELEASE_NOTES.md" << EOF
# QuikGit v${VERSION}

## Features
- GitHub OAuth authentication with QR code support
- Repository search and filtering
- Multi-repository cloning with progress tracking
- Automatic dependency detection and installation
- Support for multiple programming languages and frameworks
- Terminal user interface with mouse and keyboard support

## Installation

### macOS (Homebrew)
\`\`\`bash
brew tap lvcasx1/tap
brew install quikgit
\`\`\`

### Arch Linux (AUR)
\`\`\`bash
yay -S quikgit
\`\`\`

### Manual Installation
Download the appropriate binary for your platform from the release assets.

## Usage
\`\`\`bash
quikgit --help
\`\`\`

## Supported Languages
- Go, Node.js, Python, Ruby, Rust
- Java, C++, C#, Swift, PHP, Dart
- And more!

## What's Changed
- Initial release of QuikGit v${VERSION}

**Full Changelog**: https://github.com/lvcasx1/quikgit/commits/v${VERSION}
EOF
    
    log_success "Release notes generated"
}

# Display build summary
display_summary() {
    log_success "Build Summary"
    echo "=============="
    echo "Version: $VERSION"
    echo "Build directory: $BUILD_DIR"
    echo "Distribution directory: $DIST_DIR"
    echo ""
    echo "Generated files:"
    if [ -d "$DIST_DIR" ]; then
        ls -la "$DIST_DIR"
    fi
    echo ""
    log_success "Build complete!"
}

# Main build function
main() {
    local command="${1:-all}"
    
    case $command in
        "deps")
            install_deps
            ;;
        "clean")
            clean_build
            ;;
        "format")
            format_code
            ;;
        "lint")
            lint_code
            ;;
        "test")
            run_tests
            ;;
        "build")
            cross_compile
            ;;
        "package")
            package_releases
            ;;
        "docker")
            build_docker
            ;;
        "all"|"release")
            log_info "Starting full release process for QuikGit v$VERSION"
            check_dependencies
            clean_build
            install_deps
            format_code
            lint_code
            run_tests
            goreleaser_release
            display_summary
            ;;
        "legacy-release")
            log_info "Starting legacy build process for QuikGit v$VERSION"
            log_warning "This is the old release process. Consider using 'release' with GoReleaser instead."
            check_dependencies
            clean_build
            install_deps
            format_code
            lint_code
            run_tests
            cross_compile
            package_releases
            generate_checksums
            create_homebrew_formula
            create_aur_package
            build_docker
            generate_release_notes
            display_summary
            ;;
        "goreleaser")
            goreleaser_release
            ;;
        "help"|"-h"|"--help")
            cat << EOF
QuikGit Build Script

Usage: $0 [command]

Commands:
  deps            Install dependencies
  clean           Clean build artifacts
  format          Format source code
  lint            Lint source code
  test            Run tests
  build           Cross-compile for all platforms (uses GoReleaser)
  package         Package releases (uses GoReleaser)
  docker          Build Docker image
  goreleaser      Run GoReleaser release (smart: snapshot if no tag, full release if tagged)
  all             Run complete release process using GoReleaser (default)
  release         Alias for 'all' - uses GoReleaser for modern release workflow
  legacy-release  Run old build process (without GoReleaser)
  help            Show this help message

Modern Workflow (Recommended):
  ./scripts/build.sh release        # Uses GoReleaser, creates GitHub release if tagged

Legacy Workflow (Deprecated):
  ./scripts/build.sh legacy-release # Uses old build process

Environment variables:
  VERSION       Set build version (default: 1.0.0)
  GITHUB_TOKEN  GitHub token for releases (required for publishing)

Setup Environment Variables:
  export GITHUB_TOKEN="ghp_your_token_here"              # Required for GitHub releases
  export AUR_SSH_PRIVATE_KEY="$(cat ~/.ssh/aur_rsa)"     # Optional: for AUR publishing
  export APT_REPO_URL="https://apt.yourrepo.com"         # Optional: for APT publishing
  export APT_REPO_KEY="your_apt_signing_key"             # Optional: for APT publishing

Release Process:
  1. git tag v1.0.0 && git push origin v1.0.0
  2. ./scripts/build.sh release

What Gets Published Automatically:
  ✅ GitHub Releases (with binaries, archives, checksums)
  ✅ Homebrew Formula (to lvcasx1/homebrew-tap)  
  ✅ Docker Images (to ghcr.io/lvcasx1/quikgit)
  ✅ DEB/RPM Packages (generated, ready for distribution)
  ✅ AUR Package (published to aur.archlinux.org if AUR_SSH_PRIVATE_KEY set)
EOF
            ;;
        *)
            log_error "Unknown command: $command"
            log_info "Run '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"