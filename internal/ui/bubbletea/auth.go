package bubbletea

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type AuthStep int

const (
	AuthStepWelcome AuthStep = iota
	AuthStepPersonalToken
	AuthStepTokenInput
	AuthStepAuthenticating
	AuthStepSuccess
	AuthStepError
)

type AuthModel struct {
	app   *Application
	step  AuthStep
	input string
	err   error

	// Authentication state
	authInProgress bool

	// UI state
	blinking  bool
}

type AuthErrorMsg struct {
	Err error
}

type AuthSuccessMsg struct{}

type BlinkMsg struct{}

func NewAuthModel(app *Application) *AuthModel {
	return &AuthModel{
		app:  app,
		step: AuthStepWelcome,
	}
}

func (m *AuthModel) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return BlinkMsg{}
		}),
	)
}

func (m *AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case AuthErrorMsg:
		m.err = msg.Err
		m.step = AuthStepError
		m.authInProgress = false
		return m, nil
	case AuthSuccessMsg:
		m.step = AuthStepSuccess
		m.authInProgress = false
		return m, tea.Batch(
			m.app.UpdateAuthStatus(true),
			tea.Tick(2*time.Second, func(time.Time) tea.Msg {
				return tea.KeyMsg{Type: tea.KeyEsc}
			}),
		)
	case BlinkMsg:
		m.blinking = !m.blinking
		return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return BlinkMsg{}
		})
	}
	return m, nil
}

func (m *AuthModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.step {
	case AuthStepWelcome:
		switch msg.String() {
		case "esc", "q":
			return m, m.app.NavigateTo(StateMainMenu)
		case "enter", " ":
			m.step = AuthStepPersonalToken
		case "t": // Allow direct jump to token input
			m.step = AuthStepTokenInput
			m.input = ""
		}

	case AuthStepPersonalToken:
		switch msg.String() {
		case "esc":
			m.step = AuthStepWelcome
		case "enter", " ":
			m.step = AuthStepTokenInput
			m.input = ""
		}

	case AuthStepTokenInput:
		switch msg.String() {
		case "esc":
			m.step = AuthStepPersonalToken
			m.input = ""
		case "enter":
			if strings.TrimSpace(m.input) != "" {
				return m, m.authenticateWithToken(strings.TrimSpace(m.input))
			}
		case "backspace", "ctrl+h":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "ctrl+v":
			// Handle paste from clipboard - this is a placeholder
			// In a real implementation you'd use a clipboard library
			return m, nil
		default:
			// Allow all printable characters and common symbols
			if len(msg.Runes) == 1 {
				char := msg.Runes[0]
				if char >= 32 && char <= 126 { // Printable ASCII range
					m.input += string(char)
				}
			} else if msg.Type == tea.KeyRunes {
				// Handle pasted content
				m.input += string(msg.Runes)
			}
		}


	case AuthStepAuthenticating:
		switch msg.String() {
		case "esc", "q":
			m.authInProgress = false
			m.step = AuthStepWelcome
		}

	case AuthStepSuccess, AuthStepError:
		switch msg.String() {
		case "esc", "enter", " ", "q":
			return m, m.app.NavigateTo(StateMainMenu)
		}
	}

	return m, nil
}

func (m *AuthModel) authenticateWithToken(token string) tea.Cmd {
	m.step = AuthStepAuthenticating
	m.authInProgress = true

	return func() tea.Msg {
		os.Setenv("GITHUB_TOKEN", token)
		m.app.authManager.SetToken(token)

		if !m.app.authManager.IsAuthenticated() {
			return AuthErrorMsg{Err: fmt.Errorf("invalid token or authentication failed")}
		}

		if err := m.app.authManager.SaveToken(); err != nil {
			return AuthErrorMsg{Err: fmt.Errorf("failed to save token: %w", err)}
		}

		return AuthSuccessMsg{}
	}
}


