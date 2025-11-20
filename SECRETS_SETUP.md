# Required GitHub Secrets Setup

You need to set up **3 secrets** in your GitHub repository before creating a release.

Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions

---

## Secret 1: HOMEBREW_TAP_GITHUB_TOKEN

### Purpose
Publishes releases to your Homebrew tap for macOS/Linux users.

### Setup Steps

#### 1. Create the homebrew-tap repository
- Go to: https://github.com/new
- Repository name: **homebrew-tap**
- Visibility: **Public**
- ✅ Check "Add a README file"
- Click "Create repository"

#### 2. Generate Personal Access Token
```
URL: https://github.com/settings/tokens/new
```

Fill in:
- **Note**: `Homebrew tap for quikgit releases`
- **Expiration**: No expiration
- **Select scopes**:
  - ✅ repo (Full control of private repositories)

Click "Generate token" and **COPY IT NOW**

#### 3. Add to GitHub Secrets
```
Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions
Click: "New repository secret"

Name:   HOMEBREW_TAP_GITHUB_TOKEN
Value:  [paste the token you copied]

Click "Add secret"
```

---

## Secret 2: SCOOP_GITHUB_TOKEN

### Purpose
Publishes releases to your Scoop bucket for Windows users.

### Setup Steps

#### 1. Create the scoop-bucket repository
- Go to: https://github.com/new
- Repository name: **scoop-bucket**
- Visibility: **Public**
- ✅ Check "Add a README file"
- Click "Create repository"

#### 2. Generate Personal Access Token
```
URL: https://github.com/settings/tokens/new
```

Fill in:
- **Note**: `Scoop bucket for quikgit releases`
- **Expiration**: No expiration
- **Select scopes**:
  - ✅ repo (Full control of private repositories)

Click "Generate token" and **COPY IT NOW**

*Note: You can reuse the same token from Homebrew if you want, or create a separate one.*

#### 3. Add to GitHub Secrets
```
Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions
Click: "New repository secret"

Name:   SCOOP_GITHUB_TOKEN
Value:  [paste the token you copied]

Click "Add secret"
```

---

## Secret 3: AUR_SSH_PRIVATE_KEY

### Purpose
Publishes releases to AUR (Arch User Repository) for Arch Linux users.

### Setup Steps

#### 1. Create AUR account
- Go to: https://aur.archlinux.org/register
- Create account and verify email

#### 2. Generate SSH key pair
```bash
ssh-keygen -t ed25519 -C "quikgit-aur-bot" -f ~/.ssh/aur_quikgit
```

Press Enter twice (no passphrase).

This creates two files:
- `~/.ssh/aur_quikgit` - Private key (secret)
- `~/.ssh/aur_quikgit.pub` - Public key (share with AUR)

#### 3. Add public key to AUR account
```bash
# Copy your public key
cat ~/.ssh/aur_quikgit.pub
```

- Go to: https://aur.archlinux.org/account/
- Paste the public key in "SSH Public Key" field
- Click "Update"

#### 4. Test SSH connection
```bash
ssh -T -i ~/.ssh/aur_quikgit aur@aur.archlinux.org
```

You should see: `Hi [username]! You've successfully authenticated...`

#### 5. Add private key to GitHub Secrets
```bash
# Display your ENTIRE private key (including headers)
cat ~/.ssh/aur_quikgit
```

You'll see something like:
```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
[many lines...]
-----END OPENSSH PRIVATE KEY-----
```

**Copy the ENTIRE output** (all lines from BEGIN to END)

```
Go to: https://github.com/lvcasx1/quikgit/settings/secrets/actions
Click: "New repository secret"

Name:   AUR_SSH_PRIVATE_KEY
Value:  [paste the ENTIRE private key including -----BEGIN and -----END lines]

Click "Add secret"
```

---

## Verification

After adding all 3 secrets, go to:
```
https://github.com/lvcasx1/quikgit/settings/secrets/actions
```

You should see:
- ✅ HOMEBREW_TAP_GITHUB_TOKEN
- ✅ SCOOP_GITHUB_TOKEN
- ✅ AUR_SSH_PRIVATE_KEY

(Plus GITHUB_TOKEN which is automatic)

---

## Summary: Exact Secret Names

Copy these names EXACTLY (they're case-sensitive):

```
HOMEBREW_TAP_GITHUB_TOKEN
SCOOP_GITHUB_TOKEN
AUR_SSH_PRIVATE_KEY
```

---

## Troubleshooting

### "Bad credentials" for Homebrew/Scoop
- Make sure the repository (`homebrew-tap` or `scoop-bucket`) exists
- Make sure it's **public**
- Verify the token has **repo** scope
- Try regenerating the token

### "Permission denied" for AUR
- Test SSH connection: `ssh -T -i ~/.ssh/aur_quikgit aur@aur.archlinux.org`
- Make sure you copied the **entire** private key including headers
- Verify public key is added to your AUR account settings

### "Repository not found"
- The repository names must be exactly: `homebrew-tap` and `scoop-bucket`
- Both must be **public** repositories
- Both must exist under your GitHub account (`lvcasx1`)

---

## After Setup

Once all 3 secrets are configured, you can create a release:

```bash
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

The workflow will automatically:
1. Build binaries for all platforms
2. Create DEB/RPM/APK packages
3. Publish to Homebrew tap
4. Publish to Scoop bucket
5. Publish to AUR
6. Create GitHub release with all assets

Users can then install via:
- **Homebrew**: `brew install lvcasx1/tap/quikgit`
- **Scoop**: `scoop bucket add lvcasx1 https://github.com/lvcasx1/scoop-bucket && scoop install quikgit`
- **AUR**: `yay -S quikgit-bin`
- **DEB**: Download from releases
- **RPM**: Download from releases
