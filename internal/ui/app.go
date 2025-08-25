package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/lvcasx1/quikgit/internal/auth"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
	"github.com/lvcasx1/quikgit/pkg/config"
)

// Icon mappings from emojis to Nerd Font icons
var IconMap = map[string]string{
	"📁":  "󰉋", // folder icon
	"🔗":  "󰌷", // link icon
	"⚙️": "󰒓", // gear icon
	"📊":  "󰃰", // chart icon
	"✅":  "󰄬", // checkmark
	"❌":  "󰅖", // x mark
	"🚀":  "󰯁", // rocket
	"🔍":  "󰍉", // search icon
	"⚡":  "󰓃", // lightning bolt
	"📚":  "󰂺", // books icon
	"🏷️": "󰓹", // tag icon
	"ℹ️": "󰋼", // info icon
	"⚠️": "󰀪", // warning icon
	"🔄":  "󰑐", // refresh icon
	"⭐":  "󰓎", // star icon
	"🍴":  "󰃻", // fork icon
}

// AppState represents the current state of the application
type AppState int

const (
	StateLoading AppState = iota
	StateAuth
	StateMainMenu
	StateSearch
	StateResults
	StateQuickClone
	StateCloning
	StateInstalling
	StateSettings
	StateHelp
)

// Application is the main TView application
type Application struct {
	app          *tview.Application
	config       *config.Config
	authManager  *auth.AuthManager
	githubClient *ghClient.Client
	ctx          context.Context
	cancel       context.CancelFunc

	// UI components
	pages          *tview.Pages
	mainMenu       *tview.List
	searchForm     *tview.Form
	resultsList    *tview.List
	quickCloneForm *tview.Form
	progressModal  *tview.Modal
	settingsText   *tview.TextView
	helpText       *tview.TextView

	// State
	currentState    AppState
	selectedRepos   []*ghClient.Repository
	searchResults   []*ghClient.Repository
	selectedIndices map[int]bool // Track selected items in lists
	message         string
	error           error
}

// NewApplication creates a new TView-based application
func NewApplication(cfg *config.Config) *Application {
	ctx, cancel := context.WithCancel(context.Background())

	app := &Application{
		app:             tview.NewApplication(),
		config:          cfg,
		authManager:     auth.NewAuthManager(),
		ctx:             ctx,
		cancel:          cancel,
		currentState:    StateLoading,
		selectedIndices: make(map[int]bool),
	}

	app.initializeUI()
	return app
}

// setupColorfulTheme sets up a colorful theme with transparent backgrounds
func (a *Application) setupColorfulTheme() {
	// Set transparent/terminal default backgrounds
	a.pages.SetBackgroundColor(tcell.ColorDefault)
	
	// Set up colorful theme while maintaining transparency
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	tview.Styles.ContrastBackgroundColor = tcell.ColorDefault
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorDefault
	
	// Colorful borders and text
	tview.Styles.BorderColor = tcell.ColorDarkCyan
	tview.Styles.TitleColor = tcell.ColorLightGreen
	tview.Styles.GraphicsColor = tcell.ColorDarkCyan
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorLightBlue
	tview.Styles.TertiaryTextColor = tcell.ColorYellow
	tview.Styles.InverseTextColor = tcell.ColorBlack
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorLightGray
	
	// Set rounded border style
	tview.Borders.Horizontal = '─'
	tview.Borders.Vertical = '│'
	tview.Borders.TopLeft = '╭'
	tview.Borders.TopRight = '╮'
	tview.Borders.BottomLeft = '╰'
	tview.Borders.BottomRight = '╯'
}

// initializeUI sets up all the UI components
func (a *Application) initializeUI() {
	a.pages = tview.NewPages()

	// Set up colorful theme with terminal transparency
	a.setupColorfulTheme()

	a.app.SetRoot(a.pages, true)

	// Create all UI components
	a.createMainMenu()
	a.createSearchForm()
	a.createResultsList()
	a.createQuickCloneForm()
	a.createProgressModal()
	a.createSettingsText()
	a.createHelpText()

	// Set up global key bindings
	a.app.SetInputCapture(a.globalInputCapture)
}

