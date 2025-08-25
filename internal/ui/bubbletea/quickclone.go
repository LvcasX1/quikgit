package bubbletea

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuickCloneModel struct {
	app *Application
}

func NewQuickCloneModel(app *Application) *QuickCloneModel {
	return &QuickCloneModel{app: app}
}

func (m *QuickCloneModel) Init() tea.Cmd {
	return nil
}

func (m *QuickCloneModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, m.app.NavigateTo(StateMainMenu)
		}
	}
	return m, nil
}

func (m *QuickCloneModel) View() string {
	// Use full screen dimensions with fallback
	width := m.app.width
	height := m.app.height - 3
	if width == 0 {
		width = 120
	}
	if height <= 0 {
		height = 30
	}

	// Create centered placeholder
	contentStyle := lipgloss.NewStyle().
		Padding(2, 3).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Width(width - 40).
		Align(lipgloss.Center)

	content := contentStyle.Render("󰓅 Quick Clone\n\n(Implementation coming soon)\n\nPress Esc to go back")

	return lipgloss.Place(
		width,
		height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}
