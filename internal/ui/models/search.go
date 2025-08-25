package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type SearchModel struct {
	// Input fields
	searchInput textinput.Model
	userInput   textinput.Model
	orgInput    textinput.Model
	topicInput  textinput.Model

	// State
	focused      int
	githubClient *ghClient.Client
	searching    bool
	error        error

	// Search options
	language string
	sort     string
	order    string

	// Styles
	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
	titleStyle    lipgloss.Style
	errorStyle    lipgloss.Style
	helpStyle     lipgloss.Style
}

type SearchResultMsg struct {
	Results []*ghClient.Repository
	Total   int
	Error   error
}

const (
	searchInputIndex = iota
	userInputIndex
	orgInputIndex
	topicInputIndex
)

func NewSearchModel() *SearchModel {
	// Initialize text inputs
	searchInput := textinput.New()
	searchInput.Placeholder = "Search repositories..."
	searchInput.Focus()
	searchInput.CharLimit = 100
	searchInput.Width = 50

	userInput := textinput.New()
	userInput.Placeholder = "Filter by user (optional)"
	userInput.CharLimit = 50
	userInput.Width = 50

	orgInput := textinput.New()
	orgInput.Placeholder = "Filter by organization (optional)"
	orgInput.CharLimit = 50
	orgInput.Width = 50

	topicInput := textinput.New()
	topicInput.Placeholder = "Filter by topic (optional)"
	topicInput.CharLimit = 50
	topicInput.Width = 50

	return &SearchModel{
		searchInput: searchInput,
		userInput:   userInput,
		orgInput:    orgInput,
		topicInput:  topicInput,
		focused:     searchInputIndex,
		language:    "",
		sort:        "stars",
		order:       "desc",
		activeStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Padding(0, 1),
		inactiveStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center),
		errorStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
	}
}

func (m *SearchModel) SetGitHubClient(client *ghClient.Client) {
	m.githubClient = client
}

func (m *SearchModel) Update(msg tea.Msg) (*SearchModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.focused++
				if m.focused > topicInputIndex {
					m.focused = searchInputIndex
				}
			} else {
				m.focused--
				if m.focused < searchInputIndex {
					m.focused = topicInputIndex
				}
			}
			m.updateFocus()

		case "enter":
			if m.githubClient != nil && !m.searching {
				return m, m.performSearch()
			}

		case "ctrl+l":
			return m, m.showLanguageSelector()

		case "ctrl+s":
			return m, m.showSortSelector()

		default:
			// Update the focused input
			switch m.focused {
			case searchInputIndex:
				m.searchInput, cmd = m.searchInput.Update(msg)
			case userInputIndex:
				m.userInput, cmd = m.userInput.Update(msg)
			case orgInputIndex:
				m.orgInput, cmd = m.orgInput.Update(msg)
			case topicInputIndex:
				m.topicInput, cmd = m.topicInput.Update(msg)
			}
			cmds = append(cmds, cmd)
		}

	case SearchResultMsg:
		m.searching = false
		if msg.Error != nil {
			m.error = msg.Error
		} else {
			// Send results to parent
			return m, func() tea.Msg {
				return AppMsg{Type: "search_completed", Payload: msg.Results}
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *SearchModel) View() string {
	var content strings.Builder

	content.WriteString(m.titleStyle.Render("Repository Search"))
	content.WriteString("\n\n")

	// Search input
	style := m.inactiveStyle
	if m.focused == searchInputIndex {
		style = m.activeStyle
	}
	content.WriteString("Search Query:\n")
	content.WriteString(style.Render(m.searchInput.View()))
	content.WriteString("\n\n")

	// User filter
	style = m.inactiveStyle
	if m.focused == userInputIndex {
		style = m.activeStyle
	}
	content.WriteString("User Filter:\n")
	content.WriteString(style.Render(m.userInput.View()))
	content.WriteString("\n\n")

	// Organization filter
	style = m.inactiveStyle
	if m.focused == orgInputIndex {
		style = m.activeStyle
	}
	content.WriteString("Organization Filter:\n")
	content.WriteString(style.Render(m.orgInput.View()))
	content.WriteString("\n\n")

	// Topic filter
	style = m.inactiveStyle
	if m.focused == topicInputIndex {
		style = m.activeStyle
	}
	content.WriteString("Topic Filter:\n")
	content.WriteString(style.Render(m.topicInput.View()))
	content.WriteString("\n\n")

	// Current filters
	filters := []string{}
	if m.language != "" {
		filters = append(filters, fmt.Sprintf("Language: %s", m.language))
	}
	if m.sort != "" {
		filters = append(filters, fmt.Sprintf("Sort: %s (%s)", m.sort, m.order))
	}

	if len(filters) > 0 {
		content.WriteString("Active Filters: " + strings.Join(filters, ", "))
		content.WriteString("\n\n")
	}

	// Status
	if m.searching {
		content.WriteString("🔍 Searching...")
	} else if m.error != nil {
		content.WriteString(m.errorStyle.Render(fmt.Sprintf("❌ Error: %s", m.error.Error())))
		m.error = nil // Clear error after displaying
	}

	content.WriteString("\n\n")

	// Help text
	helpText := []string{
		"Tab: Next field",
		"Enter: Search",
		"Ctrl+L: Select language",
		"Ctrl+S: Sort options",
		"Esc: Back",
	}
	content.WriteString(m.helpStyle.Render(strings.Join(helpText, " • ")))

	return content.String()
}

func (m *SearchModel) updateFocus() {
	m.searchInput.Blur()
	m.userInput.Blur()
	m.orgInput.Blur()
	m.topicInput.Blur()

	switch m.focused {
	case searchInputIndex:
		m.searchInput.Focus()
	case userInputIndex:
		m.userInput.Focus()
	case orgInputIndex:
		m.orgInput.Focus()
	case topicInputIndex:
		m.topicInput.Focus()
	}
}

func (m *SearchModel) performSearch() tea.Cmd {
	if m.githubClient == nil {
		return nil
	}

	m.searching = true
	m.error = nil

	searchOpts := ghClient.SearchOptions{
		Query:        m.searchInput.Value(),
		User:         m.userInput.Value(),
		Organization: m.orgInput.Value(),
		Topic:        m.topicInput.Value(),
		Language:     m.language,
		Sort:         m.sort,
		Order:        m.order,
		Limit:        30,
		Page:         1,
	}

	return func() tea.Msg {
		ctx := context.Background()
		results, _, err := m.githubClient.SearchRepositories(ctx, searchOpts)

		return SearchResultMsg{
			Results: results,
			Error:   err,
		}
	}
}

func (m *SearchModel) showLanguageSelector() tea.Cmd {
	// TODO: Implement language selector popup
	return nil
}

func (m *SearchModel) showSortSelector() tea.Cmd {
	// TODO: Implement sort selector popup
	return nil
}
