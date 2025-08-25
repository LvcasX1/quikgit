package bubbletea

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuModel struct {
	app     *Application
	choices []menuChoice
	cursor  int
	// Pre-computed styles to avoid recreation
	titleStyle       lipgloss.Style
	instructionStyle lipgloss.Style
	statusStyle      lipgloss.Style
	// Pre-computed colors for performance
	selectedBorder lipgloss.Color
	normalBorder   lipgloss.Color
	// Pre-computed styles for status messages
	authSuccessStyle lipgloss.Style
	authErrorStyle   lipgloss.Style
	messageStyle     lipgloss.Style
	errorStyle       lipgloss.Style
	// Pre-computed styles for menu items
	selectedTitleStyle    lipgloss.Style
	unselectedTitleStyle  lipgloss.Style
	unavailableTitleStyle lipgloss.Style
	selectedDescStyle     lipgloss.Style
	unselectedDescStyle   lipgloss.Style
	unavailableDescStyle  lipgloss.Style
	// Cache for card base styles
	cardBaseStyle lipgloss.Style
	// Pre-rendered static content cache
	menuTitles       []string // Pre-rendered titles with icons
	menuDescriptions []string // Pre-rendered descriptions with unavailable text
	// Pre-computed card styles for each state (selected/unselected/unavailable)
	selectedCardStyle    lipgloss.Style
	unselectedCardStyle  lipgloss.Style
	unavailableCardStyle lipgloss.Style
	// Cache rendered content to avoid re-rendering on every frame
	cachedTitle        string
	cachedInstructions string
	initialized        bool // Track if we've rendered with real dimensions
	// Enhanced caching fields
	cachedStatus  string
	lastAuthState bool
	lastMessage   string
	lastError     string
	cachedMenu    string
	lastCursor    int
	lastWidth     int
	// Reusable slices to reduce allocations
	sections    []string
	menuItems   []string
	statusParts []string
}

type menuChoice struct {
	title       string
	description string
	icon        string
	action      AppState
	available   bool
}

func NewMainMenuModel(app *Application) *MainMenuModel {
	choices := []menuChoice{
		{
			title:       "Search Repositories",
			description: "Search GitHub repositories by query, language, and filters",
			icon:        "󰍉",
			action:      StateSearch,
			available:   app.githubClient != nil,
		},
		{
			title:       "Quick Clone",
			description: "Quickly clone a repository by owner/name or URL",
			icon:        "󰓅",
			action:      StateQuickClone,
			available:   true,
		},
		{
			title:       "Authentication",
			description: "Configure GitHub authentication token",
			icon:        "󰍁",
			action:      StateAuth,
			available:   true,
		},
	}

	model := &MainMenuModel{
		app:     app,
		choices: choices,
		cursor:  0,
	}

	// Pre-compute common styles once
	model.titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("206")).
		Background(lipgloss.Color("235")).
		Bold(true).
		Padding(3, 4).
		MarginBottom(2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("206")).
		Align(lipgloss.Center)

	model.instructionStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(2).
		Align(lipgloss.Center)

	model.statusStyle = lipgloss.NewStyle().
		Padding(1, 2).
		MarginBottom(2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Align(lipgloss.Center)

	// Pre-compute colors for card rendering
	model.selectedBorder = lipgloss.Color("205")
	model.normalBorder = lipgloss.Color("240")

	// Pre-compute all status message styles
	model.authSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Bold(true)

	model.authErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	model.messageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Italic(true)

	model.errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Bold(true)

	// Pre-compute all menu item text styles
	model.selectedTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	model.unselectedTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	model.unavailableTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	model.selectedDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Italic(true)

	model.unselectedDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true)

	model.unavailableDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("238")).
		Italic(true)

	// Pre-compute base card style (common properties)
	model.cardBaseStyle = lipgloss.NewStyle().
		Padding(1, 3).
		MarginBottom(1).
		BorderStyle(lipgloss.RoundedBorder())

	// Pre-render static menu content that doesn't change
	model.preRenderMenuContent()

	// Initialize reusable slices with proper capacity
	model.sections = make([]string, 0, 4)             // title, status, menu, instructions
	model.menuItems = make([]string, 0, len(choices)) // one per menu choice
	model.statusParts = make([]string, 0, 3)          // auth, message, error

	return model
}

// preRenderMenuContent pre-computes all static content that doesn't change between renders
func (m *MainMenuModel) preRenderMenuContent() {
	// Pre-render menu titles with icons (these never change)
	m.menuTitles = make([]string, len(m.choices))
	m.menuDescriptions = make([]string, len(m.choices))

	for i, choice := range m.choices {
		m.menuTitles[i] = fmt.Sprintf("%s %s", choice.icon, choice.title)

		desc := choice.description
		if !choice.available {
			desc += "\n󰀪 Requires authentication"
		}
		m.menuDescriptions[i] = desc
	}
}

// updateCardStyles updates card styles when screen size changes
func (m *MainMenuModel) updateCardStyles(cardWidth int) {
	// Only update if width actually changed to avoid unnecessary work
	if m.selectedCardStyle.GetWidth() == cardWidth {
		return
	}

	// Pre-compute all card style variants
	m.selectedCardStyle = m.cardBaseStyle.
		Width(cardWidth).
		BorderForeground(m.selectedBorder).
		Bold(true)

	m.unselectedCardStyle = m.cardBaseStyle.
		Width(cardWidth).
		BorderForeground(m.normalBorder).
		Bold(false)

	m.unavailableCardStyle = m.cardBaseStyle.
		Width(cardWidth).
		BorderForeground(m.normalBorder).
		Bold(false)
}

