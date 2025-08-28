package bubbletea

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/v66/github"

	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type SearchModel struct {
	app             *Application
	queryInput      textinput.Model
	languageOptions []string
	languageCursor  int
	sortOptions     []string
	sortCursor      int
	scopeOptions    []string // New: organization vs all
	scopeCursor     int      // New: current scope selection
	includeForks    bool
	focusedField    int // 0=query, 1=language, 2=sort, 3=scope, 4=forks
	searching       bool
	searchError     error
}

func NewSearchModel(app *Application) *SearchModel {
	// Create text input for query with conservative width
	queryInput := textinput.New()
	queryInput.Placeholder = "Search repositories..."
	queryInput.Focus()
	queryInput.CharLimit = 100
	queryInput.Width = 30 // Smaller initial width

	// Restore previous query if available
	if app.searchSession != nil && app.searchSession.LastQuery != "" {
		queryInput.SetValue(app.searchSession.LastQuery)
		queryInput.SetCursor(len(app.searchSession.LastQuery))
	}

	model := &SearchModel{
		app:             app,
		queryInput:      queryInput,
		languageOptions: []string{"Any", "Go", "JavaScript", "TypeScript", "Python", "Java", "C++", "C", "Rust", "Ruby", "PHP"},
		sortOptions:     []string{"Best match", "Stars", "Forks", "Updated", "Created"},
		scopeOptions:    []string{"Organization", "All"},
		focusedField:    0,
	}

	// Restore session state if available
	if app.searchSession != nil {
		model.languageCursor = app.searchSession.LanguageCursor
		model.sortCursor = app.searchSession.SortCursor
		model.scopeCursor = app.searchSession.ScopeCursor
		model.includeForks = app.searchSession.IncludeForks
	} else {
		// Default values
		model.languageCursor = 0
		model.sortCursor = 0
		model.scopeCursor = 0
		model.includeForks = false
	}

	return model
}

func (m *SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Save current search state before navigating away
			m.saveSearchState()
			return m, m.app.NavigateTo(StateMainMenu)
		case "tab", "shift+tab":
			return m.handleTabNavigation(msg.String() == "shift+tab")
		case "enter":
			// Allow search from any field as long as query is not empty
			if strings.TrimSpace(m.queryInput.Value()) != "" {
				return m.performSearch()
			}
		case "up", "down":
			return m.handleArrowNavigation(msg.String() == "up")
		case " ":
			if m.focusedField == 4 { // Updated field index for forks
				m.includeForks = !m.includeForks
				// Save state immediately when forks toggle changes
				m.saveSearchState()
			}
		}
	case SearchResultMsg:
		m.searching = false
		if msg.Error != nil {
			m.searchError = msg.Error
		} else {
			m.app.searchResults = msg.Results
			return m, m.app.NavigateTo(StateSearchResults)
		}
	}

	// Update text input
	var cmd tea.Cmd
	m.queryInput, cmd = m.queryInput.Update(msg)
	return m, cmd
}

