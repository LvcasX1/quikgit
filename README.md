# QuikGit 🚀

[![Go Version](https://img.shields.io/github/go-mod/go-version/lvcasx1/quikgit)](https://golang.org/)
[![Release](https://img.shields.io/github/v/release/lvcasx1/quikgit)](https://github.com/lvcasx1/quikgit/releases)
[![License](https://img.shields.io/github/license/lvcasx1/quikgit)](LICENSE)
[![AUR Package](https://img.shields.io/aur/version/quikgit)](https://aur.archlinux.org/packages/quikgit)

A powerful Terminal User Interface (TUI) for GitHub repository management that streamlines the process of discovering, cloning, and setting up projects in your local development environment.

![QuikGit Demo](docs/images/demo.gif)

## ✨ Features

### 🔐 **Seamless Authentication**
- **QR Code Login**: Authenticate with GitHub using a QR code
- **Secure Token Management**: Safely store and manage access tokens
- **Session Persistence**: Stay logged in between sessions

### 🔍 **Powerful Repository Discovery**
- **Advanced Search**: Search by name, language, topics, and more
- **Smart Filtering**: Filter by user, organization, or specific criteria  
- **Real-time Results**: Instant search with pagination support
- **Rich Information**: View stars, forks, languages, and last updated

### 📦 **Efficient Multi-Repository Cloning**
- **Parallel Processing**: Clone multiple repositories simultaneously
- **Progress Tracking**: Real-time progress with animated progress bars
- **Smart Conflict Handling**: Handle existing directories gracefully
- **SSH & HTTPS Support**: Choose your preferred cloning method

### 🛠️ **Automatic Dependency Installation**
- **Language Detection**: Automatically detect project types
- **Smart Installation**: Run appropriate dependency managers
- **Concurrent Processing**: Install dependencies for multiple projects
- **Error Handling**: Continue processing even if some installations fail

### 🎨 **Beautiful Terminal Interface**
- **Responsive Design**: Adapts to different terminal sizes
- **Mouse Support**: Full mouse interaction support
- **Keyboard Navigation**: Comprehensive keyboard shortcuts
- **Animated Transitions**: Smooth animations using Harmonica
- **Customizable Themes**: Multiple color schemes

### 📋 **Supported Project Types**

| Language | Files | Commands |
|----------|-------|----------|
| **Go** | `go.mod`, `go.sum` | `go mod tidy`, `go mod download` |
| **Node.js** | `package.json` | `npm install` / `yarn install` |
| **Python** | `requirements.txt`, `Pipfile`, `pyproject.toml` | `pip install`, `pipenv install`, `poetry install` |
| **Ruby** | `Gemfile` | `bundle install` |
| **Rust** | `Cargo.toml` | `cargo build` |
| **Java** | `pom.xml`, `build.gradle` | `mvn install`, `gradle build` |
| **C++** | `CMakeLists.txt` | `cmake`, `make` |
| **C#** | `*.csproj`, `*.sln` | `dotnet restore`, `dotnet build` |
| **PHP** | `composer.json` | `composer install` |
| **Swift** | `Package.swift` | `swift build` |
| **Dart** | `pubspec.yaml` | `flutter pub get` |

## 🚀 Installation

### macOS (Homebrew)

```bash
brew tap lvcasx1/tap
brew install quikgit
```

### Arch Linux (AUR)

```bash
# Using yay
yay -S quikgit

# Using paru
paru -S quikgit

# Manual
git clone https://aur.archlinux.org/quikgit.git
cd quikgit
makepkg -si
```

### Ubuntu/Debian (.deb package)

```bash
# Download the .deb package from releases
wget https://github.com/lvcasx1/quikgit/releases/download/v1.0.0/quikgit_1.0.0_amd64.deb
sudo dpkg -i quikgit_1.0.0_amd64.deb

# Install dependencies if needed
sudo apt-get install -f
```

### Manual Installation

1. Download the latest release for your platform:
   ```bash
   curl -LO https://github.com/lvcasx1/quikgit/releases/download/v1.0.0/quikgit-1.0.0-linux-amd64.tar.gz
   ```

2. Extract and install:
   ```bash
   tar -xzf quikgit-1.0.0-linux-amd64.tar.gz
   sudo mv quikgit-1.0.0-linux-amd64 /usr/local/bin/quikgit
   sudo chmod +x /usr/local/bin/quikgit
   ```

### Build from Source

```bash
git clone https://github.com/lvcasx1/quikgit.git
cd quikgit
go build -o quikgit ./cmd/quikgit
sudo mv quikgit /usr/local/bin/
```

### Docker

```bash
docker run -it --rm -v $(pwd):/workspace ghcr.io/lvcasx1/quikgit:latest
```

## 🎯 Quick Start

### Step 1: Authentication Setup (One-time)

Run the setup script for easy authentication configuration:

```bash
# After installation, run the setup script
./setup-oauth.sh
```

**Or manually choose your authentication method:**

- **Option A - Personal Access Token** (2 minutes):
  ```bash
  export GITHUB_TOKEN=your_token_here  # Get from: https://github.com/settings/tokens/new
  ```

- **Option B - OAuth with QR Code** (5 minutes):
  ```bash
  export GITHUB_CLIENT_ID=your_client_id  # Create app at: https://github.com/settings/applications/new
  ```

### Step 2: Launch QuikGit

```bash
quikgit
```

**Then:**
1. **Authenticate**: Scan the QR code or visit the provided URL
2. **Search Repositories**: Use the search interface to find repositories  
3. **Select and Clone**: Choose repositories and let QuikGit handle cloning and setup

## ⌨️ Keyboard Shortcuts

### Global
- `q` / `Ctrl+C`: Quit application
- `Esc`: Go back to previous screen  
- `F1` / `?`: Show help
- `Tab`: Navigate between fields
- `Enter`: Confirm/Select

### Search Interface
- `Ctrl+L`: Select language filter
- `Ctrl+S`: Change sort options
- `Tab`: Switch between input fields

### Repository Lists
- `Space`: Toggle selection
- `a`: Select all repositories
- `n`: Clear all selections
- `j/k` or `↑/↓`: Navigate items

### During Operations
- `d`: Toggle detailed output view
- `Ctrl+C`: Cancel ongoing operations

## 🛠️ Configuration

QuikGit stores configuration in `~/.quikgit/config.yaml`:

```yaml
github:
  prefer_ssh: false
  ssh_key_path: ~/.ssh/id_rsa
  default_user: ""
  default_org: ""

clone:
  concurrent: 3
  use_current_dir: true
  create_subdirs: false
  default_path: ~/projects

install:
  enabled: true
  concurrent: 3
  timeout_minutes: 10
  skip_on_error: false
  auto_install: true

ui:
  theme: default
  show_icons: true
  animations_speed: normal
  mouse_support: true
  show_line_numbers: false

defaults:
  search_sort: stars
  search_order: desc
  results_per_page: 30
  preferred_auth: https
```

### Environment Variables

- `QUIKGIT_CONFIG`: Path to custom configuration file
- `QUIKGIT_DEBUG`: Enable debug logging
- `GITHUB_TOKEN`: Pre-set GitHub personal access token

## 📚 Usage Examples

### Basic Repository Search
```bash
quikgit
# 1. Select "Search repositories"
# 2. Enter search terms
# 3. Select repositories
# 4. Press Enter to clone
```

### Command Line Options
```bash
# Show version
quikgit --version

# Use custom config
quikgit --config /path/to/config.yaml

# Enable debug mode
quikgit --debug

# Show help
quikgit --help
```

## 🔧 Development

### Prerequisites
- Go 1.21 or later
- Git
- Make (optional)

### Building

```bash
# Clone the repository
git clone https://github.com/lvcasx1/quikgit.git
cd quikgit

# Install dependencies
go mod download

# Build
make build
# or
go build -o quikgit ./cmd/quikgit

# Run tests
make test
# or
go test ./...

# Build for all platforms
make cross-compile
```

### Development Workflow

```bash
# Format code
make fmt

# Lint code  
make lint

# Run tests
make test

# Run all checks
make check

# Development build with debugging
make dev

# Run development version
make run-dev
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Ways to Contribute
- 🐛 Report bugs
- 💡 Request features
- 📝 Improve documentation
- 🔧 Submit pull requests
- ⭐ Star the repository

## 🙏 Acknowledgments

QuikGit is built with amazing open-source libraries:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Harmonica](https://github.com/charmbracelet/harmonica) - Animations
- [go-github](https://github.com/google/go-github) - GitHub API client
- [go-git](https://github.com/go-git/go-git) - Git implementation
- [BubbleZone](https://github.com/lrstanley/bubblezone) - Mouse support

## 📊 Statistics

- **Languages Supported**: 15+
- **Package Managers**: 10+
- **Platforms**: Linux, macOS, Windows
- **Architectures**: AMD64, ARM64

## 🔮 Roadmap

- [ ] GitHub Enterprise support
- [ ] GitLab and Bitbucket integration
- [ ] Custom project templates
- [ ] Repository history and favorites
- [ ] Team collaboration features
- [ ] Plugin system
- [ ] Shell completion
- [ ] Man page documentation

## 🆘 Support

- 📖 [Documentation](https://github.com/lvcasx1/quikgit/wiki)
- 🐛 [Issue Tracker](https://github.com/lvcasx1/quikgit/issues)
- 💬 [Discussions](https://github.com/lvcasx1/quikgit/discussions)
- 📧 Email: support@quikgit.dev

## 📈 Metrics

![GitHub stars](https://img.shields.io/github/stars/lvcasx1/quikgit?style=social)
![GitHub forks](https://img.shields.io/github/forks/lvcasx1/quikgit?style=social)
![GitHub issues](https://img.shields.io/github/issues/lvcasx1/quikgit)
![GitHub pull requests](https://img.shields.io/github/issues-pr/lvcasx1/quikgit)

---

**Made with ❤️ by the QuikGit Team**

*Streamline your GitHub workflow, one repository at a time.*