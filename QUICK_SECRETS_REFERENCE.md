# Quick Secrets Reference

## The 3 Secrets You Need

Go to: **https://github.com/lvcasx1/quikgit/settings/secrets/actions**

### 1. HOMEBREW_TAP_GITHUB_TOKEN
```bash
# 1. Create repository
Go to: https://github.com/new
Name: homebrew-tap
Visibility: Public
✅ Add README

# 2. Generate token
Go to: https://github.com/settings/tokens/new
Scopes: ✅ repo
Generate and COPY

# 3. Add to secrets
Name: HOMEBREW_TAP_GITHUB_TOKEN
Value: [paste token]
```

### 2. SCOOP_GITHUB_TOKEN
```bash
# 1. Create repository
Go to: https://github.com/new
Name: scoop-bucket
Visibility: Public
✅ Add README

# 2. Generate token (or reuse HOMEBREW token)
Go to: https://github.com/settings/tokens/new
Scopes: ✅ repo
Generate and COPY

# 3. Add to secrets
Name: SCOOP_GITHUB_TOKEN
Value: [paste token]
```

### 3. AUR_SSH_PRIVATE_KEY
```bash
# 1. Register at AUR
https://aur.archlinux.org/register

# 2. Generate SSH key
ssh-keygen -t ed25519 -C "quikgit-aur" -f ~/.ssh/aur_quikgit
[Press Enter twice - no passphrase]

# 3. Add PUBLIC key to AUR
cat ~/.ssh/aur_quikgit.pub
# Copy output, paste at: https://aur.archlinux.org/account/

# 4. Test connection
ssh -T -i ~/.ssh/aur_quikgit aur@aur.archlinux.org
# Should see: "Hi [username]! You've successfully authenticated..."

# 5. Add PRIVATE key to GitHub secrets
cat ~/.ssh/aur_quikgit
# Copy ENTIRE output (including -----BEGIN and -----END)

Name: AUR_SSH_PRIVATE_KEY
Value: [paste entire private key]
```

## Verification Checklist

- [ ] homebrew-tap repository exists and is public
- [ ] scoop-bucket repository exists and is public
- [ ] AUR account created
- [ ] AUR SSH public key added to account
- [ ] AUR SSH connection tested successfully
- [ ] HOMEBREW_TAP_GITHUB_TOKEN secret added
- [ ] SCOOP_GITHUB_TOKEN secret added
- [ ] AUR_SSH_PRIVATE_KEY secret added

## Test Release

```bash
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2
```

Watch: https://github.com/lvcasx1/quikgit/actions

## Common Issues

**"Bad credentials" for Homebrew/Scoop**
→ Repository doesn't exist or isn't public
→ Token doesn't have "repo" scope

**"Permission denied" for AUR**
→ Wrong key copied (make sure it's the PRIVATE key, not .pub)
→ Public key not added to AUR account
→ Test with: `ssh -T -i ~/.ssh/aur_quikgit aur@aur.archlinux.org`

**Secret names are case-sensitive!**
→ Must be exactly: `HOMEBREW_TAP_GITHUB_TOKEN` (not homebrew_tap_github_token)
→ Must be exactly: `SCOOP_GITHUB_TOKEN`
→ Must be exactly: `AUR_SSH_PRIVATE_KEY`