// createMainMenu creates the main menu
func (a *Application) createMainMenu() {
	a.mainMenu = tview.NewList()
	a.mainMenu.SetBorder(true).SetTitle(" QuikGit v1.0.1 - GitHub Repository Manager ")

	// Add menu items with Nerd Font icons
	a.mainMenu.AddItem(IconMap["🔍"]+" Search repositories", "Search and browse GitHub repositories", '1', func() {
		a.showSearch()
	})
	a.mainMenu.AddItem(IconMap["⚡"]+" Quick clone", "Clone a repository instantly", '2', func() {
		a.showQuickClone()
	})
	a.mainMenu.AddItem(IconMap["📚"]+" View history", "View cloning history", '3', func() {
		a.showHistory()
	})
	a.mainMenu.AddItem(IconMap["⚙️"]+" Settings", "View and modify settings", '4', func() {
		a.showSettings()
	})

	// Style the main menu with colorful theme
	a.mainMenu.SetHighlightFullLine(true)
	a.mainMenu.SetSelectedTextColor(tcell.ColorBlack)
	a.mainMenu.SetSelectedBackgroundColor(tcell.ColorLightGreen)
	a.mainMenu.SetMainTextColor(tcell.ColorWhite)
	a.mainMenu.SetSecondaryTextColor(tcell.ColorLightBlue)
	a.mainMenu.SetBackgroundColor(tcell.ColorDefault)
}

// createSearchForm creates the search form
func (a *Application) createSearchForm() {
	a.searchForm = tview.NewForm()
	a.searchForm.SetBorder(true).SetTitle(" " + IconMap["🔍"] + " Search Repositories ")

	// Add input field with focus styling
	queryField := tview.NewInputField()
	queryField.SetLabel("Query").SetText("").SetFieldWidth(50)
	queryField.SetFieldBackgroundColor(tcell.ColorDefault)
	queryField.SetFieldTextColor(tcell.ColorDefault)
	// Add focus handlers for input highlighting
	queryField.SetFocusFunc(func() {
		queryField.SetFieldBackgroundColor(tcell.ColorDarkBlue)
		queryField.SetFieldTextColor(tcell.ColorWhite)
	})
	queryField.SetBlurFunc(func() {
		queryField.SetFieldBackgroundColor(tcell.ColorDefault)
		queryField.SetFieldTextColor(tcell.ColorDefault)
	})

	a.searchForm.AddFormItem(queryField)
	a.searchForm.AddDropDown("Language", []string{"Any", "Go", "JavaScript", "Python", "Java", "C++", "Rust", "TypeScript", "C#", "PHP", "Ruby"}, 0, nil)
	a.searchForm.AddDropDown("Sort by", []string{"Best match", "Stars", "Forks", "Updated", "Created"}, 0, nil)
	a.searchForm.AddCheckbox("Include forks", false, nil)
	a.searchForm.AddButton("Search", a.performSearch)
	a.searchForm.AddButton("Back", func() {
		a.showMainMenu()
	})

	a.searchForm.SetButtonsAlign(tview.AlignCenter)
}

// createResultsList creates the results list
func (a *Application) createResultsList() {
	a.resultsList = tview.NewList()
	a.resultsList.SetBorder(true).SetTitle(" Search Results ")
	a.resultsList.ShowSecondaryText(true)
	a.resultsList.SetHighlightFullLine(true)
}

