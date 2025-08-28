package bubbletea

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type authStep int

const (
	stepTokenInput authStep = iota
	stepValidating
	stepSuccess
	stepError
)

type AuthModel struct {
	app            *Application
	step           authStep
	tokenInput     textinput.Model
	status         string
	error          error
	tokenSubmitted bool
	secureMode     bool // When true, ESC is disabled and only Ctrl+C or q can exit
}

// AuthStatusMsg represents authentication status updates
type AuthStatusMsg struct {
	Success bool
	Error   error
	Message string
}

type AuthErrorMsg struct {
	Err error
}

type AuthSuccessMsg struct{}

type BlinkMsg struct{}

func NewAuthModel(app *Application) *AuthModel {
	return newAuthModel(app, false)
}

func NewSecureAuthModel(app *Application) *AuthModel {
	return newAuthModel(app, true)
}

func newAuthModel(app *Application, secure bool) *AuthModel {
	// Create token input with obfuscation
	tokenInput := textinput.New()
	tokenInput.Placeholder = "ghp_your_github_token_here"
	tokenInput.CharLimit = 100
	tokenInput.Width = 50
	tokenInput.EchoMode = textinput.EchoPassword // Hide characters with asterisks
	tokenInput.EchoCharacter = '*'
	tokenInput.Focus()

	return &AuthModel{
		app:        app,
		step:       stepTokenInput,
		tokenInput: tokenInput,
		secureMode: secure,
	}
}

func (m *AuthModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *AuthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Always allow these to quit the application
			return m, tea.Quit
		case "esc":
			// Only allow ESC navigation if not in secure mode
			if !m.secureMode {
				return m, m.app.NavigateTo(StateMainMenu)
			}
			// In secure mode, ESC is ignored
		case "enter":
			if m.step == stepTokenInput && !m.tokenSubmitted {
				return m.submitToken()
			} else if m.step == stepError {
				// Reset to token input on error
				m.step = stepTokenInput
				m.error = nil
				m.tokenSubmitted = false
				return m, nil
			}
		}

	case StateChangeMsg:
		// Handle direct state change messages in secure mode
		if m.secureMode && msg.NewState == StateMainMenu {
			return m, m.app.NavigateTo(StateMainMenu)
		}

	case AuthStatusMsg:
		if msg.Success {
			m.step = stepSuccess
			m.status = msg.Message
			// Update app authentication state
			m.app.isAuthenticated = true
			// Navigate based on mode
			if m.secureMode {
				// In secure mode, navigate directly to main menu after success
				return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return StateChangeMsg{NewState: StateMainMenu}
				})
			} else {
				// In normal mode, use ESC navigation
				return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
					return tea.KeyMsg{Type: tea.KeyEsc}
				})
			}
		} else {
			m.step = stepError
			m.error = msg.Error
			m.tokenSubmitted = false
		}
	}

	// Update text input
	var cmd tea.Cmd
	m.tokenInput, cmd = m.tokenInput.Update(msg)
	return m, cmd
}

func (m *AuthModel) submitToken() (tea.Model, tea.Cmd) {
	token := strings.TrimSpace(m.tokenInput.Value())
	if token == "" {
		m.error = fmt.Errorf("token cannot be empty")
		return m, nil
	}

	m.tokenSubmitted = true
	m.step = stepValidating
	m.status = "Validating token..."

	return m, func() tea.Msg {
		// Check if token from environment variable exists and prefer it
		if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
			token = envToken
		}

		// Set and validate the token
		m.app.authManager.SetToken(token)

		// Test the token by making an API call
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := m.app.authManager.GetClient()
		if client == nil {
			return AuthStatusMsg{Success: false, Error: fmt.Errorf("failed to create GitHub client")}
		}

		_, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return AuthStatusMsg{Success: false, Error: fmt.Errorf("invalid token: %w", err)}
		}

		// Save the token
		if err := m.app.authManager.SaveToken(); err != nil {
			return AuthStatusMsg{Success: false, Error: fmt.Errorf("failed to save token: %w", err)}
		}

		// Create GitHub client
		m.app.githubClient = ghClient.NewClient(m.app.authManager.GetClient())

		return AuthStatusMsg{Success: true, Message: "Token authenticated successfully!"}
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
	case stepTokenInput:
		content = m.renderTokenInput(width)
	case stepValidating:
		content = m.renderValidating(width)
	case stepSuccess:
		content = m.renderSuccess(width)
	case stepError:
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

func (m *AuthModel) renderTokenInput(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	var titleText string
	if m.secureMode {
		titleText = "󰌾 GitHub Token Required"
	} else {
		titleText = "󰌆 GitHub Personal Access Token"
	}
	title := titleStyle.Render(titleText)

	// Token input
	inputStyle := lipgloss.NewStyle().
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		MarginBottom(2)

	input := inputStyle.Render(m.tokenInput.View())

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Italic(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	// Create clickable link using ANSI escape sequences
	clickableURL := "\033]8;;https://github.com/settings/tokens/new\033\\🔗 https://github.com/settings/tokens/new\033]8;;\033\\"

	var instructionText string
	if m.secureMode {
		instructionText = "Authentication is required to continue.\n\n" +
			"Create a token at:\n" +
			clickableURL + "\n\n" +
			"Required scopes: repo, read:user, read:org"
	} else {
		instructionText = "Create a token at:\n" +
			clickableURL + "\n\n" +
			"Required scopes: repo, read:user, read:org"
	}
	
	instructions := instructionStyle.Render(instructionText)

	// Error display
	var errorSection string
	if m.error != nil {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1)
		errorSection = errorStyle.Render("󰅖 " + m.error.Error())
	}

	var footerText string
	if m.secureMode {
		footerText = "Enter to submit • Ctrl+C or q to quit"
	} else {
		footerText = "Enter to submit • Esc to go back"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render(footerText)

	sections := []string{title, input, instructions}
	if errorSection != "" {
		sections = append(sections, errorSection)
	}
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}

func (m *AuthModel) renderValidating(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰔟 Validating Token...")

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		MarginBottom(2)

	status := statusStyle.Render(m.status)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Please wait...")

	return lipgloss.JoinVertical(lipgloss.Center, title, status, footer)
}

func (m *AuthModel) renderSuccess(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰄬 Authentication Successful!")

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Align(lipgloss.Center).
		MarginBottom(2)

	status := statusStyle.Render(m.status)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Returning to main menu...")

	return lipgloss.JoinVertical(lipgloss.Center, title, status, footer)
}

func (m *AuthModel) renderError(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰅖 Authentication Failed")

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Align(lipgloss.Center).
		MarginBottom(2)

	errorMsg := errorStyle.Render(m.error.Error())

	var footerText string
	if m.secureMode {
		footerText = "Enter to try again • Ctrl+C or q to quit"
	} else {
		footerText = "Enter to try again • Esc to go back"
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render(footerText)

	return lipgloss.JoinVertical(lipgloss.Center, title, errorMsg, footer)
}
