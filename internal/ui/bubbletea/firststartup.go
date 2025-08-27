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

	"github.com/lvcasx1/quikgit/internal/auth"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type FirstStartupModel struct {
	app            *Application
	step           firstStartupStep
	tokenInput     textinput.Model
	status         string
	error          error
	tokenSubmitted bool
}

type firstStartupStep int

const (
	fsStepWelcome firstStartupStep = iota
	fsStepTokenInput
	fsStepValidating
	fsStepSuccess
	fsStepError
)

// FirstStartupStatusMsg represents first startup authentication status updates
type FirstStartupStatusMsg struct {
	Success bool
	Error   error
	Message string
}

func NewFirstStartupModel(app *Application) *FirstStartupModel {
	// Create token input with obfuscation
	tokenInput := textinput.New()
	tokenInput.Placeholder = "ghp_your_github_token_here"
	tokenInput.CharLimit = 100
	tokenInput.Width = 50
	tokenInput.EchoMode = textinput.EchoPassword // Hide characters with asterisks
	tokenInput.EchoCharacter = '*'
	tokenInput.Focus()

	return &FirstStartupModel{
		app:        app,
		step:       fsStepWelcome,
		tokenInput: tokenInput,
	}
}

func (m *FirstStartupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *FirstStartupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.step == fsStepWelcome {
				m.step = fsStepTokenInput
				return m, nil
			} else if m.step == fsStepTokenInput && !m.tokenSubmitted {
				return m.submitToken()
			} else if m.step == fsStepError {
				// Reset to token input on error
				m.step = fsStepTokenInput
				m.error = nil
				m.tokenSubmitted = false
				m.tokenInput.SetValue("")
				return m, nil
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}

	case FirstStartupStatusMsg:
		if msg.Success {
			m.step = fsStepSuccess
			m.status = msg.Message
			// Update app authentication state
			m.app.isAuthenticated = true
			// Navigate to main menu after a short delay
			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return StateChangeMsg{NewState: StateMainMenu}
			})
		} else {
			m.step = fsStepError
			m.error = msg.Error
			m.tokenSubmitted = false
		}

	case StateChangeMsg:
		// Handle navigation to main menu
		if msg.NewState == StateMainMenu {
			return m, m.app.NavigateTo(StateMainMenu)
		}
	}

	// Update text input only when in token input step
	if m.step == fsStepTokenInput {
		var cmd tea.Cmd
		m.tokenInput, cmd = m.tokenInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *FirstStartupModel) submitToken() (tea.Model, tea.Cmd) {
	token := strings.TrimSpace(m.tokenInput.Value())
	if token == "" {
		m.error = fmt.Errorf("token cannot be empty")
		m.step = fsStepError
		return m, nil
	}

	m.tokenSubmitted = true
	m.step = fsStepValidating
	m.status = "Validating token..."

	return m, func() tea.Msg {
		// Check if token from environment variable exists and prefer it
		if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
			token = envToken
		}

		// Initialize auth manager if not already done
		if m.app.authManager == nil {
			m.app.authManager = auth.NewAuthManager()
		}

		// Set and validate the token
		m.app.authManager.SetToken(token)

		// Test the token by making an API call
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client := m.app.authManager.GetClient()
		if client == nil {
			return FirstStartupStatusMsg{Success: false, Error: fmt.Errorf("failed to create GitHub client")}
		}

		_, _, err := client.Users.Get(ctx, "")
		if err != nil {
			return FirstStartupStatusMsg{Success: false, Error: fmt.Errorf("invalid token: %w", err)}
		}

		// Save the token
		if err := m.app.authManager.SaveToken(); err != nil {
			return FirstStartupStatusMsg{Success: false, Error: fmt.Errorf("failed to save token: %w", err)}
		}

		// Create GitHub client
		m.app.githubClient = ghClient.NewClient(m.app.authManager.GetClient())

		return FirstStartupStatusMsg{Success: true, Message: "Welcome to QuikGit! Authentication successful!"}
	}
}

func (m *FirstStartupModel) View() string {
	// Use full screen dimensions with fallback
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
	case fsStepWelcome:
		content = m.renderWelcome(width)
	case fsStepTokenInput:
		content = m.renderTokenInput(width)
	case fsStepValidating:
		content = m.renderValidating(width)
	case fsStepSuccess:
		content = m.renderSuccess(width)
	case fsStepError:
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

func (m *FirstStartupModel) renderWelcome(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰊤 Welcome to QuikGit!")

	// Welcome message
	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Align(lipgloss.Center).
		MarginBottom(3).
		Width(width - 20)

	welcomeText := welcomeStyle.Render(
		"QuikGit is your GitHub Repository Manager\n\n" +
			"To get started, you'll need to authenticate with GitHub\n" +
			"using a Personal Access Token.")

	// Instructions
	instructionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	instructions := instructionStyle.Render("Press Enter to continue")

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Press q or Ctrl+C to quit")

	return lipgloss.JoinVertical(lipgloss.Center, title, welcomeText, instructions, footer)
}

func (m *FirstStartupModel) renderTokenInput(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰌆 GitHub Personal Access Token")

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

	instructions := instructionStyle.Render(
		"Create a token at:\n" +
			clickableURL + "\n\n" +
			"Required scopes: repo, read:user, read:org")

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

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Enter to submit • q or Ctrl+C to quit")

	sections := []string{title, input, instructions}
	if errorSection != "" {
		sections = append(sections, errorSection)
	}
	sections = append(sections, footer)

	return lipgloss.JoinVertical(lipgloss.Center, sections...)
}

func (m *FirstStartupModel) renderValidating(width int) string {
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

func (m *FirstStartupModel) renderSuccess(width int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true).
		Align(lipgloss.Center).
		MarginBottom(2)

	title := titleStyle.Render("󰄬 Setup Complete!")

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Align(lipgloss.Center).
		MarginBottom(2)

	status := statusStyle.Render(m.status)

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Taking you to the main menu...")

	return lipgloss.JoinVertical(lipgloss.Center, title, status, footer)
}

func (m *FirstStartupModel) renderError(width int) string {
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

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Align(lipgloss.Center).
		Render("Enter to try again • q or Ctrl+C to quit")

	return lipgloss.JoinVertical(lipgloss.Center, title, errorMsg, footer)
}