// createQuickCloneForm creates the quick clone form
func (a *Application) createQuickCloneForm() {
	a.quickCloneForm = tview.NewForm()
	a.quickCloneForm.SetBorder(true).SetTitle(" " + IconMap["⚡"] + " Quick Clone ")

	// Add input field with focus styling
	repoField := tview.NewInputField()
	repoField.SetLabel("Repository").SetText("").SetFieldWidth(60)
	repoField.SetFieldBackgroundColor(tcell.ColorDefault)
	repoField.SetFieldTextColor(tcell.ColorDefault)
	// Add focus handlers for input highlighting
	repoField.SetFocusFunc(func() {
		repoField.SetFieldBackgroundColor(tcell.ColorDarkBlue)
		repoField.SetFieldTextColor(tcell.ColorWhite)
	})
	repoField.SetBlurFunc(func() {
		repoField.SetFieldBackgroundColor(tcell.ColorDefault)
		repoField.SetFieldTextColor(tcell.ColorDefault)
	})

	a.quickCloneForm.AddFormItem(repoField)
	a.quickCloneForm.AddTextView("Help", "Examples:\n  microsoft/vscode\n  https://github.com/golang/go", 60, 3, false, false)
	a.quickCloneForm.AddButton("Clone", a.performQuickCloneEnhanced)
	a.quickCloneForm.AddButton("Back", func() {
		a.showMainMenu()
	})

	a.quickCloneForm.SetButtonsAlign(tview.AlignCenter)
}

// createProgressModal creates the progress modal
func (a *Application) createProgressModal() {
	a.progressModal = tview.NewModal()
	a.progressModal.SetText("Processing...")
}

// createSettingsText creates the settings text view
func (a *Application) createSettingsText() {
	a.settingsText = tview.NewTextView()
	a.settingsText.SetBorder(true).SetTitle(" " + IconMap["⚙️"] + " Settings ")
	a.settingsText.SetDynamicColors(true)
	a.settingsText.SetWordWrap(true)
}

// createHelpText creates the help text view
func (a *Application) createHelpText() {
	a.helpText = tview.NewTextView()
	a.helpText.SetBorder(true).SetTitle(" " + IconMap["ℹ️"] + " Help ")
	a.helpText.SetDynamicColors(true)
	a.helpText.SetWordWrap(true)

	helpContent := `[-]QuikGit - GitHub Repository Manager[-]

[-]Features:[-]
• Search and browse GitHub repositories
• Clone multiple repositories at once  
• Automatic dependency installation
• Support for various project types

[-]Keyboard Shortcuts:[-]
  q, Ctrl+C: Quit application
  Esc, h: Go back to previous screen
  F1, ?: Show this help
  Tab: Switch between input fields
  Enter, l: Confirm/Select
  Space: Toggle selection (in lists)
  
[-]Vim Navigation (outside input fields):[-]
  h: Go back/left
  j: Move down
  k: Move up  
  l: Select/enter

[-]Input Fields:[-]
• Blue background indicates focused input field
• Normal text input works in focused fields
• Tab to navigate between form elements

[-]Supported Project Types:[-]
• Go (go.mod)
• Node.js (package.json) 
• Python (requirements.txt, Pipfile, pyproject.toml)
• Ruby (Gemfile)
• Rust (Cargo.toml)
• Java (pom.xml, build.gradle)
• C++ (CMakeLists.txt)
• C# (.csproj, .sln)
• Swift (Package.swift)
• PHP (composer.json)`

	a.helpText.SetText(helpContent)
}

