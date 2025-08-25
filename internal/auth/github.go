package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

const (
	// NOTE: For a production deployment, replace this with your own GitHub OAuth App Client ID
	// Create one at: https://github.com/settings/applications/new
	// This allows users to authenticate via QR code without environment variables
	defaultClientID = "YOUR_GITHUB_OAUTH_CLIENT_ID_HERE"
	deviceCodeURL   = "https://github.com/login/device/code"
	tokenURL        = "https://github.com/login/oauth/access_token"
	scope           = "repo,read:user,read:org"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type AuthManager struct {
	client *github.Client
	token  string
}

func NewAuthManager() *AuthManager {
	return &AuthManager{}
}

func (a *AuthManager) InitiateDeviceFlow() (*DeviceCodeResponse, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		clientID = defaultClientID
	}

	// Check if using placeholder client ID
	if clientID == "YOUR_GITHUB_OAUTH_CLIENT_ID_HERE" {
		return nil, fmt.Errorf("OAuth client ID not configured. For QR code authentication, set GITHUB_CLIENT_ID or use GITHUB_TOKEN instead")
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("scope", scope)

	req, err := http.NewRequest("POST", deviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initiate device flow: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s. Please check your GITHUB_CLIENT_ID", resp.Status)
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device response: %w", err)
	}

	return &deviceResp, nil
}

func (a *AuthManager) PollForToken(deviceCode string, interval int) (*TokenResponse, error) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	if clientID == "" {
		clientID = defaultClientID
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(15 * time.Minute)

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("authentication timeout")
		case <-ticker.C:
			resp, err := http.PostForm(tokenURL, data)
			if err != nil {
				continue
			}

			var tokenResp TokenResponse
			if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			if tokenResp.Error == "authorization_pending" {
				continue
			}

			if tokenResp.Error == "slow_down" {
				ticker.Reset(time.Duration(interval+5) * time.Second)
				continue
			}

			if tokenResp.Error != "" {
				return nil, fmt.Errorf("authentication error: %s", tokenResp.Error)
			}

			if tokenResp.AccessToken != "" {
				return &tokenResp, nil
			}
		}
	}
}

func (a *AuthManager) SetToken(token string) {
	a.token = token
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	a.client = github.NewClient(tc)
}

func (a *AuthManager) GetClient() *github.Client {
	return a.client
}

func (a *AuthManager) IsAuthenticated() bool {
	if a.client == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := a.client.Users.Get(ctx, "")
	return err == nil
}

func (a *AuthManager) GetUser() (*github.User, error) {
	if a.client == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, _, err := a.client.Users.Get(ctx, "")
	return user, err
}

func (a *AuthManager) SaveToken() error {
	if a.token == "" {
		return fmt.Errorf("no token to save")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".quikgit")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tokenFile := filepath.Join(configDir, "token")

	// Encrypt the token with a simple base64 encoding (in production, use proper encryption)
	encryptedToken := base64.StdEncoding.EncodeToString([]byte(a.token))

	if err := os.WriteFile(tokenFile, []byte(encryptedToken), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (a *AuthManager) LoadToken() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	tokenFile := filepath.Join(homeDir, ".quikgit", "token")

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return fmt.Errorf("failed to read token file: %w", err)
	}

	// Decrypt the token (simple base64 decoding)
	token, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return fmt.Errorf("failed to decode token: %w", err)
	}

	a.SetToken(string(token))
	return nil
}

func generateRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
