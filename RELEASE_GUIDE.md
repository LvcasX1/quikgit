# Release Guide for QuikGit

This document explains how to create releases for QuikGit using the automated GoReleaser system.

## Quick Start (Basic Release)

**For your first release, no setup is required!** The basic configuration will:
- Build binaries for all platforms (Linux, macOS, Windows)
- Create DEB/RPM/APK packages
- Create a GitHub Release with downloadable assets
- Generate checksums

Package manager publishing (Homebrew, AUR, Scoop) is **disabled by default** and can be enabled later.

## Prerequisites for Basic Release

**Nothing!** The `GITHUB_TOKEN` is automatically provided by GitHub Actions.

Just create and push a tag:
```bash
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

## Optional: Enable Package Manager Publishing

These are **optional** and only needed if you want automatic publishing to package managers.

### Setup for Homebrew (macOS/Linux)

1. **Create the tap repository**:
   - Go to https://github.com/new
   - Create a repository named `homebrew-tap`
   - Make it public
   - Initialize with a README

2. **Create and add GitHub token**:
   - Go to: https://github.com/settings/tokens/new
   - Token name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Select scope: `repo` (full control)
   - Generate token
   - Copy the token

3. **Add secret to QuikGit repository**:
   - Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions
   - Click "New repository secret"
   - Name: `HOMEBREW_TAP_GITHUB_TOKEN`
   - Value: Paste the token
   - Click "Add secret"

4. **Enable in `.goreleaser.yaml`**:
   - Change `skip_upload: true` to `skip_upload: auto` in the `brews` section

### Setup for AUR (Arch Linux)

1. **Create AUR account**: https://aur.archlinux.org/register

2. **Generate SSH key**:
   ```bash
   ssh-keygen -t ed25519 -C "quikgit-aur-bot" -f ~/.ssh/aur_quikgit
   ```

3. **Add public key to AUR**:
   - Go to: https://aur.archlinux.org/account/
   - Click "My Account"
   - Add the content of `~/.ssh/aur_quikgit.pub` to "SSH Public Key"

4. **Add private key as GitHub secret**:
   - Copy the **entire content** of `~/.ssh/aur_quikgit` (the private key, not .pub)
   - Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions
   - Add secret named `AUR_SSH_PRIVATE_KEY`
   - Paste the entire private key content

5. **Enable in `.goreleaser.yaml`**:
   - Change `skip_upload: true` to `skip_upload: auto` in the `aurs` section

### Setup for Scoop (Windows)

1. **Create scoop-bucket repository**:
   - Go to https://github.com/new
   - Create repository named `scoop-bucket`
   - Make it public
   - Initialize with README

2. **Create and add GitHub token**:
   - Create token with `repo` scope (same as Homebrew)
   - Add as secret named `SCOOP_GITHUB_TOKEN`

3. **Enable in `.goreleaser.yaml`**:
   - Change `skip_upload: true` to `skip_upload: auto` in the `scoops` section

## Creating a Release

### 1. Update Version Information

Before creating a release, ensure:
- Update `CHANGELOG.md` with the new version's changes (if you maintain one)
- The code is in a releasable state
- All tests pass locally: `make test`

### 2. Create and Push a Git Tag

```bash
# Create a new tag (use semantic versioning)
git tag -a v1.0.2 -m "Release v1.0.2"