// globalInputCapture handles global key bindings
func (a *Application) globalInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlC:
		a.quit()
		return nil
	case tcell.KeyEscape:
		a.handleBack()
		return nil
	case tcell.KeyF1:
		a.showHelp()
		return nil
	}

	// Check if we're currently focused on an input field - if so, don't intercept hjkl
	focusedPrimitive := a.app.GetFocus()
	if focusedPrimitive != nil {
		switch focusedPrimitive.(type) {
		case *tview.InputField:
			// Let input fields handle their own key events, only intercept non-text keys
			if event.Key() == tcell.KeyRune {
				switch event.Rune() {
				case 'q', 'Q':
					if a.currentState == StateMainMenu {
						a.quit()
						return nil
					}
				case '?':
					a.showHelp()
					return nil
				}
			}
			return event
		}
	}

	// Handle character keys for non-input components
	if event.Key() == tcell.KeyRune {
		switch event.Rune() {
		case 'q', 'Q':
			if a.currentState == StateMainMenu {
				a.quit()
				return nil
			}
		case '?':
			a.showHelp()
			return nil
		// Vim navigation keys (only when not in input field)
		case 'h':
			// Left/Back - same as Escape
			a.handleBack()
			return nil
		case 'j':
			// Down - delegate to focused component
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			// Up - delegate to focused component
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'l':
			// Right/Select - same as Enter
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		}
	}

	return event
}

// Run starts the application
func (a *Application) Run() error {
	defer a.cleanup()

	// Check authentication first
	go a.checkAuthentication()

	// Show loading screen initially
	a.showLoading()

	return a.app.Run()
}

// checkAuthentication checks if user is authenticated
func (a *Application) checkAuthentication() {
	time.Sleep(500 * time.Millisecond) // Brief loading delay

	if err := a.authManager.LoadToken(); err == nil && a.authManager.IsAuthenticated() {
		a.githubClient = ghClient.NewClient(a.authManager.GetClient())
		a.app.QueueUpdateDraw(func() {
			a.message = "Welcome! Authentication successful."
			a.showMainMenu()
		})
		return
	}

	// Need authentication
	a.app.QueueUpdateDraw(func() {
		a.showAuth()
	})
}

// showLoading displays the loading screen
func (a *Application) showLoading() {
	loadingText := tview.NewTextView()
	loadingText.SetText("Loading QuikGit...\n\nChecking authentication...")
	loadingText.SetTextAlign(tview.AlignCenter)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(loadingText, 5, 0, true).
			AddItem(nil, 0, 1, false), 40, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddPage("loading", flex, true, true)
	a.currentState = StateLoading
}

// showAuth displays the authentication screen
func (a *Application) showAuth() {
	authText := tview.NewTextView()
	authText.SetText("Authentication required.\n\nPress Enter to authenticate with GitHub.")
	authText.SetTextAlign(tview.AlignCenter)

	authText.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			go a.performAuth()
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(authText, 5, 0, true).
			AddItem(nil, 0, 1, false), 40, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddPage("auth", flex, true, true)
	a.currentState = StateAuth
}

// showMainMenu displays the main menu
func (a *Application) showMainMenu() {
	// Center the main menu
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.mainMenu, 12, 0, true).
			AddItem(a.createMessageView(), 3, 0, false).
			AddItem(nil, 0, 1, false), 60, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	// Switch to main page, creating it if it doesn't exist
	if !a.pages.HasPage("main") {
		a.pages.AddPage("main", flex, true, false)
	}
	a.pages.SwitchToPage("main")
	a.currentState = StateMainMenu
	a.app.SetFocus(a.mainMenu)
}

// createMessageView creates a view for messages
func (a *Application) createMessageView() *tview.TextView {
	msgView := tview.NewTextView()
	msgView.SetTextAlign(tview.AlignCenter)
	msgView.SetDynamicColors(true)

	var content string
	if a.message != "" {
		content = fmt.Sprintf("[-]"+IconMap["✅"]+" %s[-]", a.message)
		a.message = "" // Clear message after showing
	}
	if a.error != nil {
		content = fmt.Sprintf("[-]"+IconMap["❌"]+" Error: %s[-]", a.error.Error())
		a.error = nil // Clear error after showing
	}

	msgView.SetText(content)
	return msgView
}

// showSearch displays the search form
func (a *Application) showSearch() {
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.searchForm, 12, 0, true).
			AddItem(nil, 0, 1, false), 80, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("search", flex, true)
	a.currentState = StateSearch
	a.app.SetFocus(a.searchForm)

	// Focus the first input field (Query field) by default
	a.searchForm.SetFocus(0)
}