func (m *SearchModel) View() string {
	// Use full screen dimensions with fallback
	width := m.app.width
	height := m.app.height - 3
	if width == 0 {
		width = 120
	}
	if height <= 0 {
		height = 30
	}

	var sections []string

	// Title
	titleStyle := TitleStyle.Copy().Width(width - 20)
	title := titleStyle.Render("󰍉 Search GitHub Repositories")
	sections = append(sections, title)

	// Search form
	formSection := m.renderForm(width)
	sections = append(sections, formSection)

	// Error display
	if m.searchError != nil {
		errorStyle := ErrorStyle.Copy().
			Width(width - 20).
			Align(lipgloss.Center).
			MarginTop(1)
		sections = append(sections, errorStyle.Render("󰅖 "+m.searchError.Error()))
	}

	// Instructions
	instructions := m.renderInstructions(width)
	sections = append(sections, instructions)

	// Center content
	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m *SearchModel) renderForm(width int) string {
	// Calculate responsive form width
	formWidth := width - 40
	if formWidth < 60 {
		formWidth = 60
	}

	// Update query input width to fit form - account for more padding
	inputWidth := formWidth - 20 // Account for form padding, border padding, and margins
	if inputWidth < 15 {
		inputWidth = 15
	}
	if inputWidth > 50 { // Cap maximum width
		inputWidth = 50
	}
	m.queryInput.Width = inputWidth

	formStyle := lipgloss.NewStyle().
		Padding(2, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(formWidth)

	var formFields []string

	// Query input with highlighted border
	queryStyle := lipgloss.NewStyle().MarginBottom(1)
	if m.focusedField == 0 {
		queryLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("► Query:")
		// Highlighted input container with rounded border
		inputContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			MarginTop(1).
			Render(m.queryInput.View())
		formFields = append(formFields, queryStyle.Render(queryLabel+"\n"+inputContainer))
	} else {
		queryLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Query:")
		// Normal input container with subtle border
		inputContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			MarginTop(1).
			Render(m.queryInput.View())
		formFields = append(formFields, queryStyle.Render(queryLabel+"\n"+inputContainer))
	}

	// Language selection
	languageField := m.renderLanguageField()
	formFields = append(formFields, languageField)

	// Sort selection
	sortField := m.renderSortField()
	formFields = append(formFields, sortField)

	// Scope selection (new field)
	scopeField := m.renderScopeField()
	formFields = append(formFields, scopeField)

	// Include forks checkbox
	forksField := m.renderForksField()
	formFields = append(formFields, forksField)

	// Search button
	if strings.TrimSpace(m.queryInput.Value()) != "" {
		var searchButton string
		if m.searching {
			searchButton = lipgloss.NewStyle().
				Foreground(lipgloss.Color("39")).
				Bold(true).
				Align(lipgloss.Center).
				MarginTop(1).
				Render("󰔟 Searching...")
			formFields = append(formFields, searchButton)
		} else {
			searchButton = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				Align(lipgloss.Center).
				MarginTop(1).
				Render("󰊤 Press Enter to Search")
			formFields = append(formFields, searchButton)
		}
	}

	formContent := lipgloss.JoinVertical(lipgloss.Left, formFields...)
	return formStyle.Render(formContent)
}

func (m *SearchModel) renderLanguageField() string {
	fieldStyle := lipgloss.NewStyle().MarginBottom(1)

	var label string
	if m.focusedField == 1 {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("► Language:")
	} else {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Language:")
	}

	// Show current selection and options if focused
	currentLang := m.languageOptions[m.languageCursor]
	if m.focusedField == 1 {
		options := fmt.Sprintf("< %s > (↑/↓ to change)", currentLang)
		// Highlighted container with rounded border
		optionsContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Foreground(lipgloss.Color("39")).
			Italic(true).
			MarginTop(1).
			Render(options)
		return fieldStyle.Render(label + "\n" + optionsContainer)
	} else {
		// Normal container with subtle border
		valueContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("246")).
			MarginTop(1).
			Render(currentLang)
		return fieldStyle.Render(label + "\n" + valueContainer)
	}
}

func (m *SearchModel) renderSortField() string {
	fieldStyle := lipgloss.NewStyle().MarginBottom(1)

	var label string
	if m.focusedField == 2 {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("► Sort by:")
	} else {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Sort by:")
	}

	currentSort := m.sortOptions[m.sortCursor]
	if m.focusedField == 2 {
		options := fmt.Sprintf("< %s > (↑/↓ to change)", currentSort)
		// Highlighted container with rounded border
		optionsContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Foreground(lipgloss.Color("39")).
			Italic(true).
			MarginTop(1).
			Render(options)
		return fieldStyle.Render(label + "\n" + optionsContainer)
	} else {
		// Normal container with subtle border
		valueContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("246")).
			MarginTop(1).
			Render(currentSort)
		return fieldStyle.Render(label + "\n" + valueContainer)
	}
}

