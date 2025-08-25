package models

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/go-github/v66/github"
	"github.com/lvcasx1/quikgit/internal/auth"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
	"github.com/lvcasx1/quikgit/pkg/config"
)

type AppState int

const (
	StateLoading AppState = iota
	StateAuth
	StateMainMenu
	StateSearch
	StateResults
	StateCloning
	StateInstalling
	StateSettings
	StateHelp
)

type App struct {
	state        AppState
	width        int
	height       int
	config       *config.Config
	authManager  *auth.AuthManager
	githubClient *ghClient.Client
	ctx          context.Context
	cancel       context.CancelFunc

	// Sub-models
	authModel       *AuthModel
	searchModel     *SearchModel
	resultsModel    *ResultsModel
	cloningModel    *CloningModel
	installingModel *InstallingModel

	// UI state
	loading       bool
	error         error
	message       string
	menuSelection int // For main menu navigation
}

type AppMsg struct {
	Type    string
	Payload interface{}
}

func NewApp(cfg *config.Config) *App {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		state:       StateLoading,
		config:      cfg,
		authManager: auth.NewAuthManager(),
		ctx:         ctx,
		cancel:      cancel,
	}

	// Initialize sub-models
	app.authModel = NewAuthModel(app.authManager)
	app.searchModel = NewSearchModel()
	app.resultsModel = NewResultsModel()
	app.cloningModel = NewCloningModel()
	app.installingModel = NewInstallingModel()

	return app
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.checkAuthentication(),
		tea.WindowSize(),
	)
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Handle global messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.state != StateAuth && a.state != StateCloning && a.state != StateInstalling {
				a.cancel()
				return a, tea.Quit
			}
		case "esc":
			return a.handleBack()
		case "f1", "?":
			a.setState(StateHelp)
		}

	case AppMsg:
		return a.handleAppMsg(msg)
	}

	// Delegate to current state model
	switch a.state {
	case StateLoading:
		// Loading state handled by global messages

	case StateAuth:
		a.authModel, cmd = a.authModel.Update(msg)
		cmds = append(cmds, cmd)

	case StateMainMenu:
		// Handle main menu
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "up", "k":
				a.menuSelection--
				if a.menuSelection < 0 {
					a.menuSelection = 3 // 4 menu items (0-3)
				}
			case "down", "j":
				a.menuSelection++
				if a.menuSelection > 3 {
					a.menuSelection = 0
				}
			case "enter":
				a.selectMenuItem()
			case "1", "s":
				a.menuSelection = 0
				a.selectMenuItem()
			case "2", "c":
				a.menuSelection = 1
				a.selectMenuItem()
			case "3", "h":
				a.menuSelection = 2
				a.selectMenuItem()
			case "4", "config":
				a.menuSelection = 3
				a.selectMenuItem()
			}
		}

	case StateSearch:
		a.searchModel, cmd = a.searchModel.Update(msg)
		cmds = append(cmds, cmd)

	case StateResults:
		a.resultsModel, cmd = a.resultsModel.Update(msg)
		cmds = append(cmds, cmd)

	case StateCloning:
		a.cloningModel, cmd = a.cloningModel.Update(msg)
		cmds = append(cmds, cmd)

	case StateInstalling:
		a.installingModel, cmd = a.installingModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a *App) View() string {
	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	switch a.state {
	case StateLoading:
		return a.renderLoading()
	case StateAuth:
		return a.authModel.View()
	case StateMainMenu:
		return a.renderMainMenu()
	case StateSearch:
		return a.searchModel.View()
	case StateResults:
		return a.resultsModel.View()
	case StateCloning:
		return a.cloningModel.View()
	case StateInstalling:
		return a.installingModel.View()
	case StateSettings:
		return a.renderSettings()
	case StateHelp:
		return a.renderHelp()
	default:
		return "Unknown state"
	}
}

func (a *App) setState(state AppState) {
	a.state = state

	// Initialize state-specific data
	switch state {
	case StateAuth:
		// Auth model handles its own initialization
	case StateSearch:
		if a.githubClient != nil {
			a.searchModel.SetGitHubClient(a.githubClient)
		}
	case StateResults:
		// Results model gets data from search model
	}
}

func (a *App) handleBack() (tea.Model, tea.Cmd) {
	switch a.state {
	case StateSearch, StateSettings, StateHelp:
		a.setState(StateMainMenu)
	case StateResults:
		a.setState(StateSearch)
	case StateCloning, StateInstalling:
		// These states should complete before allowing back
		return a, nil
	default:
		// Can't go back from main menu or auth
		return a, nil
	}
	return a, nil
}

