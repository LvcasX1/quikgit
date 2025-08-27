package bubbletea

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SplashModel struct {
	app            *Application
	validatingAuth bool
}

// SplashTimeoutMsg is sent when the splash screen timeout occurs
type SplashTimeoutMsg struct{}

// AuthValidationMsg contains the result of authentication validation
type AuthValidationMsg struct {
	Valid bool
	Error error
}

func NewSplashModel(app *Application) *SplashModel {
	return &SplashModel{
		app: app,
	}
}

func (m *SplashModel) Init() tea.Cmd {
	// Start authentication validation immediately
	return m.validateAuthentication()
}

func (m *SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow skipping splash screen with any key, but only if validation is complete
		if !m.validatingAuth {
			if m.app.isAuthenticated {
				return m, m.app.NavigateTo(StateMainMenu)
			} else {
				return m, m.app.NavigateTo(StateFirstStartup)
			}
		}
	case AuthValidationMsg:
		m.validatingAuth = false
		if msg.Valid {
			// Authentication is valid, go to main menu
			return m, m.app.NavigateTo(StateMainMenu)
		} else {
			// Authentication is invalid or missing, go to first startup
			m.app.isAuthenticated = false
			m.app.githubClient = nil
			return m, m.app.NavigateTo(StateFirstStartup)
		}
	case SplashTimeoutMsg:
		// Timeout during validation, proceed based on current auth state
		if m.app.isAuthenticated {
			return m, m.app.NavigateTo(StateMainMenu)
		} else {
			return m, m.app.NavigateTo(StateFirstStartup)
		}
	}
	return m, nil
}

// navigateToNextState determines where to go after splash based on auth status
func (m *SplashModel) navigateToNextState() tea.Cmd {
	if m.app.isAuthenticated {
		return m.app.NavigateTo(StateMainMenu)
	} else {
		return m.app.NavigateTo(StateAuth)
	}
}

func (m *SplashModel) View() string {
	// Get screen dimensions (if available)
	width := m.app.width
	height := m.app.height
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}

	// ASCII art for QuikGit
	asciiArt := `
 ██████╗ ██╗   ██╗██╗██╗  ██╗ ██████╗ ██╗████████╗
██╔═══██╗██║   ██║██║██║ ██╔╝██╔════╝ ██║╚══██╔══╝
██║   ██║██║   ██║██║█████╔╝ ██║  ███╗██║   ██║   
██║   ██║██║   ██║██║██╔═██╗ ██║   ██║██║   ██║   
╚██████╔╝╚██████╔╝██║██║  ██╗╚██████╔╝██║   ██║   
 ╚══▀▀═╝  ╚═════╝ ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚═╝   ╚═╝   
                                                   
            GitHub Repository Manager              `

	// Style the ASCII art
	artStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center)

	styledArt := artStyle.Render(asciiArt)

	// Loading indicator with authentication status
	loadingStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		MarginTop(3)

	var loading string
	if m.validatingAuth {
		if m.app.isAuthenticated {
			loading = loadingStyle.Render("Validating authentication...")
		} else {
			loading = loadingStyle.Render("Welcome to QuikGit! Initializing...")
		}
	} else {
		loading = loadingStyle.Render("Loading... (press any key to continue)")
	}

	// Combine all elements
	content := lipgloss.JoinVertical(lipgloss.Center, styledArt, loading)

	// Center on screen
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *SplashModel) validateAuthentication() tea.Cmd {
	m.validatingAuth = true

	return tea.Batch(
		// Start validation with minimum display time
		tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
			// Check if we have an auth manager and it claims to be authenticated
			if m.app.authManager == nil || !m.app.isAuthenticated {
				// For unauthenticated users, show splash for minimum time then go to first startup
				return AuthValidationMsg{Valid: false}
			}

			// For authenticated users, validate by making an actual API call with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			client := m.app.authManager.GetClient()
			if client == nil {
				return AuthValidationMsg{Valid: false}
			}

			_, _, err := client.Users.Get(ctx, "")
			if err != nil {
				return AuthValidationMsg{Valid: false, Error: err}
			}

			return AuthValidationMsg{Valid: true}
		}),
		// Set a timeout to prevent hanging
		tea.Tick(6*time.Second, func(t time.Time) tea.Msg {
			return SplashTimeoutMsg{}
		}),
	)
}