// showQuickClone displays the quick clone form
func (a *Application) showQuickClone() {
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.quickCloneForm, 12, 0, true).
			AddItem(nil, 0, 1, false), 80, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("quickclone", flex, true)
	a.currentState = StateQuickClone
	a.app.SetFocus(a.quickCloneForm)

	// Focus the first input field (Repository field) by default
	a.quickCloneForm.SetFocus(0)
}

// showSettings displays the settings
func (a *Application) showSettings() {
	content := fmt.Sprintf(`[-]Configuration Options:[-]

[-]Clone Settings:[-]
• Default directory: %s
• Use current directory: %t
• Create owner subdirectories: %t
• Concurrent clones: %d
• Use SSH: %t

[-]Installation Settings:[-]
• Auto-install dependencies: %t
• Concurrent installs: %d
• Skip on error: %t

[-]UI Settings:[-]
• Theme: %s
• Show icons: %t
• Mouse support: %t

[-]Note:[-] Settings are currently read-only.
To modify settings, edit ~/.quikgit/config.yaml

Press Esc/h to return to the main menu.`,
		a.config.Clone.DefaultPath,
		a.config.Clone.UseCurrentDir,
		a.config.Clone.CreateSubdirs,
		a.config.Clone.Concurrent,
		a.config.GitHub.PreferSSH,
		a.config.Install.Enabled,
		a.config.Install.Concurrent,
		a.config.Install.SkipOnError,
		a.config.UI.Theme,
		a.config.UI.ShowIcons,
		a.config.UI.MouseSupport)

	a.settingsText.SetText(content)

	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.settingsText, 0, 1, true).
			AddItem(nil, 0, 1, false), 80, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("settings", flex, true)
	a.currentState = StateSettings
	a.app.SetFocus(a.settingsText)
}

// showHelp displays the help screen
func (a *Application) showHelp() {
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(a.helpText, 0, 1, true).
			AddItem(nil, 0, 1, false), 80, 0, true).
		AddItem(nil, 0, 1, false)

	// Set background color to terminal default
	flex.SetBackgroundColor(tcell.ColorDefault)

	a.pages.AddAndSwitchToPage("help", flex, true)
	a.currentState = StateHelp
	a.app.SetFocus(a.helpText)
}

// showHistory handles history view (placeholder)
func (a *Application) showHistory() {
	a.message = "History feature coming soon! Check ~/.quikgit/ for cloned repositories."
	a.showMainMenu()
}

// handleBack handles the back navigation
func (a *Application) handleBack() {
	switch a.currentState {
	case StateSearch, StateSettings, StateHelp, StateQuickClone:
		a.showMainMenu()
	case StateResults:
		a.showSearch()
	case StateCloning:
		// Allow cancel during cloning
		if a.pages.HasPage("cloning") {
			// This will be handled by the cloning progress view input capture
		}
	case StateInstalling:
		// Allow cancel during installation  
		if a.pages.HasPage("installing") {
			// This will be handled by the installation progress view input capture
		}
	default:
		// Can't go back from main menu or auth, or from loading
		if a.currentState != StateMainMenu && a.currentState != StateAuth && a.currentState != StateLoading {
			a.showMainMenu()
		}
	}
}

// performAuth handles GitHub authentication
func (a *Application) performAuth() {
	// TODO: Implement GitHub OAuth flow
	a.app.QueueUpdateDraw(func() {
		a.message = "Authentication not yet implemented in TView version"
		a.showMainMenu()
	})
}

// performSearch handles repository search
func (a *Application) performSearch() {
	a.performSearchEnhanced()
}

// quit exits the application
func (a *Application) quit() {
	a.app.Stop()
}

// cleanup performs cleanup when the application exits
func (a *Application) cleanup() {
	if a.cancel != nil {
		a.cancel()
	}
}
