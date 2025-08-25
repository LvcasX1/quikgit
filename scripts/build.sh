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

# Cross-compile for all platforms
cross_compile() {
    log_info "Starting cross-compilation..."
    mkdir -p "$DIST_DIR"
    
    # Define platforms
    local platforms=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
        "windows/arm64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS='/' read -r os arch <<< "$platform"
        build_platform "$os" "$arch" "$DIST_DIR"
    done
    
    log_success "Cross-compilation complete"
}

# Package releases
package_releases() {
    log_info "Packaging releases..."
    
    cd "$DIST_DIR"
    
    for file in ${APP_NAME}-${VERSION}-*; do
        if [[ "$file" == *"windows"* ]]; then
            log_info "Creating ZIP for $file"
            zip "${file}.zip" "$file"
        else
            log_info "Creating TAR.GZ for $file"
            tar -czf "${file}.tar.gz" "$file"
        fi
    done
    
    cd ..
    log_success "Packaging complete"
}

# Generate checksums
generate_checksums() {
    log_info "Generating checksums..."
    
    cd "$DIST_DIR"
    sha256sum *.tar.gz *.zip > checksums.sha256
    cd ..
    
    log_success "Checksums generated"
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
            log_info "Starting full build process for QuikGit v$VERSION"
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
        "help"|"-h"|"--help")
            cat << EOF
QuikGit Build Script

Usage: $0 [command]

Commands:
  deps        Install dependencies
  clean       Clean build artifacts
  format      Format source code
  lint        Lint source code
  test        Run tests
  build       Cross-compile for all platforms
  package     Package releases
  docker      Build Docker image
  all         Run complete build process (default)
  release     Alias for 'all'
  help        Show this help message

Environment variables:
  VERSION     Set build version (default: 1.0.0)
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