func (m *SearchModel) renderScopeField() string {
	fieldStyle := lipgloss.NewStyle().MarginBottom(1)

	var label string
	if m.focusedField == 3 {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("► Search scope:")
	} else {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Search scope:")
	}

	currentScope := m.scopeOptions[m.scopeCursor]
	if m.focusedField == 3 {
		options := fmt.Sprintf("< %s > (↑/↓ to change)", currentScope)
		// Highlighted container with rounded border
		optionsContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Foreground(lipgloss.Color("39")).
			Italic(true).
			MarginTop(1).
			Render(options)
		return fieldStyle.Render(label + "\n" + optionsContainer)
	} else {
		// Normal container with subtle border
		valueContainer := lipgloss.NewStyle().
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Foreground(lipgloss.Color("246")).
			MarginTop(1).
			Render(currentScope)
		return fieldStyle.Render(label + "\n" + valueContainer)
	}
}

func (m *SearchModel) renderForksField() string {
	fieldStyle := lipgloss.NewStyle().MarginBottom(1)

	var label string
	if m.focusedField == 4 { // Updated field index
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("► Include forks:")
	} else {
		label = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Include forks:")
	}

	var checkbox string
	if m.includeForks {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("󰄬 Yes")
	} else {
		checkbox = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("󰄱 No")
	}

	if m.focusedField == 4 { // Updated field index
		checkbox += lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Italic(true).
			Render(" (Space to toggle)")
	}

	return fieldStyle.Render(label + "\n" + checkbox)
}

func (m *SearchModel) renderInstructions(width int) string {
	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(2).
		Width(width).
		Align(lipgloss.Center)

	return instructionsStyle.Render(
		"Tab to navigate fields • Enter to search from any field • Esc to go back",
	)
}

func (m *SearchModel) handleTabNavigation(backwards bool) (tea.Model, tea.Cmd) {
	if backwards {
		m.focusedField--
		if m.focusedField < 0 {
			m.focusedField = 4 // Updated for new scope field
		}
	} else {
		m.focusedField++
		if m.focusedField > 4 { // Updated for new scope field
			m.focusedField = 0
		}
	}

	// Update text input focus
	if m.focusedField == 0 {
		m.queryInput.Focus()
	} else {
		m.queryInput.Blur()
	}

	return m, nil
}

func (m *SearchModel) handleArrowNavigation(up bool) (tea.Model, tea.Cmd) {
	switch m.focusedField {
	case 1: // Language
		if up {
			if m.languageCursor > 0 {
				m.languageCursor--
			}
		} else {
			if m.languageCursor < len(m.languageOptions)-1 {
				m.languageCursor++
			}
		}
	case 2: // Sort
		if up {
			if m.sortCursor > 0 {
				m.sortCursor--
			}
		} else {
			if m.sortCursor < len(m.sortOptions)-1 {
				m.sortCursor++
			}
		}
	case 3: // Scope (new field)
		if up {
			if m.scopeCursor > 0 {
				m.scopeCursor--
			}
		} else {
			if m.scopeCursor < len(m.scopeOptions)-1 {
				m.scopeCursor++
			}
		}
	}
	
	// Save state immediately when selections change
	m.saveSearchState()
	return m, nil
}

func (m *SearchModel) performSearch() (tea.Model, tea.Cmd) {
	if m.searching {
		return m, nil
	}

	m.searching = true
	m.searchError = nil

	// Save current search state to session
	m.saveSearchState()

	query := strings.TrimSpace(m.queryInput.Value())
	language := m.languageOptions[m.languageCursor]
	sortBy := m.sortOptions[m.sortCursor]
	scope := m.scopeOptions[m.scopeCursor]

	return m, m.searchRepositories(query, language, sortBy, scope, m.includeForks)
}

// saveSearchState saves the current search filters to the application session
func (m *SearchModel) saveSearchState() {
	if m.app.searchSession != nil {
		m.app.searchSession.LastQuery = strings.TrimSpace(m.queryInput.Value())
		m.app.searchSession.LanguageCursor = m.languageCursor
		m.app.searchSession.SortCursor = m.sortCursor
		m.app.searchSession.ScopeCursor = m.scopeCursor
		m.app.searchSession.IncludeForks = m.includeForks
	}
}

