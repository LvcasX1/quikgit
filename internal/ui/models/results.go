package models

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	ghClient "github.com/lvcasx1/quikgit/internal/github"
)

type ResultItem struct {
	repo     *ghClient.Repository
	selected bool
}

func (r ResultItem) FilterValue() string {
	return r.repo.FullName + " " + r.repo.Description
}

func (r ResultItem) Title() string {
	title := r.repo.FullName
	if r.selected {
		title = "✓ " + title
	}
	return title
}

func (r ResultItem) Description() string {
	var desc strings.Builder

	if r.repo.Description != "" {
		desc.WriteString(r.repo.Description)
	} else {
		desc.WriteString("No description")
	}

	desc.WriteString(fmt.Sprintf(" | ⭐ %d", r.repo.Stars))
	if r.repo.Language != "" {
		desc.WriteString(fmt.Sprintf(" | %s", r.repo.Language))
	}
	desc.WriteString(fmt.Sprintf(" | Updated: %s", r.repo.UpdatedAt.Format("2006-01-02")))

	return desc.String()
}

type ResultsModel struct {
	list     list.Model
	results  []*ghClient.Repository
	selected map[int]bool

	// Styles
	titleStyle  lipgloss.Style
	helpStyle   lipgloss.Style
	statusStyle lipgloss.Style
}

func NewResultsModel() *ResultsModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Search Results"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)

	return &ResultsModel{
		list:     l,
		selected: make(map[int]bool),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")),
	}
}

func (m *ResultsModel) SetResults(results []*ghClient.Repository) {
	m.results = results
	m.selected = make(map[int]bool)

	items := make([]list.Item, len(results))
	for i, repo := range results {
		items[i] = ResultItem{repo: repo, selected: false}
	}

	m.list.SetItems(items)
	m.updateListItems()
}

func (m *ResultsModel) Update(msg tea.Msg) (*ResultsModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for status

	case tea.KeyMsg:
		switch msg.String() {
		case " ":
			// Toggle selection
			if len(m.list.Items()) > 0 {
				index := m.list.Index()
				m.selected[index] = !m.selected[index]
				m.updateListItems()

				// Move to next item
				if index < len(m.list.Items())-1 {
					m.list.CursorDown()
				}
			}

		case "a":
			// Select all
			for i := range m.results {
				m.selected[i] = true
			}
			m.updateListItems()

		case "n":
			// Clear selection
			m.selected = make(map[int]bool)
			m.updateListItems()

		case "enter":
			// Clone selected repositories
			selectedRepos := m.getSelectedRepositories()
			if len(selectedRepos) > 0 {
				return m, func() tea.Msg {
					return AppMsg{Type: "clone_selected", Payload: selectedRepos}
				}
			}

		default:
			m.list, cmd = m.list.Update(msg)
		}
	default:
		m.list, cmd = m.list.Update(msg)
	}

	return m, cmd
}

func (m *ResultsModel) View() string {
	var content strings.Builder

	selectedCount := len(m.getSelectedRepositories())
	status := fmt.Sprintf("Selected: %d/%d repositories", selectedCount, len(m.results))

	content.WriteString(m.statusStyle.Render(status))
	content.WriteString("\n")
	content.WriteString(m.list.View())
	content.WriteString("\n")

	// Help text
	helpText := []string{
		"Space: Toggle selection",
		"a: Select all",
		"n: Clear selection",
		"Enter: Clone selected",
		"Esc: Back",
	}
	content.WriteString(m.helpStyle.Render(strings.Join(helpText, " • ")))

	return content.String()
}

func (m *ResultsModel) updateListItems() {
	items := make([]list.Item, len(m.results))
	for i, repo := range m.results {
		items[i] = ResultItem{
			repo:     repo,
			selected: m.selected[i],
		}
	}
	m.list.SetItems(items)
}

func (m *ResultsModel) getSelectedRepositories() []*github.Repository {
	var selected []*github.Repository
	for i, isSelected := range m.selected {
		if isSelected && i < len(m.results) {
			selected = append(selected, m.results[i])
		}
	}
	return selected
}
