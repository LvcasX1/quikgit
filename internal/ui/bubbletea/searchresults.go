package bubbletea

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type SearchResultsModel struct {
	app           *Application
	cursor        int
	selectedRepos map[int]bool
	viewport      int // For scrolling when we have many results
}

func NewSearchResultsModel(app *Application) *SearchResultsModel {
	return &SearchResultsModel{
		app:           app,
		cursor:        0,
		selectedRepos: make(map[int]bool),
		viewport:      0,
	}
}

func (m *SearchResultsModel) Init() tea.Cmd {
	return nil
}

func (m *SearchResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, m.app.NavigateTo(StateSearch)
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.app.searchResults)-1 {
				m.cursor++
			}
		case " ":
			// Toggle selection
			m.selectedRepos[m.cursor] = !m.selectedRepos[m.cursor]
		case "a":
			// Select all
			for i := range m.app.searchResults {
				m.selectedRepos[i] = true
			}
		case "d":
			// Deselect all
			m.selectedRepos = make(map[int]bool)
		case "enter", "c":
			// Clone selected or current repository
			return m.startCloning()
		case "i":
			// Show repository info (could implement details view)
			// For now, just show in a simple way
		}
	}

	return m, nil
}

func (m *SearchResultsModel) View() string {
	// Use full screen dimensions with fallback
	width := m.app.width
	height := m.app.height - 3
	if width == 0 {
		width = 120
	}
	if height <= 0 {
		height = 30
	}

	if len(m.app.searchResults) == 0 {
		// Center the no results message using full screen layout
		content := m.renderNoResults()
		return lipgloss.Place(
			width,
			height,
			lipgloss.Center,
			lipgloss.Center,
			content,
		)
	}

	var sections []string

	// Title with results count and position
	selectedCount := m.countSelected()
	title := fmt.Sprintf("󰍉 Search Results (%d found", len(m.app.searchResults))
	if selectedCount > 0 {
		title += fmt.Sprintf(", %d selected", selectedCount)
	}
	// Add position indicator when scrollable (more than 5 results)
	if len(m.app.searchResults) > 5 {
		title += fmt.Sprintf(" - %d/%d", m.cursor+1, len(m.app.searchResults))
	}
	title += ")"

	titleStyle := TitleStyle.Copy().Width(width - 20)
	sections = append(sections, titleStyle.Render(title))

	// Repository cards
	cardsSection := m.renderRepositoryCards()
	sections = append(sections, cardsSection)

	// Instructions
	instructions := m.renderInstructions(selectedCount, width)
	sections = append(sections, instructions)

	// Join all sections with center alignment for better presentation
	content := lipgloss.JoinVertical(lipgloss.Center, sections...)

	// Use center alignment both horizontally and vertically for balanced presentation
	// Since we limit to 5 repositories max, we won't have overflow issues
	return lipgloss.Place(
		width,
		height,
		lipgloss.Center, // Center horizontally
		lipgloss.Center, // Center vertically for better balance
		content,
	)
}

func (m *SearchResultsModel) renderNoResults() string {
	noResultsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Bold(true).
		Align(lipgloss.Center).
		MarginTop(5)

	return noResultsStyle.Render("󰋼 No repositories found\nTry adjusting your search criteria")
}

func (m *SearchResultsModel) renderRepositoryCards() string {
	var cards []string

	// Set maximum visible repositories to 5 for better UX
	maxVisible := 5

	start := 0
	end := len(m.app.searchResults)

	if len(m.app.searchResults) > maxVisible {
		// Center the cursor in the visible range
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.app.searchResults) {
			end = len(m.app.searchResults)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		repo := m.app.searchResults[i]
		card := m.renderRepositoryCard(repo, i, i == m.cursor, m.selectedRepos[i])
		cards = append(cards, card)
	}

	// Add scroll indicators if needed
	if start > 0 {
		scrollUp := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("󰁝 %d more results above", start))
		cards = append([]string{scrollUp}, cards...)
	}

	if end < len(m.app.searchResults) {
		remaining := len(m.app.searchResults) - end
		scrollDown := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("󰁅 %d more results below", remaining))
		cards = append(cards, scrollDown)
	}

	return lipgloss.JoinVertical(lipgloss.Center, cards...)
}