# Push the tag to GitHub
git push origin v1.0.2
```

### 3. Automated Release Process

Once you push the tag, GitHub Actions will automatically:

1. **Build** binaries for all platforms:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)

2. **Create packages**:
   - `.deb` for Debian/Ubuntu
   - `.rpm` for Fedora/RHEL/CentOS
   - `.apk` for Alpine Linux

3. **Publish to package managers** (if enabled):
   - Homebrew tap (`lvcasx1/homebrew-tap`)
   - AUR (`quikgit-bin`)
   - Scoop bucket (`lvcasx1/scoop-bucket`)

4. **Create GitHub Release** with:
   - Release notes (auto-generated from commits)
   - Downloadable archives (.tar.gz, .zip)
   - Package files (.deb, .rpm, .apk)
   - Checksums file

### 4. Monitor the Release

- Go to: https://github.com/lvcasx1/quikgit/actions
- Watch the "Release" workflow
- It typically takes 5-10 minutes to complete

## Installation Methods After Release

### Basic Installation (Always Available)

#### Direct Binary Download
Download from: https://github.com/lvcasx1/quikgit/releases/latest

#### Debian/Ubuntu
```bash
# Download .deb from GitHub releases
sudo dpkg -i quikgit_1.0.2_Linux_x86_64.deb
```

#### Fedora/RHEL/CentOS
```bash
# Download .rpm from GitHub releases
sudo rpm -i quikgit_1.0.2_Linux_x86_64.rpm
```

### Package Manager Installation (When Enabled)

#### Homebrew (macOS/Linux)
```bash
brew tap lvcasx1/tap
brew install quikgit
```

#### AUR (Arch Linux)
```bash
yay -S quikgit-bin
```

#### Scoop (Windows)
```bash
scoop bucket add lvcasx1 https://github.com/lvcasx1/scoop-bucket
scoop install quikgit
```

## Troubleshooting

### Release Workflow Fails

1. **Check GitHub Actions logs**: https://github.com/lvcasx1/quikgit/actions
2. Common issues:
   - Test failures during build
   - Invalid tag format (must be `vX.Y.Z`)
   - Go version mismatch

### Package Manager Publishing Fails

If you enabled package managers and they fail:

**Homebrew/Scoop Issues:**
- Ensure the tap/bucket repository exists
- Verify GitHub token has `repo` scope
- Check token is added as repository secret

**AUR Issues:**
- Verify SSH key is correctly configured
- Test SSH connection: `ssh -T aur@aur.archlinux.org`
- Check private key format (entire key, including headers)

## Testing Locally

To test the GoReleaser configuration without publishing:

```bash
# Install goreleaser
brew install goreleaser  # macOS
# or
go install github.com/goreleaser/goreleaser/v2@latest

# Test release (creates snapshot, doesn't publish)
goreleaser release --snapshot --clean

# Check the generated files in ./dist/
ls -la dist/
```

## Version Numbering

QuikGit follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Incompatible API changes
- **MINOR** version (v1.1.0): New functionality (backwards compatible)
- **PATCH** version (v1.0.1): Bug fixes (backwards compatible)

Examples:
- `v1.0.0` → First stable release
- `v1.1.0` → Added new features
- `v1.0.1` → Bug fixes only
- `v2.0.0` → Breaking changes

## Continuous Integration

Every push and pull request to `main` branch triggers:
- Automated tests on Linux, macOS, and Windows
- Code formatting checks
- Build verification for all platforms
- Coverage reports

See: `.github/workflows/ci.yml`

## Configuration Summary

| Feature | Status | Setup Required |
|---------|--------|----------------|
| GitHub Releases | ✅ Always enabled | None |
| Binary Archives | ✅ Always enabled | None |
| DEB Packages | ✅ Always enabled | None |
| RPM Packages | ✅ Always enabled | None |
| APK Packages | ✅ Always enabled | None |
| Homebrew | ⚠️ Disabled by default | Create tap repo + token |
| AUR | ⚠️ Disabled by default | SSH key setup |
| Scoop | ⚠️ Disabled by default | Create bucket repo + token |

## Quick Reference

| Command | Description |
|---------|-------------|
| `git tag -a v1.0.2 -m "Release v1.0.2"` | Create release tag |
| `git push origin v1.0.2` | Trigger release |
| `git tag -d v1.0.2` | Delete local tag |
| `git push origin :refs/tags/v1.0.2` | Delete remote tag |
| `goreleaser release --snapshot --clean` | Test locally |

## Support

If you encounter issues with the release process:
1. Check GitHub Actions logs
2. Verify tag format matches `vX.Y.Z`
3. Test locally with `goreleaser release --snapshot`
4. Review `.goreleaser.yaml` configuration
5. For package manager issues, ensure optional setup is complete