func (m *AuthModel) View() string {
	width := m.app.width
	height := m.app.height - 3
	if width == 0 {
		width = 120
	}
	if height <= 0 {
		height = 30
	}

	var content string

	switch m.step {
	case AuthStepWelcome:
		content = m.renderWelcome(width)
	case AuthStepPersonalToken:
		content = m.renderPersonalTokenInfo(width)
	case AuthStepTokenInput:
		content = m.renderTokenInput(width)
	case AuthStepAuthenticating:
		content = m.renderAuthenticating(width)
	case AuthStepSuccess:
		content = m.renderSuccess(width)
	case AuthStepError:
		content = m.renderError(width)
	}

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *AuthModel) renderWelcome(screenWidth int) string {
	title := TitleStyle.Render("🔑 QuikGit GitHub Authentication")
	
	// Create main content card
	contentCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render(`Welcome to QuikGit authentication setup!

To use QuikGit, you'll need a GitHub personal access token.
This allows you to:
• Search and clone repositories
• Access private repositories
• Use all QuikGit features`)

	// Create instruction card
	instructionCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("39")).
		Render("Press Enter to continue • T to jump to token input • Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		contentCard,
		"",
		instructionCard,
	)
}


func (m *AuthModel) renderPersonalTokenInfo(screenWidth int) string {
	title := TitleStyle.Render("📝 Personal Access Token Setup")
	
	// Instructions card
	instructionsCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render(`To create a personal access token:

1. Open this URL in your browser:
   https://github.com/settings/tokens/new

2. Fill out the form:
   • Note: 'QuikGit CLI Access'
   • Expiration: Choose your preference  
   • Scopes: Check 'repo', 'read:user', and 'read:org'

3. Click 'Generate token' and copy it`)

	// URL highlight card
	urlCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Render("🔗 https://github.com/settings/tokens/new")

	// Navigation card
	navCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("241")).
		Foreground(lipgloss.Color("241")).
		Render("Press Enter to continue or Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		instructionsCard,
		"",
		urlCard,
		"",
		navCard,
	)
}

func (m *AuthModel) renderTokenInput(screenWidth int) string {
	title := TitleStyle.Render("Enter Your Token")
	
	cursor := ""
	if m.blinking {
		cursor = "▋"
	} else {
		cursor = " "
	}

	// Mask the token for security
	displayToken := strings.Repeat("*", len(m.input))
	
	// Instructions card
	instructionsCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render("Paste your GitHub personal access token below:")

	// Input field card
	inputCard := SelectedCardStyle.Copy().
		Width(screenWidth - 40).
		Render(fmt.Sprintf("Token: %s%s", displayToken, cursor))

	// Help card
	helpCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("39")).
		Render("💡 Tip: You can paste with Ctrl+V or right-click")

	// Navigation card
	navCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("241")).
		Foreground(lipgloss.Color("241")).
		Render("Press Enter to authenticate or Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		instructionsCard,
		"",
		inputCard,
		"",
		helpCard,
		"",
		navCard,
	)
}



func (m *AuthModel) renderAuthenticating(screenWidth int) string {
	title := TitleStyle.Render("🔄 Authenticating...")
	
	// Token verification
	statusCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render("Verifying your GitHub token...")

	navCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("241")).
		Foreground(lipgloss.Color("241")).
		Render("Press Esc to cancel")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		statusCard,
		"",
		navCard,
	)
}

func (m *AuthModel) renderSuccess(screenWidth int) string {
	title := SuccessStyle.Render("✅ Authentication Successful!")
	
	// Success message card
	successCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("46")).
		Foreground(lipgloss.Color("46")).
		Render("Great! You're now authenticated with GitHub.")

	// Features card
	featuresCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render(`You can now:
• Search for repositories
• Clone public and private repositories  
• Access all QuikGit features`)

	// Status card
	statusCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("39")).
		Render("Returning to main menu...")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		successCard,
		"",
		featuresCard,
		"",
		statusCard,
	)
}

func (m *AuthModel) renderError(screenWidth int) string {
	title := ErrorStyle.Render("❌ Authentication Failed")
	
	errorMsg := "An error occurred during authentication."
	if m.err != nil {
		errorMsg = m.err.Error()
	}
	
	// Error message card
	errorCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("196")).
		Foreground(lipgloss.Color("196")).
		Render(errorMsg)

	// Help card
	helpCard := CardStyle.Copy().
		Width(screenWidth - 40).
		Render("Please try again or choose a different authentication method.")

	// Navigation card
	navCard := BaseStyle.Copy().
		Width(screenWidth - 40).
		BorderForeground(lipgloss.Color("241")).
		Foreground(lipgloss.Color("241")).
		Render("Press Enter or Esc to go back")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		errorCard,
		"",
		helpCard,
		"",
		navCard,
	)
}