func (m *MainMenuModel) Init() tea.Cmd {
	return nil
}

func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor < len(m.choices) && m.choices[m.cursor].available {
				return m, m.app.NavigateTo(m.choices[m.cursor].action)
			}
		}
	}

	return m, nil
}

func (m *MainMenuModel) View() string {
	// Wait for actual screen dimensions to avoid visual jumps and performance issues
	width := m.app.width
	height := m.app.height - 3 // Leave space for footer

	// If we don't have real dimensions yet, return a simple loading state
	if width == 0 || height <= 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true).
			Align(lipgloss.Center).
			Render("󰊤 QuikGit")
	}

	// Reset and reuse slices to reduce allocations
	m.sections = m.sections[:0]

	// Cache title and instructions that don't change often
	if !m.initialized || width != m.lastWidth || m.cachedTitle == "" {
		titleText := "󰊤  Q U I K G I T  -  G I T H U B   R E P O S I T O R Y   M A N A G E R"
		m.cachedTitle = m.titleStyle.Width(width - 8).Render(titleText)
		m.cachedInstructions = m.instructionStyle.Width(width).Render(
			"Use ↑/↓ or j/k to navigate • Enter or Space to select • q to quit",
		)
		m.lastWidth = width
		m.initialized = true
	}

	m.sections = append(m.sections, m.cachedTitle)

	// Status info (cache when possible for performance)
	statusText := m.getCachedStatus(width)
	if statusText != "" {
		m.sections = append(m.sections, statusText)
	}

	// Menu options (cache when cursor hasn't changed)
	menuSection := m.getCachedMenu(width)
	m.sections = append(m.sections, menuSection)

	// Use cached instructions
	m.sections = append(m.sections, m.cachedInstructions)

	// Join vertically and fill the screen
	content := lipgloss.JoinVertical(lipgloss.Center, m.sections...)

	// Use full screen with proper centering
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *MainMenuModel) renderStatus(width int) string {
	// Reset and reuse slice to reduce allocations
	m.statusParts = m.statusParts[:0]

	// Authentication status using session-based check (no expensive auth validation)
	if m.app.isAuthenticated {
		m.statusParts = append(m.statusParts, m.authSuccessStyle.Render("󰄬 GitHub authenticated"))
	} else {
		m.statusParts = append(m.statusParts, m.authErrorStyle.Render("󰅖 GitHub authentication required"))
	}

	// Messages
	if m.app.message != "" {
		m.statusParts = append(m.statusParts, m.messageStyle.Render("󰋽 "+m.app.message))
	}

	if m.app.error != nil {
		m.statusParts = append(m.statusParts, m.errorStyle.Render("󰅖 "+m.app.error.Error()))
	}

	if len(m.statusParts) == 0 {
		return ""
	}

	return m.statusStyle.Width(width - 40).Render(strings.Join(m.statusParts, "\n"))
}

func (m *MainMenuModel) renderMenu(width int) string {
	// Pre-calculate card width once
	cardWidth := width - 40
	if cardWidth < 40 {
		cardWidth = 40
	}

	// Update card styles only if width changed
	m.updateCardStyles(cardWidth)

	// Reset and reuse slice to reduce allocations
	m.menuItems = m.menuItems[:0]

	for i, choice := range m.choices {
		// Select pre-computed card style based on state
		var cardStyle lipgloss.Style
		var titleStyle, descStyle lipgloss.Style

		if !choice.available {
			cardStyle = m.unavailableCardStyle
			titleStyle = m.unavailableTitleStyle
			descStyle = m.unavailableDescStyle
		} else if i == m.cursor {
			cardStyle = m.selectedCardStyle
			titleStyle = m.selectedTitleStyle
			descStyle = m.selectedDescStyle
		} else {
			cardStyle = m.unselectedCardStyle
			titleStyle = m.unselectedTitleStyle
			descStyle = m.unselectedDescStyle
		}

		// Use pre-rendered static content
		styledTitle := titleStyle.Render(m.menuTitles[i])
		styledDesc := descStyle.Render(m.menuDescriptions[i])

		// Single content join and render
		cardContent := lipgloss.JoinVertical(lipgloss.Left, styledTitle, styledDesc)
		m.menuItems = append(m.menuItems, cardStyle.Render(cardContent))
	}

	return lipgloss.JoinVertical(lipgloss.Center, m.menuItems...)
}

// getCachedStatus returns cached status or rebuilds if changed
func (m *MainMenuModel) getCachedStatus(width int) string {
	// Use session-based auth state (no repeated expensive auth checks)
	currentAuthState := m.app.isAuthenticated
	currentMessage := m.app.message
	currentError := ""
	if m.app.error != nil {
		currentError = m.app.error.Error()
	}

	// Only rebuild if something changed
	if m.cachedStatus == "" ||
		currentAuthState != m.lastAuthState ||
		currentMessage != m.lastMessage ||
		currentError != m.lastError ||
		width != m.lastWidth {

		m.cachedStatus = m.renderStatus(width)
		m.lastAuthState = currentAuthState
		m.lastMessage = currentMessage
		m.lastError = currentError
	}

	return m.cachedStatus
}

// getCachedMenu returns cached menu or rebuilds if cursor changed
func (m *MainMenuModel) getCachedMenu(width int) string {
	// Only rebuild if cursor changed or width changed
	if m.cachedMenu == "" || m.cursor != m.lastCursor || width != m.lastWidth {
		m.cachedMenu = m.renderMenu(width)
		m.lastCursor = m.cursor
	}

	return m.cachedMenu
}
