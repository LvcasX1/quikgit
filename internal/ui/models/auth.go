package models

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lvcasx1/quikgit/internal/auth"
)

type AuthState int

const (
	AuthStateInitial AuthState = iota
	AuthStateGeneratingQR
	AuthStateWaitingForAuth
	AuthStateSuccess
	AuthStateError
)

type AuthModel struct {
	state       AuthState
	authManager *auth.AuthManager

	// QR Code data
	deviceCode *auth.DeviceCodeResponse
	qrCode     string

	// UI state
	loading bool
	error   error
	message string

	// Styles
	boxStyle         lipgloss.Style
	titleStyle       lipgloss.Style
	errorStyle       lipgloss.Style
	qrStyle          lipgloss.Style
	instructionStyle lipgloss.Style
}

type AuthSuccessMsg struct {
	User interface{}
}

type AuthErrorMsg struct {
	Error error
}

type QRGeneratedMsg struct {
	DeviceCode *auth.DeviceCodeResponse
	QRCode     string
}

func NewAuthModel(authManager *auth.AuthManager) *AuthModel {
	return &AuthModel{
		state:       AuthStateInitial,
		authManager: authManager,
		boxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(2),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		qrStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Align(lipgloss.Center),
		instructionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Align(lipgloss.Center).
			MarginTop(1),
	}
}

func (m *AuthModel) Init() tea.Cmd {
	return m.startDeviceFlow()
}

func (m *AuthModel) Update(msg tea.Msg) (*AuthModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			if m.state == AuthStateError {
				m.state = AuthStateInitial
				m.error = nil
				return m, m.startDeviceFlow()
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case QRGeneratedMsg:
		m.deviceCode = msg.DeviceCode
		m.qrCode = msg.QRCode
		m.state = AuthStateWaitingForAuth
		m.loading = false
		return m, m.pollForToken()

	case AuthSuccessMsg:
		m.state = AuthStateSuccess
		m.loading = false
		return m, func() tea.Msg {
			return AppMsg{Type: "auth_success", Payload: msg.User}
		}

	case AuthErrorMsg:
		m.state = AuthStateError
		m.error = msg.Error
		m.loading = false
		return m, nil
	}

	return m, nil
}

func (m *AuthModel) View() string {
	switch m.state {
	case AuthStateInitial, AuthStateGeneratingQR:
		return m.renderLoading()
	case AuthStateWaitingForAuth:
		return m.renderQRCode()
	case AuthStateSuccess:
		return m.renderSuccess()
	case AuthStateError:
		return m.renderError()
	default:
		return "Unknown auth state"
	}
}

func (m *AuthModel) startDeviceFlow() tea.Cmd {
	return func() tea.Msg {
		m.loading = true
		m.state = AuthStateGeneratingQR

		// First check if user has provided a personal access token via environment
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			m.authManager.SetToken(token)
			if err := m.authManager.SaveToken(); err != nil {
				return AuthErrorMsg{Error: fmt.Errorf("failed to save token: %w", err)}
			}

			user, err := m.authManager.GetUser()
			if err != nil {
				return AuthErrorMsg{Error: fmt.Errorf("failed to validate token: %w", err)}
			}

			return AuthSuccessMsg{User: user}
		}

		// Initiate OAuth device flow with embedded client ID
		deviceResp, err := m.authManager.InitiateDeviceFlow()
		if err != nil {
			// If OAuth fails, provide fallback instructions
			return AuthErrorMsg{Error: fmt.Errorf("Failed to initialize GitHub authentication: %w\n\nAlternative: Use a personal access token:\n1. Visit: https://github.com/settings/tokens/new\n2. Create token with 'repo' scope\n3. Set: export GITHUB_TOKEN=your_token\n4. Restart QuikGit", err)}
		}

		qrCode, err := auth.GenerateQRCode(deviceResp.VerificationURI + "?user_code=" + deviceResp.UserCode)
		if err != nil {
			return AuthErrorMsg{Error: fmt.Errorf("failed to generate QR code: %w", err)}
		}

		return QRGeneratedMsg{
			DeviceCode: deviceResp,
			QRCode:     auth.FormatQRForTerminal(qrCode),
		}
	}
}

func (m *AuthModel) pollForToken() tea.Cmd {
	return func() tea.Msg {
		tokenResp, err := m.authManager.PollForToken(m.deviceCode.DeviceCode, m.deviceCode.Interval)
		if err != nil {
			return AuthErrorMsg{Error: err}
		}

		// Set the token and create GitHub client
		m.authManager.SetToken(tokenResp.AccessToken)

		// Save the token
		if err := m.authManager.SaveToken(); err != nil {
			// Non-fatal error, just log it
			fmt.Printf("Warning: Could not save token: %v\n", err)
		}

		// Get user info to confirm authentication
		user, err := m.authManager.GetUser()
		if err != nil {
			return AuthErrorMsg{Error: fmt.Errorf("authentication successful but failed to get user info: %w", err)}
		}

		return AuthSuccessMsg{User: user}
	}
}

func (m *AuthModel) renderLoading() string {
	content := m.titleStyle.Render("GitHub Authentication") + "\n\n"

	if m.state == AuthStateGeneratingQR {
		content += "Generating QR code..."
	} else {
		content += "Initializing authentication..."
	}

	content += "\n\n" + m.instructionStyle.Render("Please wait...")

	return m.boxStyle.Render(content)
}

func (m *AuthModel) renderQRCode() string {
	var content strings.Builder

	content.WriteString(m.titleStyle.Render("GitHub Authentication"))
	content.WriteString("\n\n")

	content.WriteString("Scan the QR code with your mobile device or visit:\n")
	content.WriteString(fmt.Sprintf("🔗 %s\n\n", m.deviceCode.VerificationURI))
	content.WriteString(fmt.Sprintf("And enter code: %s\n\n", m.deviceCode.UserCode))

	content.WriteString(m.qrStyle.Render(m.qrCode))
	content.WriteString("\n")

	expiresIn := time.Duration(m.deviceCode.ExpiresIn) * time.Second
	content.WriteString(m.instructionStyle.Render(fmt.Sprintf("Code expires in: %v", expiresIn)))
	content.WriteString("\n")
	content.WriteString(m.instructionStyle.Render("Waiting for authentication..."))
	content.WriteString("\n\n")
	content.WriteString(m.instructionStyle.Render("Press Ctrl+C to cancel"))

	return m.boxStyle.Render(content.String())
}

func (m *AuthModel) renderSuccess() string {
	content := m.titleStyle.Render("Authentication Successful!") + "\n\n"
	content += "✅ Successfully authenticated with GitHub\n"
	content += "🚀 Loading QuikGit..."

	return m.boxStyle.Render(content)
}

func (m *AuthModel) renderError() string {
	content := m.titleStyle.Render("GitHub Authentication Setup") + "\n\n"

	// Split error message into lines for better formatting
	errorLines := strings.Split(m.error.Error(), "\n")
	for _, line := range errorLines {
		if strings.HasPrefix(line, "❌") || strings.Contains(line, "error:") {
			content += m.errorStyle.Render(line) + "\n"
		} else if strings.HasPrefix(line, "1.") || strings.HasPrefix(line, "2.") || strings.HasPrefix(line, "3.") || strings.HasPrefix(line, "4.") {
			content += m.instructionStyle.Render(line) + "\n"
		} else if line != "" {
			content += line + "\n"
		} else {
			content += "\n"
		}
	}

	content += "\n" + m.instructionStyle.Render("Press 'r' to retry or Ctrl+C to quit")

	return m.boxStyle.Render(content)
}
