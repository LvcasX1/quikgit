#!/bin/bash

# QuikGit OAuth Setup Script
# This script helps users set up GitHub OAuth for QR code authentication

echo "🚀 QuikGit OAuth Setup"
echo "======================"
echo ""
echo "This script will help you set up GitHub OAuth authentication for QuikGit."
echo "This enables QR code login without needing personal access tokens."
echo ""

# Check if GITHUB_CLIENT_ID is already set
if [ -n "$GITHUB_CLIENT_ID" ]; then
    echo "✅ GITHUB_CLIENT_ID is already set: $GITHUB_CLIENT_ID"
    echo ""
    echo "To test QuikGit with OAuth:"
    echo "  ./build/quikgit"
    exit 0
fi

# Check if GITHUB_TOKEN is set
if [ -n "$GITHUB_TOKEN" ]; then
    echo "✅ GITHUB_TOKEN is already set"
    echo "QuikGit will use your personal access token for authentication."
    echo ""
    echo "To test QuikGit:"
    echo "  ./build/quikgit"
    exit 0
fi

echo "Choose your authentication method:"
echo ""
echo "1. 🔑 Personal Access Token (Fastest - 2 minutes)"
echo "2. 📱 OAuth with QR Code (More setup - 5 minutes)"
echo ""
read -p "Enter your choice (1 or 2): " choice

case $choice in
    1)
        echo ""
        echo "📝 Setting up Personal Access Token..."
        echo ""
        echo "1. Open this URL in your browser:"
        echo "   https://github.com/settings/tokens/new"
        echo ""
        echo "2. Fill out the form:"
        echo "   - Note: 'QuikGit CLI Access'"
        echo "   - Expiration: Choose your preference"
        echo "   - Scopes: Check 'repo', 'read:user', and 'read:org'"
        echo ""
        echo "3. Click 'Generate token' and copy it"
        echo ""
        read -p "4. Paste your token here: " token
        
        if [ -n "$token" ]; then
            echo ""
            echo "Setting GITHUB_TOKEN..."
            echo "export GITHUB_TOKEN=$token" >> ~/.bashrc
            echo "export GITHUB_TOKEN=$token" >> ~/.zshrc 2>/dev/null || true
            export GITHUB_TOKEN=$token
            echo ""
            echo "✅ Token configured! Run QuikGit now:"
            echo "   source ~/.bashrc  # or restart your terminal"
            echo "   ./build/quikgit"
        else
            echo "❌ No token provided. Please run this script again."
            exit 1
        fi
        ;;
    2)
        echo ""
        echo "📱 Setting up OAuth Application..."
        echo ""
        echo "1. Open this URL in your browser:"
        echo "   https://github.com/settings/applications/new"
        echo ""
        echo "2. Fill out the OAuth App form:"
        echo "   - Application name: 'My QuikGit'"
        echo "   - Homepage URL: 'https://github.com/lvcasx1/quikgit'"
        echo "   - Authorization callback URL: 'http://localhost:8080'"
        echo "   - Enable Device Flow: ✅ (Important!)"
        echo ""
        echo "3. Click 'Register application'"
        echo "4. Copy the 'Client ID' from the app settings page"
        echo ""
        read -p "5. Paste your Client ID here: " client_id
        
        if [ -n "$client_id" ]; then
            echo ""
            echo "Setting GITHUB_CLIENT_ID..."
            echo "export GITHUB_CLIENT_ID=$client_id" >> ~/.bashrc
            echo "export GITHUB_CLIENT_ID=$client_id" >> ~/.zshrc 2>/dev/null || true
            export GITHUB_CLIENT_ID=$client_id
            echo ""
            echo "✅ OAuth configured! Run QuikGit now:"
            echo "   source ~/.bashrc  # or restart your terminal" 
            echo "   ./build/quikgit"
        else
            echo "❌ No client ID provided. Please run this script again."
            exit 1
        fi
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again and choose 1 or 2."
        exit 1
        ;;
esac

echo ""
echo "🎉 Setup complete! QuikGit is ready to use."