func (m *SearchResultsModel) renderRepositoryCard(repo *ghClient.Repository, index int, focused, selected bool) string {
	// Calculate responsive card width
	cardWidth := m.app.width - 20
	if cardWidth < 60 {
		cardWidth = 60
	}

	// Determine card style based on state - NO BACKGROUND COLORS
	var cardStyle lipgloss.Style

	if focused && selected {
		// Focused and selected - most prominent border only
		cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Padding(1, 2).
			MarginBottom(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("205")).
			Bold(true)
	} else if focused {
		// Just focused - bright border only
		cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Padding(1, 2).
			MarginBottom(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("39"))
	} else if selected {
		// Just selected - green border only
		cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Padding(1, 2).
			MarginBottom(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("46"))
	} else {
		// Default state - subtle border
		cardStyle = lipgloss.NewStyle().
			Width(cardWidth).
			Padding(1, 2).
			MarginBottom(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
	}

	// Build card content
	var contentParts []string

	// Header with repo name, owner, and selection indicator
	headerParts := []string{}

	// Selection indicator
	if selected {
		headerParts = append(headerParts, lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("󰄬"))
	} else {
		headerParts = append(headerParts, "  ")
	}

	// Repository name (most prominent)
	repoNameStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	headerParts = append(headerParts, repoNameStyle.Render("󰉋 "+repo.FullName))

	// Language badge
	if repo.Language != "" {
		langStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Background(lipgloss.Color("238")).
			Padding(0, 1).
			Bold(true)
		headerParts = append(headerParts, langStyle.Render(repo.Language))
	}

	header := strings.Join(headerParts, " ")
	contentParts = append(contentParts, header)

	// Description
	if repo.Description != "" {
		desc := repo.Description
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Italic(true).
			MarginTop(1)
		contentParts = append(contentParts, descStyle.Render(desc))
	}

	// Stats line
	var statsParts []string

	// Stars
	if repo.Stars > 0 {
		starStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
		statsParts = append(statsParts, starStyle.Render(fmt.Sprintf("󰓎 %d", repo.Stars)))
	}

	// Forks
	if repo.Forks > 0 {
		forkStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
		statsParts = append(statsParts, forkStyle.Render(fmt.Sprintf("󰓁 %d", repo.Forks)))
	}

	// Updated date
	if !repo.UpdatedAt.IsZero() {
		dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		statsParts = append(statsParts, dateStyle.Render("󰃭 "+repo.UpdatedAt.Format("2006-01-02")))
	}

	// Private indicator
	if repo.Private {
		privateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		statsParts = append(statsParts, privateStyle.Render("󰍁 Private"))
	}

	if len(statsParts) > 0 {
		stats := strings.Join(statsParts, "  ")
		statsStyle := lipgloss.NewStyle().MarginTop(1)
		contentParts = append(contentParts, statsStyle.Render(stats))
	}

	cardContent := lipgloss.JoinVertical(lipgloss.Left, contentParts...)
	return cardStyle.Render(cardContent)
}

func (m *SearchResultsModel) renderInstructions(selectedCount int, width int) string {
	var instructions []string

	instructions = append(instructions, "↑/↓ or j/k: navigate")
	instructions = append(instructions, "Space: select/deselect")
	instructions = append(instructions, "a: select all")
	instructions = append(instructions, "d: deselect all")

	if selectedCount > 0 {
		instructions = append(instructions, fmt.Sprintf("Enter/c: clone %d selected", selectedCount))
	} else {
		instructions = append(instructions, "Enter/c: clone current")
	}

	instructions = append(instructions, "Esc: back to search")

	instructionsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(2).
		Width(width).
		Align(lipgloss.Center)

	return instructionsStyle.Render(strings.Join(instructions, " • "))
}

func (m *SearchResultsModel) countSelected() int {
	count := 0
	for _, selected := range m.selectedRepos {
		if selected {
			count++
		}
	}
	return count
}

func (m *SearchResultsModel) startCloning() (tea.Model, tea.Cmd) {
	var reposToClone []*ghClient.Repository

	selectedCount := m.countSelected()
	if selectedCount > 0 {
		// Clone selected repositories
		for i, selected := range m.selectedRepos {
			if selected && i < len(m.app.searchResults) {
				reposToClone = append(reposToClone, m.app.searchResults[i])
			}
		}
	} else {
		// Clone current repository
		if m.cursor < len(m.app.searchResults) {
			reposToClone = []*ghClient.Repository{m.app.searchResults[m.cursor]}
		}
	}

	if len(reposToClone) == 0 {
		return m, nil
	}

	// Set selected repos in app
	m.app.selectedRepos = reposToClone

	// Navigate to cloning state
	return m, m.app.NavigateTo(StateCloning)
}