func (m *SearchModel) searchRepositories(query, language, sortBy, scope string, includeForks bool) tea.Cmd {
	return func() tea.Msg {
		if scope == "Organization" {
			// For organization scope, search each user/org separately and combine results
			return m.searchOrganizationRepositories(query, language, sortBy, includeForks)
		}

		// Create context with timeout for better responsiveness
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Build search query for "All" scope
		searchQuery := query
		if language != "Any" {
			searchQuery += " language:" + strings.ToLower(language)
		}
		if !includeForks {
			searchQuery += " fork:false"
		}

		// Map sort options
		var sort string
		switch sortBy {
		case "Stars":
			sort = "stars"
		case "Forks":
			sort = "forks"
		case "Updated":
			sort = "updated"
		case "Created":
			sort = "created"
		default:
			sort = "" // best match
		}

		opts := ghClient.SearchOptions{
			Query: searchQuery,
			Sort:  sort,
			Order: "desc",
			Page:  1,
			Limit: 20, // Reduced from 30 for faster response
		}

		// Perform search immediately without artificial delay
		repos, _, err := m.app.githubClient.SearchRepositories(ctx, opts)

		return SearchResultMsg{
			Results: repos,
			Error:   err,
		}
	}
}

func (m *SearchModel) searchOrganizationRepositories(query, language, sortBy string, includeForks bool) tea.Msg {
	// Create context with timeout for better responsiveness
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get authenticated user and organizations concurrently
	var wg sync.WaitGroup
	var user *github.User
	var orgs []*github.Organization
	var userErr, orgsErr error

	wg.Add(2)

	// Get user info
	go func() {
		defer wg.Done()
		user, userErr = m.app.githubClient.GetAuthenticatedUser(ctx)
	}()

	// Get organizations
	go func() {
		defer wg.Done()
		orgs, orgsErr = m.app.githubClient.GetUserOrganizations(ctx)
	}()

	wg.Wait()

	if userErr != nil {
		return SearchResultMsg{
			Results: nil,
			Error:   fmt.Errorf("failed to get authenticated user: %w", userErr),
		}
	}

	// Prepare concurrent searches
	type searchResult struct {
		repos []*ghClient.Repository
		err   error
	}

	var searches []func() searchResult

	// Add user search
	if user != nil && user.Login != nil {
		userOpts := ghClient.SearchOptions{
			Query:    query,
			User:     *user.Login,
			Language: language,
			Sort:     strings.ToLower(sortBy),
			Order:    "desc",
			Page:     1,
			Limit:    20, // Reduced from 30 for faster response
		}

		if language == "Any" {
			userOpts.Language = ""
		}

		if !includeForks {
			userOpts.Query += " fork:false"
		}

		searches = append(searches, func() searchResult {
			repos, _, err := m.app.githubClient.SearchRepositories(ctx, userOpts)
			return searchResult{repos: repos, err: err}
		})
	}

	// Add organization searches
	if orgsErr == nil {
		for _, org := range orgs {
			if org.Login != nil {
				// Capture org login for closure
				orgLogin := *org.Login
				orgOpts := ghClient.SearchOptions{
					Query:        query,
					Organization: orgLogin,
					Language:     language,
					Sort:         strings.ToLower(sortBy),
					Order:        "desc",
					Page:         1,
					Limit:        20, // Reduced from 30 for faster response
				}

				if language == "Any" {
					orgOpts.Language = ""
				}

				if !includeForks {
					orgOpts.Query += " fork:false"
				}

				searches = append(searches, func() searchResult {
					repos, _, err := m.app.githubClient.SearchRepositories(ctx, orgOpts)
					return searchResult{repos: repos, err: err}
				})
			}
		}
	}

	// Execute all searches concurrently
	results := make([]searchResult, len(searches))
	wg = sync.WaitGroup{}
	wg.Add(len(searches))

	for i, search := range searches {
		go func(index int, searchFunc func() searchResult) {
			defer wg.Done()
			results[index] = searchFunc()
		}(i, search)
	}

	wg.Wait()

	// Combine all results
	var allRepos []*ghClient.Repository
	for _, result := range results {
		if result.err == nil {
			allRepos = append(allRepos, result.repos...)
		}
	}

	return SearchResultMsg{
		Results: allRepos,
		Error:   nil,
	}
}

// SearchResultMsg contains search results
type SearchResultMsg struct {
	Results []*ghClient.Repository
	Error   error
}