func (a *App) handleAppMsg(msg AppMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case "auth_required":
		a.setState(StateAuth)
		return a, a.authModel.Init()

	case "auth_success":
		user := msg.Payload.(*github.User)
		a.githubClient = ghClient.NewClient(a.authManager.GetClient())
		a.message = fmt.Sprintf("Welcome, %s!", user.GetLogin())
		a.setState(StateMainMenu)

	case "auth_failed":
		err := msg.Payload.(error)
		a.error = err
		a.setState(StateAuth)

	case "search_completed":
		results := msg.Payload.([]*ghClient.Repository)
		a.resultsModel.SetResults(results)
		a.setState(StateResults)

	case "clone_selected":
		repos := msg.Payload.([]*ghClient.Repository)
		a.cloningModel.SetRepositories(repos)
		a.setState(StateCloning)

	case "clone_completed":
		// Move to installation phase
		repos := msg.Payload.([]string)
		a.installingModel.SetRepositories(repos)
		a.setState(StateInstalling)

	case "install_completed":
		a.message = "Installation completed successfully!"
		a.setState(StateMainMenu)

	case "error":
		a.error = msg.Payload.(error)
	}

	return a, nil
}

func (a *App) checkAuthentication() tea.Cmd {
	return func() tea.Msg {
		// Try to load saved token
		if err := a.authManager.LoadToken(); err == nil && a.authManager.IsAuthenticated() {
			// Get user info
			if user, err := a.authManager.GetUser(); err == nil {
				return AppMsg{Type: "auth_success", Payload: user}
			}
		}

		// Need to authenticate
		return AppMsg{Type: "auth_required", Payload: nil}
	}
}

func (a *App) renderLoading() string {
	return fmt.Sprintf("Loading QuikGit...\n\nChecking authentication...")
}

func (a *App) renderMainMenu() string {
	menu := `
┌─────────────────────────────────────────┐
│              QuikGit v1.0               │
│         GitHub Repository Manager        │
└─────────────────────────────────────────┘

Welcome to QuikGit! What would you like to do?

`

	// Menu items with highlighting
	menuItems := []string{
		"Search repositories",
		"Quick clone",
		"View history",
		"Settings",
	}

	for i, item := range menuItems {
		prefix := "  "
		if i == a.menuSelection {
			prefix = "► "
		}
		menu += fmt.Sprintf("%s[%d] %s\n", prefix, i+1, item)
	}

	menu += `
Navigation:
  ↑/↓: Navigate    Enter: Select    q: Quit
  F1: Help         Esc: Back

`

	if a.message != "" {
		menu += fmt.Sprintf("\n✓ %s\n", a.message)
		a.message = "" // Clear message after showing
	}

	if a.error != nil {
		menu += fmt.Sprintf("\n✗ Error: %s\n", a.error)
		a.error = nil // Clear error after showing
	}

	return menu
}

func (a *App) selectMenuItem() {
	switch a.menuSelection {
	case 0: // Search repositories
		a.setState(StateSearch)
	case 1: // Quick clone
		a.handleQuickClone()
	case 2: // View history
		a.handleHistory()
	case 3: // Settings
		a.setState(StateSettings)
	}
}

func (a *App) handleQuickClone() {
	a.message = "Quick clone feature coming soon! Use Search repositories for now."
}

func (a *App) handleHistory() {
	a.message = "History feature coming soon! Check ~/.quikgit/ for cloned repositories."
}

func (a *App) renderSettings() string {
	return `
┌─────────────────────────────────────────┐
│               Settings                  │
└─────────────────────────────────────────┘

Configuration Options:

Clone Settings:
• Default directory: ` + a.config.Clone.DefaultPath + `
• Concurrent clones: ` + fmt.Sprintf("%d", a.config.Clone.Concurrent) + `
• Use SSH: ` + fmt.Sprintf("%t", a.config.GitHub.PreferSSH) + `

Installation Settings:
• Auto-install dependencies: ` + fmt.Sprintf("%t", a.config.Install.Enabled) + `
• Concurrent installs: ` + fmt.Sprintf("%d", a.config.Install.Concurrent) + `
• Skip on error: ` + fmt.Sprintf("%t", a.config.Install.SkipOnError) + `

UI Settings:
• Theme: ` + a.config.UI.Theme + `
• Show icons: ` + fmt.Sprintf("%t", a.config.UI.ShowIcons) + `
• Mouse support: ` + fmt.Sprintf("%t", a.config.UI.MouseSupport) + `

Note: Settings are currently read-only.
To modify settings, edit ~/.quikgit/config.yaml

Press Esc to return to the main menu.
`
}

func (a *App) renderHelp() string {
	return `
┌─────────────────────────────────────────┐
│                 Help                    │
└─────────────────────────────────────────┘

QuikGit - GitHub Repository Manager

Features:
• Search and browse GitHub repositories
• Clone multiple repositories at once  
• Automatic dependency installation
• Support for various project types

Keyboard Shortcuts:
  q, Ctrl+C: Quit application
  Esc: Go back to previous screen
  F1, ?: Show this help
  Tab: Switch between input fields
  Enter: Confirm/Select
  Space: Toggle selection (in lists)

Supported Project Types:
• Go (go.mod)
• Node.js (package.json) 
• Python (requirements.txt, Pipfile, pyproject.toml)
• Ruby (Gemfile)
• Rust (Cargo.toml)
• Java (pom.xml, build.gradle)
• C++ (CMakeLists.txt)
• C# (.csproj, .sln)
• Swift (Package.swift)
• PHP (composer.json)

Press Esc to return to the main menu.
`
}

func (a *App) Cleanup() {
	if a.cancel != nil {
		a.cancel()
	}
}
