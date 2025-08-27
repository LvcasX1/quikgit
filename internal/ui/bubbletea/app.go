package bubbletea

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lvcasx1/quikgit/internal/auth"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
	"github.com/lvcasx1/quikgit/pkg/config"
)

// AppState represents the current state of the application
type AppState int

const (
	StateSplash AppState = iota
	StateFirstStartup
	StateMainMenu
	StateAuth
	StateSearch
	StateSearchResults
	StateCloning
	StateInstalling
	StateQuickClone
)

// Application represents the main Bubble Tea application
type Application struct {
	// Core state
	state  AppState
	width  int
	height int
	ctx    context.Context

	// Configuration and managers
	config       *config.Config
	authManager  *auth.AuthManager
	githubClient *ghClient.Client

	// Session state (set once at startup to avoid repeated auth checks)
	isAuthenticated bool

	// UI components
	currentView tea.Model

	// Application data
	searchResults   []*ghClient.Repository
	selectedRepos   []*ghClient.Repository
	selectedIndices map[int]bool
	clonedPaths     []string // Paths of successfully cloned repositories

	// Messages and errors
	message string
	error   error
}

// NewApplication creates a new Bubble Tea application
func NewApplication(cfg *config.Config) *Application {
	ctx := context.Background()

	app := &Application{
		state:           StateSplash, // Start with splash screen
		width:           0,           // Let splash handle initial sizing
		height:          0,           // Let splash handle initial sizing
		ctx:             ctx,
		config:          cfg,
		selectedIndices: make(map[int]bool),
	}

	// Initialize auth manager and load existing token
	app.authManager = auth.NewAuthManager()
	if err := app.authManager.LoadToken(); err == nil && app.authManager.IsAuthenticated() {
		// Token loaded successfully and is valid
		app.githubClient = ghClient.NewClient(app.authManager.GetClient())
		app.isAuthenticated = true
	} else if envToken := os.Getenv("GITHUB_TOKEN"); envToken != "" {
		// Try to use token from environment variable
		app.authManager.SetToken(envToken)
		if app.authManager.IsAuthenticated() {
			app.githubClient = ghClient.NewClient(app.authManager.GetClient())
			app.isAuthenticated = true
		} else {
			app.isAuthenticated = false
		}
	} else {
		app.isAuthenticated = false
	}

	// Always start with splash screen regardless of authentication status
	app.currentView = NewSplashModel(app)

	return app
}

// Run starts the Bubble Tea application
func (a *Application) Run() error {
	program := tea.NewProgram(a, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := program.Run()
	return err
}

// Init implements tea.Model
func (a *Application) Init() tea.Cmd {
	return a.currentView.Init()
}

// Update implements tea.Model
func (a *Application) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

		// Update current view with new size
		if a.currentView != nil {
			var cmd tea.Cmd
			a.currentView, cmd = a.currentView.Update(msg)
			return a, cmd
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}

	case StateChangeMsg:
		return a.handleStateChange(msg)

	case AuthStatusChangedMsg:
		a.isAuthenticated = msg.IsAuthenticated
		if msg.IsAuthenticated {
			// Initialize GitHub client when authentication succeeds
			a.githubClient = ghClient.NewClient(a.authManager.GetClient())
			a.message = "Authentication successful!"
		} else {
			// Clear GitHub client when authentication is lost
			a.githubClient = nil
			a.message = ""
		}
	}

	// Delegate to current view
	if a.currentView != nil {
		var cmd tea.Cmd
		a.currentView, cmd = a.currentView.Update(msg)
		return a, cmd
	}

	return a, nil
}

// View implements tea.Model
func (a *Application) View() string {
	if a.currentView == nil {
		return "Loading..."
	}

	view := a.currentView.View()

	// Add footer with app info
	footer := a.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, view, footer)
}

// renderFooter creates a footer with app information
func (a *Application) renderFooter() string {
	width := a.width
	if width == 0 {
		width = 120
	}

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center).
		Width(width).
		MarginTop(1)

	return footerStyle.Render("QuikGit • Press q or Ctrl+C to quit")
}

// handleStateChange processes state change messages
func (a *Application) handleStateChange(msg StateChangeMsg) (tea.Model, tea.Cmd) {
	// Clear any previous errors when changing state
	if msg.NewState != a.state {
		a.error = nil
	}

	a.state = msg.NewState

	switch msg.NewState {
	case StateSplash:
		a.currentView = NewSplashModel(a)
	case StateFirstStartup:
		a.currentView = NewFirstStartupModel(a)
	case StateMainMenu:
		a.currentView = NewMainMenuModel(a)
	case StateAuth:
		a.currentView = NewAuthModel(a)
	case StateSearch:
		a.currentView = NewSearchModel(a)
	case StateSearchResults:
		a.currentView = NewSearchResultsModel(a)
	case StateCloning:
		a.currentView = NewCloningModel(a)
	case StateInstalling:
		a.currentView = NewInstallationModel(a)
	case StateQuickClone:
		a.currentView = NewQuickCloneModel(a)
	}

	if a.currentView != nil {
		return a, a.currentView.Init()
	}

	return a, nil
}

// StateChangeMsg is used to change the application state
type StateChangeMsg struct {
	NewState AppState
	Data     interface{}
}

// AuthStatusChangedMsg is sent when authentication status changes
type AuthStatusChangedMsg struct {
	IsAuthenticated bool
}

// NavigateTo changes the application state
func (a *Application) NavigateTo(state AppState) tea.Cmd {
	return func() tea.Msg {
		return StateChangeMsg{NewState: state}
	}
}

// UpdateAuthStatus notifies the app when authentication status changes
func (a *Application) UpdateAuthStatus(authenticated bool) tea.Cmd {
	return func() tea.Msg {
		return AuthStatusChangedMsg{IsAuthenticated: authenticated}
	}
}

// Common styles used across the application
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	// Card styles
	CardStyle = BaseStyle.Copy().
			MarginBottom(1).
			Width(80)

	SelectedCardStyle = CardStyle.Copy().
				BorderForeground(lipgloss.Color("205")).
				Bold(true)

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center).
			MarginBottom(1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)

	// Progress styles
	ProgressBarStyle = lipgloss.NewStyle().
				Width(50).
				MarginBottom(1)
)
