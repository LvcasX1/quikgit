# QuikGit Authentication Guide

QuikGit works out of the box with GitHub authentication! Simply launch the application and authenticate with the QR code.

## Out-of-the-Box Authentication (Recommended)

**No setup required!** QuikGit includes built-in GitHub OAuth authentication:

1. **Launch QuikGit**:
   ```bash
   quikgit
   ```

2. **Scan the QR code** or visit the displayed URL on your mobile device

3. **Authorize the application** and you're ready to go!

The application automatically saves your authentication for future sessions.

## Alternative: Personal Access Token

If you prefer using a personal access token:

1. **Create a Personal Access Token**:
   - Visit [GitHub Settings > Personal access tokens](https://github.com/settings/tokens/new)
   - Create a new token with `repo`, `read:user`, and `read:org` scopes
   - Copy the token

2. **Set Environment Variable**:
   ```bash
   export GITHUB_TOKEN=your_token_here
   ```

3. **Launch QuikGit**:
   ```bash
   quikgit
   ```

## Advanced: Custom OAuth App

For organizations wanting to use their own OAuth application:

### Step 1: Create a GitHub OAuth App

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the form:
   - **Application name**: "My QuikGit"
   - **Homepage URL**: "https://github.com/yourusername/quikgit"
   - **Authorization callback URL**: "http://localhost:8080"
4. Click "Register application"
5. Copy the "Client ID"

### Step 2: Set Environment Variable

```bash
export GITHUB_CLIENT_ID=your_client_id_here
```

### Step 3: Launch QuikGit

```bash
quikgit
```

QuikGit will display a QR code for authentication.

## Alternative: Custom OAuth App (Advanced)

If you want to use your own GitHub OAuth app instead of the built-in one:

### Step 1: Create a GitHub OAuth App

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the form:
   - Application name: "My QuikGit"
   - Homepage URL: "https://github.com/lvcasx1/quikgit"
   - Authorization callback URL: "http://localhost"
4. Click "Register application"
5. Copy the "Client ID"

### Step 2: Set the Environment Variable

```bash
export GITHUB_CLIENT_ID=your_client_id_here
```

### Step 3: Run QuikGit

```bash
./quikgit
```

QuikGit will use your custom OAuth app for authentication.

## Troubleshooting

### Authentication Issues

If you encounter authentication problems:

1. **Try restarting the application** - sometimes authentication flows can get stuck
2. **Check your internet connection** - OAuth requires network access
3. **Try clearing saved authentication** - delete `~/.quikgit/token` and restart
4. **Consider using a personal access token** as an alternative (see above)

### Token Validation Failed

If you get a token validation error:

1. Check that your token hasn't expired
2. Verify the token has the correct scopes (`repo`, `read:user`, `read:org`)
3. Make sure the token string is complete and correct

### OAuth App Issues

If OAuth authentication fails:

1. Verify your GitHub OAuth app is configured correctly
2. Check that the client ID is correct
3. Ensure your OAuth app allows device flow authentication

## Security Notes

- **Never commit tokens to version control**
- Store tokens in environment variables or secure credential managers
- Use tokens with minimal required scopes
- Regularly rotate your personal access tokens
- Consider using OAuth for shared/production environments

## Testing Authentication

To test if your authentication is working:

```bash
# This should show your GitHub username if authentication works
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

## Getting Help

If you're still having issues:

1. Check that you can access GitHub normally in your browser
2. Verify your network connection
3. Try creating a fresh personal access token
4. Check the [QuikGit issues](https://github.com/lvcasx1/quikgit/issues) for similar problems